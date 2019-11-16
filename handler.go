// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2019 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package apcore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/gorilla/mux"
)

const (
	LoginFormEmailKey    = "username"
	LoginFormPasswordKey = "password"
)

type handler struct {
	router *Router
}

func newHandler(scheme string, c *config, a Application, actor pub.Actor, db *apdb, oauth *oAuth2Server, sl *sessions, clock pub.Clock, debug bool) (h *handler, err error) {
	mr := mux.NewRouter()
	mr.NotFoundHandler = a.NotFoundHandler()
	mr.MethodNotAllowedHandler = a.MethodNotAllowedHandler()
	internalErrorHandler := a.InternalServerErrorHandler()
	badRequestHandler := a.BadRequestHandler()
	getAuthWebHandler := a.GetAuthWebHandlerFunc()
	getLoginWebHandler := a.GetLoginWebHandlerFunc()

	// Static assets
	if sd := c.ServerConfig.StaticRootDirectory; len(sd) == 0 {
		err = fmt.Errorf("static_root_directory is empty")
		return
	} else {
		InfoLogger.Infof("Serving static directory: %s", sd)
		fs := http.FileServer(http.Dir(sd))
		mr.PathPrefix("/").Handler(fs)
	}

	// Dynamic routes
	r := newRouter(
		mr,
		db,
		oauth,
		actor,
		clock,
		c.ServerConfig.Host,
		scheme,
		internalErrorHandler,
		badRequestHandler)

	// Host-meta
	r.WebOnlyHandleFunc("/.well-known/host-meta", hostMetaHandler(scheme, c.ServerConfig.Host))

	// Webfinger
	r.WebOnlyHandleFunc("/.well-known/webfinger", webfingerHandler(scheme, c.ServerConfig.Host, badRequestHandler, internalErrorHandler))

	// TODO: Node-info
	// TODO: Actor routes (public key id)

	// Built-in routes for users, default supported:
	// - PostInbox
	// - PostOutbox
	// - GetInbox
	// - GetOutbox
	// - Followers
	// - Following
	// - Liked
	if a.S2SEnabled() {
		r.actorPostInbox(knownUserPaths[inboxPathKey], scheme)
		r.actorGetInbox(knownUserPaths[inboxPathKey], scheme, a.GetInboxWebHandlerFunc())
	}
	r.actorGetOutbox(knownUserPaths[outboxPathKey], scheme, a.GetOutboxWebHandlerFunc())
	if a.C2SEnabled() {
		r.actorPostOutbox(knownUserPaths[outboxPathKey], scheme)
	}
	maybeAddWebFn := func(path string, f func() (http.HandlerFunc, AuthorizeFunc)) {
		web, authFn := f()
		if web == nil {
			r.ActivityPubOnlyHandleFunc(path, authFn)
		} else {
			r.ActivityPubAndWebHandleFunc(path, authFn, web)
		}
	}
	maybeAddWebFn(knownUserPaths[followersPathKey], a.GetFollowersWebHandlerFunc)
	maybeAddWebFn(knownUserPaths[followingPathKey], a.GetFollowingWebHandlerFunc)
	maybeAddWebFn(knownUserPaths[likedPathKey], a.GetLikedWebHandlerFunc)
	maybeAddWebFn(knownUserPaths[userPathKey], a.GetUserWebHandlerFunc)

	// POST Login and GET logout routes
	r.NewRoute().Path("/login").Methods("POST").HandlerFunc(postLoginFn(sl, db.database, badRequestHandler, internalErrorHandler))
	r.NewRoute().Path("/login").Methods("GET").HandlerFunc(getLoginWebHandler)
	r.NewRoute().Path("/logout").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, authd, err := oauth.ValidateOAuth2AccessToken(w, r)
		if err != nil {
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if authd {
			if err = oauth.RemoveByAccess(t); err != nil {
				internalErrorHandler.ServeHTTP(w, r)
				return
			}
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	r.NewRoute().Path("/authorize").Methods("GET").HandlerFunc(getAuthFn(sl, internalErrorHandler, getAuthWebHandler))
	r.NewRoute().Path("/authorize").Methods("POST").HandlerFunc(postAuthFn(sl, oauth, internalErrorHandler))
	r.NewRoute().Path("/token").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauth.HandleAccessTokenRequest(w, r)
	})

	// Application-specific routes
	err = a.BuildRoutes(r, db, newFramework(scheme, c.ServerConfig.Host, oauth, db, actor, a.S2SEnabled()))
	if err != nil {
		return
	}

	if debug {
		InfoLogger.Info("Adding request logging middleware for debugging")
		r.Use(requestLogger)
		InfoLogger.Info("Adding request timing middleware for debugging")
		r.Use(timingLogger)
	}

	h = &handler{
		router: r,
	}
	return
}

