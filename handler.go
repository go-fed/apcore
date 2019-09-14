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
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/gorilla/mux"
)

type handler struct {
	router *mux.Router
}

func newHandler(c *config, a Application, actor pub.Actor, db *database, debug bool) (h *handler, err error) {
	r := mux.NewRouter()
	r.NotFoundHandler = a.NotFoundHandler()
	r.MethodNotAllowedHandler = a.MethodNotAllowedHandler()

	// Static assets
	if sd := c.ServerConfig.StaticRootDirectory; len(sd) == 0 {
		err = fmt.Errorf("static_root_directory is empty")
		return
	} else {
		InfoLogger.Infof("Serving static directory: %s", sd)
		fs := http.FileServer(http.Dir(sd))
		r.PathPrefix("/").Handler(fs)
	}

	// TODO: Webfinger
	// TODO: Node-info
	// TODO: Host-meta
	// TODO: Actor routes (public key id)

	// Application-specific routes
	err = a.BuildRoutes(&Router{
		router:            r,
		db:                db,
		actor:             actor,
		host:              c.ServerConfig.Host,
		errorHandler:      a.InternalServerErrorHandler(),
		badRequestHandler: a.BadRequestHandler(),
	}, db)
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
	return h.router
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
