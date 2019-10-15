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
	"fmt"
	"net/http"
	"net/http/httputil"
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

	// TODO: Webfinger
	// TODO: Node-info
	// TODO: Host-meta
	// TODO: Actor routes (public key id)

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

	// Built-in routes for users, default supported:
	// - PostInbox
	// - PostOutbox
	// - GetInbox
	// - GetOutbox
	// - Followers
	// - Following
	// - Liked
	if a.S2SEnabled() {
		r.ActorPostInbox(knownUserPaths[inboxPathKey], "https")
		r.ActorGetInbox(knownUserPaths[inboxPathKey], "https", a.GetInboxHandler().ServeHTTP)
	}
	r.ActorGetOutbox(knownUserPaths[outboxPathKey], "https", a.GetOutboxHandler().ServeHTTP)
	if a.C2SEnabled() {
		r.ActorPostOutbox(knownUserPaths[outboxPathKey], "https")
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
	r.ActivityPubOnlyHandleFunc(knownUserPaths[followersPathKey], scheme, authFn, a.UsernameFromPath)
	r.ActivityPubOnlyHandleFunc(knownUserPaths[followingPathKey], scheme, authFn, a.UsernameFromPath)
	r.ActivityPubOnlyHandleFunc(knownUserPaths[likedPathKey], scheme, authFn, a.UsernameFromPath)

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
