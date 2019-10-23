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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/gorilla/mux"
)

type handler struct {
	router *Router
}

func newHandler(scheme string, c *config, a Application, actor pub.Actor, db *apdb, oauth *oAuth2Server, clock pub.Clock, debug bool) (h *handler, err error) {
	mr := mux.NewRouter()
	mr.NotFoundHandler = a.NotFoundHandler()
	mr.MethodNotAllowedHandler = a.MethodNotAllowedHandler()

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
		a.InternalServerErrorHandler(),
		a.BadRequestHandler())

	// Host-meta
	r.WebOnlyHandleFunc("/.well-known/host-meta", hostMetaHandler(scheme, c.ServerConfig.Host))

	// Webfinger
	r.WebOnlyHandleFunc("/.well-known/webfinger", webfingerHandler(scheme, c.ServerConfig.Host, a.BadRequestHandler(), a.InternalServerErrorHandler()))

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
		r.actorPostInbox(knownUserPaths[inboxPathKey], "https")
		r.actorGetInbox(knownUserPaths[inboxPathKey], "https", a.GetInboxHandlerFunc())
	}
	r.actorGetOutbox(knownUserPaths[outboxPathKey], "https", a.GetOutboxHandlerFunc())
	if a.C2SEnabled() {
		r.actorPostOutbox(knownUserPaths[outboxPathKey], "https")
	}
	authFn := func(c context.Context, w http.ResponseWriter, r *http.Request) (shouldReturn bool, err error) {
		token, authenticated, err := oauth.ValidateOAuth2AccessToken(w, r)
		if err != nil {
			return
		}
		shouldReturn = !authenticated
		if !authenticated {
			return
		}
		ctx := &ctx{c}
		userId, err := ctx.TargetUserUUID()
		if err != nil {
			return
		}
		shouldReturn = token.GetUserID() != userId
		return
	}
	r.ActivityPubOnlyHandleFunc(knownUserPaths[followersPathKey], scheme, authFn, usernameFromKnownUserPath)
	r.ActivityPubOnlyHandleFunc(knownUserPaths[followingPathKey], scheme, authFn, usernameFromKnownUserPath)
	r.ActivityPubOnlyHandleFunc(knownUserPaths[likedPathKey], scheme, authFn, usernameFromKnownUserPath)

	// Application-specific routes
	err = a.BuildRoutes(r, db, newFramework(scheme, c.ServerConfig.Host, oauth, db))
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
			ErrorLogger.Errorf("error writing host-meta response:", err)
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
			ErrorLogger.Errorf("error writing webfinger response:", err)
		} else if n != len(b) {
			ErrorLogger.Errorf("error writing webfinger response: wrote %d of %d bytes", n, len(b))
		}
	}
}
