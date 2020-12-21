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

package framework

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/nodeinfo"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/framework/web"
	"github.com/go-fed/apcore/framework/webfinger"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
	"github.com/gorilla/mux"
)

const (
	LoginFormEmailKey    = "email"
	LoginFormPasswordKey = "password"
)

func BuildHandler(r *Router,
	internalErrorHandler http.Handler,
	badRequestHandler http.Handler,
	getAuthWebHandler http.Handler,
	getLoginWebHandler http.Handler,
	scheme string,
	c *config.Config,
	a app.Application,
	fr app.Framework,
	actor pub.Actor,
	db RoutingDatabase,
	users *services.Users,
	cy *services.Crypto,
	ni *services.NodeInfo,
	following *services.Following,
	followers *services.Followers,
	liked *services.Liked,
	sqldb *sql.DB,
	oauth *oauth2.Server,
	sl *web.Sessions,
	fw *Framework,
	clock pub.Clock,
	sw, apcore app.Software,
	debug bool) (rt http.Handler, err error) {

	// Static assets
	if sd := c.ServerConfig.StaticRootDirectory; len(sd) == 0 {
		err = fmt.Errorf("static_root_directory is empty")
		return
	} else {
		util.InfoLogger.Infof("Serving static directory: %s", sd)
		fs := http.FileServer(http.Dir(sd))
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	}

	// Dynamic Routes
	// Host-meta
	r.WebOnlyHandleFunc("/.well-known/host-meta",
		hostMetaHandler(scheme, c.ServerConfig.Host))

	// Webfinger
	r.WebOnlyHandleFunc("/.well-known/webfinger",
		webfingerHandler(scheme, c.ServerConfig.Host, badRequestHandler, internalErrorHandler, users))

	// Node-info
	for _, ph := range nodeinfo.GetNodeInfoHandlers(c.NodeInfoConfig, scheme, c.ServerConfig.Host, ni, users, sw, apcore) {
		r.WebOnlyHandleFunc(ph.Path, ph.Handler)
	}

	// Built-in routes for users, default supported:
	// - PostInbox
	// - PostOutbox
	// - GetInbox
	// - GetOutbox
	// - Followers
	// - Following
	// - Liked
	if sa, isS2S := a.(app.S2SApplication); isS2S {
		r.userActorPostInbox()
		r.userActorGetInbox(sa.GetInboxWebHandlerFunc(fr))
	}
	r.userActorGetOutbox(a.GetOutboxWebHandlerFunc(fr))
	if _, isC2S := a.(app.C2SApplication); isC2S {
		r.userActorPostOutbox()
	}

	defaultCollectionSize := c.DatabaseConfig.DefaultCollectionPageSize
	maxCollectionPageSize := c.DatabaseConfig.MaxCollectionPageSize
	addCollectionPageWebFn := func(path string,
		f func(app.Framework) (app.CollectionPageHandlerFunc, app.AuthorizeFunc),
		any services.AnyCPageFn,
		last services.LastCPageFn) {
		web, authFn := f(fr)
		fetch := func(ctx util.Context) (vocab.ActivityStreamsCollectionPage, error) {
			iri, err := ctx.CompleteRequestURL()
			if err != nil {
				return nil, err
			}
			return services.DoCollectionPagination(ctx,
				iri,
				defaultCollectionSize,
				maxCollectionPageSize,
				any,
				last)
		}
		r.apWebCollectionPageFetchingHandleFunc(path, authFn, web, fetch)
	}
	addCollectionPageWebFn(paths.Route(paths.FollowersPathKey),
		a.GetFollowersWebHandlerFunc,
		followers.GetPage,
		followers.GetLastPage)
	addCollectionPageWebFn(paths.Route(paths.FollowingPathKey),
		a.GetFollowingWebHandlerFunc,
		following.GetPage,
		following.GetLastPage)
	addCollectionPageWebFn(paths.Route(paths.LikedPathKey),
		a.GetLikedWebHandlerFunc,
		liked.GetPage,
		liked.GetLastPage)
	addVocabTypeWebFn := func(path string,
		f func(app.Framework) (app.VocabHandlerFunc, app.AuthorizeFunc),
		get func(util.Context, string) (vocab.Type, error)) {
		web, authFn := f(fr)
		fetch := func(ctx util.Context) (vocab.Type, error) {
			id, err := ctx.UserPathUUID()
			if err != nil {
				return nil, err
			}
			return get(ctx, id)
		}
		r.apWebVocabFetchingHandleFunc(path, authFn, web, fetch)
	}
	addVocabTypeWebFn(paths.Route(paths.UserPathKey), a.GetUserWebHandlerFunc, func(ctx util.Context, id string) (vocab.Type, error) {
		u, err := users.UserByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return u.Actor, nil
	})

	// Built-in routes for non-user actors
	for _, k := range paths.AllActors {
		r.knownActor(k)
		r.knownActorPostInbox(k)
		r.knownActorGetInbox(k, nil)
		r.knownActorPostOutbox(k)
		r.knownActorGetOutbox(k, nil)
	}

	// POST/GET login, logout, and OAuth2 routes
	r.NewRoute().
		Path("/login").
		Methods("GET").
		HandlerFunc(
			getLoginFn(oauth, sl, getLoginWebHandler))
	r.NewRoute().
		Path("/login").
		Methods("POST").
		HandlerFunc(
			postLoginFn(oauth, sl, db, badRequestHandler, internalErrorHandler, cy))
	r.NewRoute().
		Path("/logout").
		Methods("GET").
		HandlerFunc(
			getLogoutFn(oauth, sl, internalErrorHandler))
	r.NewRoute().
		Path("/oauth2/authorize").
		Methods("GET").
		HandlerFunc(
			getAuthFn(getAuthWebHandler))
	r.NewRoute().
		Path("/oauth2/authorize").
		Methods("POST").
		HandlerFunc(
			postAuthFn(oauth, sl, db, badRequestHandler, internalErrorHandler, cy))
	r.NewRoute().
		Path("/oauth2/token").
		Methods("GET").
		HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				oauth.HandleAccessTokenRequest(w, r)
			})

	// Application-specific routes
	err = a.BuildRoutes(r, db, fw)
	if err != nil {
		return
	}

	// Middleweare
	r.Use(getFirstPartyCredRefreshFn(oauth, sl))

	if debug {
		util.InfoLogger.Info("Adding request logging middleware for debugging")
		r.Use(requestLogger)
		util.InfoLogger.Info("Adding request timing middleware for debugging")
		r.Use(timingLogger)
		util.InfoLogger.Info("Printing all registered routes for debugging")
		err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			pathTemplate, err := route.GetPathTemplate()
			if err == nil {
				util.InfoLogger.Infof("ROUTE: %s", pathTemplate)
			}
			pathRegexp, err := route.GetPathRegexp()
			if err == nil {
				util.InfoLogger.Infof("Path regexp: %s", pathRegexp)
			}
			queriesTemplates, err := route.GetQueriesTemplates()
			if err == nil {
				util.InfoLogger.Infof("Queries templates: %s", strings.Join(queriesTemplates, ","))
			}
			queriesRegexps, err := route.GetQueriesRegexp()
			if err == nil {
				util.InfoLogger.Infof("Queries regexps: %s", strings.Join(queriesRegexps, ","))
			}
			methods, err := route.GetMethods()
			if err == nil {
				util.InfoLogger.Infof("Methods: %s", strings.Join(methods, ","))
			}
			util.InfoLogger.Infof("")
			return nil
		})
		if err != nil {
			util.ErrorLogger.Errorf("Error walking the registered handlers: %v", err)
		}
	}

	rt = r.router
	return
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprintf("requestLogger debugging middleware failure: %s", err), http.StatusInternalServerError)
			return
		}
		util.InfoLogger.Infof("%s", dump)
		next.ServeHTTP(w, r)
	})
}

func timingLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		util.InfoLogger.Infof("%s took %s", r.URL, end.Sub(start))
	})
}

func getFirstPartyCredRefreshFn(o *oauth2.Server, s *web.Sessions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := util.Context{r.Context()}
			sn, err := s.Get(r)
			if err != nil {
				util.ErrorLogger.Errorf("error getting session in first party cred refresh middleware")
			} else if sn.HasFirstPartyCredentialID() {
				id, errFPC := sn.FirstPartyCredentialID()
				userID, errUUID := sn.UserID()
				if errFPC != nil {
					util.ErrorLogger.Errorf("error refreshing first party cred ind middleware: %v", errFPC)
				}
				if errUUID != nil {
					util.ErrorLogger.Errorf("error refreshing first party cred ind middleware: %v", errUUID)
				}
				if errFPC == nil && errUUID == nil {
					err = o.RefreshProxyCredentialsIfNeeded(ctx, id, userID)
					if err != nil {
						util.ErrorLogger.Errorf("error refreshing first party cred ind middleware: %v", err)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
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
			util.ErrorLogger.Errorf("error writing host-meta response: %s", err)
		} else if n != len(hm) {
			util.ErrorLogger.Errorf("error writing host-meta response: wrote %d of %d bytes", n, len(hm))
		}
	}
}

func webfingerHandler(scheme, host string, badRequestHandler, internalErrorHandler http.Handler, users *services.Users) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		userAccts := strings.Split(
			strings.TrimPrefix(vals.Get("resource"), "acct:"),
			"@")
		if len(userAccts) != 2 {
			util.ErrorLogger.Errorf("error serving webfinger: bad resource: %s", vals.Get("resource"))
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		username := userAccts[0]
		s, err := users.UserByUsername(util.Context{r.Context()}, username)
		if err != nil {
			util.ErrorLogger.Errorf("error serving webfinger: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		uuid := paths.UUID(s.ID)
		wf, err := webfinger.ToWebfinger(scheme, host, username, paths.UUIDPathFor(paths.UserPathKey, uuid))
		if err != nil {
			util.ErrorLogger.Errorf("error serving webfinger: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		b, err := json.Marshal(wf)
		if err != nil {
			util.ErrorLogger.Errorf("error serving webfinger while marshalling: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/jrd+json")
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(b)
		if err != nil {
			util.ErrorLogger.Errorf("error writing webfinger response: %s", err)
		} else if n != len(b) {
			util.ErrorLogger.Errorf("error writing webfinger response: wrote %d of %d bytes", n, len(b))
		}
	}
}

func getLoginFn(oauth *oauth2.Server, sl *web.Sessions, getLoginWebHandler http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sl.Get(r)
		if err != nil {
			util.ErrorLogger.Errorf("error getting session for POST login: %s", err)
		} else {
			_, _, auth, err := oauth.ValidateFirstPartyProxyAccessToken(util.Context{r.Context()}, s)
			if err == nil && auth {
				http.Redirect(w, r, "/", http.StatusFound)
			} else if err != nil {
				// Log but don't fail the request.
				util.ErrorLogger.Errorf("error determining logged-in state in GET login: %s", err)
			}
		}
		getLoginWebHandler.ServeHTTP(w, r)
	}
}

func postLoginFn(oauth *oauth2.Server, sl *web.Sessions, db pub.Database, badRequestHandler, internalErrorHandler http.Handler, cy *services.Crypto) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sl.Get(r)
		if err != nil {
			util.ErrorLogger.Errorf("error getting session for POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		_, _, auth, err := oauth.ValidateFirstPartyProxyAccessToken(util.Context{r.Context()}, s)
		if auth && err == nil {
			// Redirect, already logged-in
			p, err := oauth2.FirstPartyOAuth2LoginRedirPath(r.URL)
			if err != nil {
				util.ErrorLogger.Errorf("error determining first party OAuth2 proxy redirection: %s", err)
				p = "/" // Go to homepage instead of failing request
			}
			http.Redirect(w, r, p, http.StatusFound)
			return
		}
		// Not yet logged in, continue processing
		if r.Form == nil {
			err = r.ParseForm()
			if err != nil {
				badRequestHandler.ServeHTTP(w, r)
				return
			}
		}
		emailV, ok := r.Form[LoginFormEmailKey]
		if !ok || len(emailV) != 1 {
			util.ErrorLogger.Errorf("error validating email or password from form")
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		email := emailV[0]
		passV, ok := r.Form[LoginFormPasswordKey]
		if !ok || len(passV) != 1 {
			util.ErrorLogger.Errorf("error validating email or password from form")
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		pass := passV[0]
		u, valid, err := cy.Valid(util.Context{r.Context()}, email, pass)
		if err != nil {
			util.ErrorLogger.Errorf("error determining password validity in POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		} else if !valid {
			http.Redirect(w, r, oauth2.AddLoginError(r.URL).String(), http.StatusFound)
			return
		}
		s.SetUserID(u)
		// Proxy the first-party login
		id, err := oauth.CreateProxyCredentials(util.Context{r.Context()}, u)
		if err != nil {
			util.ErrorLogger.Errorf("error creating proxy credentials in POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		s.SetFirstPartyCredentialID(id)
		err = s.Save(r, w)
		if err != nil {
			util.ErrorLogger.Errorf("error saving session in POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		p, err := oauth2.FirstPartyOAuth2LoginRedirPath(r.URL)
		if err != nil {
			util.ErrorLogger.Errorf("error determining first party OAuth2 proxy redirection: %s", err)
			p = "/" // Go to homepage instead of failing request
		}
		http.Redirect(w, r, p, http.StatusFound)
	}
}

func getLogoutFn(oauth *oauth2.Server, sl *web.Sessions, internalErrorHandler http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := util.Context{r.Context()}
		sn, err := sl.Get(r)
		if err != nil {
			util.ErrorLogger.Errorf("error getting session for POST login: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if err := oauth.RemoveFirstPartyProxyAccessToken(w, r, ctx, sn); err != nil {
			util.ErrorLogger.Errorf("error removing proxy credential in GET logout: %s", err)
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func getAuthFn(getAuthWebHandler http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		getAuthWebHandler.ServeHTTP(w, r)
	}
}

func postAuthFn(oauth *oauth2.Server, sl *web.Sessions, db pub.Database, badRequestHandler, internalErrorHandler http.Handler, cy *services.Crypto) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := sl.Get(r)
		if err != nil {
			util.ErrorLogger.Errorf("error getting session for POST auth: %s", err)
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
		u, valid, err := cy.Valid(util.Context{r.Context()}, email, pass)
		if err != nil {
			util.ErrorLogger.Errorf("error determining password validity in POST auth: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		} else if !valid {
			http.Redirect(w, r, oauth2.AddAuthError(r.URL).String(), http.StatusFound)
			return
		}
		s.SetUserID(u)
		err = s.Save(r, w)
		if err != nil {
			util.ErrorLogger.Errorf("error saving session in POST auth: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		oauth.HandleAuthorizationRequest(w, r)
	}
}