func (h handler) Handler() http.Handler {
	return h.router.router
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("requestLogger debugging middleware failure: %s", err), http.StatusInternalServerError)
			return
		}
		InfoLogger.Infof("%s", dump)
		next.ServeHTTP(w, r)
	})
}

func timingLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		InfoLogger.Infof("%s took %s", r.URL, end.Sub(start))
	})
}

func hostMetaHandler(scheme, host string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xrd+xml")
		w.WriteHeader(http.StatusOK)
		hm := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<XRD xmlns="http://docs.oasis-open.org/ns/xri/xrd-1.0">
  <Link rel="lrdd" type="application/xrd+xml" template="%s://%s/.well-known/webfinger?resource={uri}"/>
</XRD>`, scheme, host)
		n, err := w.Write([]byte(hm))
		if err != nil {
			ErrorLogger.Errorf("error writing host-meta response: %s", err)
		} else if n != len(hm) {
			ErrorLogger.Errorf("error writing host-meta response: wrote %d of %d bytes", n, len(hm))
		}
	}
}

func webfingerHandler(scheme, host string, badRequestHandler, internalErrorHandler http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		userAccts := strings.Split(
			strings.TrimPrefix(vals.Get("resource"), "acct:"),
			"@")
		if len(userAccts) != 2 {
			ErrorLogger.Errorf("error serving webfinger: bad resource: %s", vals.Get("resource"))
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		username := userAccts[0]
		wf, err := toWebfinger(scheme, host, username, knownUserPathFor(userPathKey, username))
		if err != nil {
			ErrorLogger.Errorf("error serving webfinger: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		b, err := json.Marshal(wf)
		if err != nil {
			ErrorLogger.Errorf("error serving webfinger while marshalling: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/jrd+json")
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(b)
		if err != nil {
			ErrorLogger.Errorf("error writing webfinger response: %s", err)
		} else if n != len(b) {
			ErrorLogger.Errorf("error writing webfinger response: wrote %d of %d bytes", n, len(b))
		}
	}
}

func postLoginFn(sl *sessions, db *database, badRequestHandler, internalErrorHandler http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sl.Get(r)
		if err != nil {
			ErrorLogger.Errorf("error getting session for POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if r.Form == nil {
			err = r.ParseForm()
			if err != nil {
				badRequestHandler.ServeHTTP(w, r)
				return
			}
		}
		emailV, ok := r.Form[LoginFormEmailKey]
		if !ok || len(emailV) != 1 {
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		email := emailV[0]
		passV, ok := r.Form[LoginFormPasswordKey]
		if !ok || len(passV) != 1 {
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		pass := passV[0]
		u, err := db.UserIDFromEmail(r.Context(), email)
		if err != nil {
			ErrorLogger.Errorf("error getting userID for email in POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		valid, err := db.Valid(r.Context(), u, pass)
		if err != nil {
			ErrorLogger.Errorf("error determining password validity in POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		} else if !valid {
			http.Redirect(w, r, "/login?login_error=true", http.StatusFound)
			return
		}
		s.SetUserID(u)
		err = s.Save(r, w)
		if err != nil {
			ErrorLogger.Errorf("error saving session in POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/authorize", http.StatusFound)
	}
}

func getAuthFn(sl *sessions, internalErrorHandler http.Handler, authWebHandler http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sl.Get(r)
		if err != nil {
			ErrorLogger.Errorf("error getting session in GET authorize: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		_, err = s.UserID()
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		authWebHandler(w, r)
	}
}

func postAuthFn(sl *sessions, oa *oAuth2Server, internalErrorHandler http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sl.Get(r)
		if err != nil {
			ErrorLogger.Errorf("error getting session in POST authorize: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if v, ok := s.OAuthRedirectFormValues(); ok {
			r.Form = v
			s.DeleteOAuthRedirectFormValues()
			err = s.Save(r, w)
			if err != nil {
				ErrorLogger.Errorf("error saving session in POST authorize: %s", err)
				internalErrorHandler.ServeHTTP(w, r)
				return
			}
		}
		oa.HandleAuthorizationRequest(w, r)
	}
}
