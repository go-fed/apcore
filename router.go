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
	"net/http"

	"github.com/go-fed/activity/pub"
	"github.com/gorilla/mux"
)

type Router struct {
	router            *mux.Router
	db                *apdb
	oauth             *oAuth2Server
	actor             pub.Actor
	clock             pub.Clock
	host              string
	errorHandler      http.Handler
	badRequestHandler http.Handler
}

func newRouter(router *mux.Router,
	db *apdb,
	oauth *oAuth2Server,
	actor pub.Actor,
	clock pub.Clock,
	host string,
	errorHandler http.Handler,
	badRequestHandler http.Handler) *Router {
	return &Router{
		router:            router,
		db:                db,
		oauth:             oauth,
		actor:             actor,
		clock:             clock,
		host:              host,
		errorHandler:      errorHandler,
		badRequestHandler: badRequestHandler,
	}
}

func (r *Router) wrap(route *mux.Route) *Route {
	return &Route{
		route:             route,
		db:                r.db,
		oauth:             r.oauth,
		actor:             r.actor,
		clock:             r.clock,
		host:              r.host,
		errorHandler:      r.errorHandler,
		badRequestHandler: r.badRequestHandler,
		notFoundHandler:   r.router.NotFoundHandler,
	}
}

func (r *Router) ActorPostInbox(path, scheme string) *Route {
	return r.wrap(r.router.NewRoute()).ActorPostInbox(path, scheme)
}

func (r *Router) ActorPostOutbox(path, scheme string) *Route {
	return r.wrap(r.router.NewRoute()).ActorPostOutbox(path, scheme)
}

func (r *Router) ActorGetInbox(path, scheme string, web func(http.ResponseWriter, *http.Request)) *Route {
	return r.wrap(r.router.NewRoute()).ActorGetInbox(path, scheme, web)
}

func (r *Router) ActorGetOutbox(path, scheme string, web func(http.ResponseWriter, *http.Request)) *Route {
	return r.wrap(r.router.NewRoute()).ActorGetOutbox(path, scheme, web)
}

func (r *Router) ActivityPubOnlyHandleFunc(path string, authFn pub.AuthenticateFunc) *Route {
	return r.wrap(r.router.NewRoute()).ActivityPubOnlyHandleFunc(path, authFn)
}

func (r *Router) ActivityPubAndWebHandleFunc(path string, authFn pub.AuthenticateFunc, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.wrap(r.router.NewRoute()).ActivityPubAndWebHandleFunc(path, authFn, f)
}

func (r *Router) HandleAuthorizationRequest(path string) *Route {
	return r.wrap(r.router.NewRoute()).HandleAuthorizationRequest(path)
}

func (r *Router) HandleAccessTokenRequest(path string) *Route {
	return r.wrap(r.router.NewRoute()).HandleAccessTokenRequest(path)
}

func (r *Router) Get(name string) *Route {
	return r.wrap(r.router.Get(name))
}

func (r *Router) WebOnlyHandle(path string, handler http.Handler) *Route {
	return r.wrap(r.router.Handle(path, handler))
}

func (r *Router) WebOnlyHandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.wrap(r.router.HandleFunc(path, f))
}

func (r *Router) Headers(pairs ...string) *Route {
	return r.wrap(r.router.Headers(pairs...))
}

func (r *Router) Host(tpl string) *Route {
	return r.wrap(r.router.Host(tpl))
}

func (r *Router) Methods(methods ...string) *Route {
	return r.wrap(r.router.Methods(methods...))
}

func (r *Router) Name(name string) *Route {
	return r.wrap(r.router.Name(name))
}

func (r *Router) NewRoute() *Route {
	return r.wrap(r.router.NewRoute())
}

func (r *Router) Path(tpl string) *Route {
	return r.wrap(r.router.Path(tpl))
}

func (r *Router) PathPrefix(tpl string) *Route {
	return r.wrap(r.router.PathPrefix(tpl))
}

func (r *Router) Queries(pairs ...string) *Route {
	return r.wrap(r.router.Queries(pairs...))
}

func (r *Router) Schemes(schemes ...string) *Route {
	return r.wrap(r.router.Schemes(schemes...))
}

func (r *Router) Use(mwf ...mux.MiddlewareFunc) {
	r.router.Use(mwf...)
}

func (r *Router) Walk(walkFn mux.WalkFunc) error {
	return r.router.Walk(walkFn)
}

type Route struct {
	route             *mux.Route
	db                *apdb
	oauth             *oAuth2Server
	actor             pub.Actor
	clock             pub.Clock
	host              string
	errorHandler      http.Handler
	badRequestHandler http.Handler
	notFoundHandler   http.Handler
}

func (r *Route) ActorPostInbox(path, scheme string) *Route {
	r.route = r.route.Path(path).Schemes(scheme).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContext(scheme, r.host, req, r.db)
			if err != nil {
				ErrorLogger.Errorf("Error building context for ActorPostInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.PostInbox(c.Context, w, req)
			if err != nil {
				ErrorLogger.Errorf("Error in ActorPostInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			} else if !isApRequest {
				r.badRequestHandler.ServeHTTP(w, req)
				return
			}
			return
		})
	return r
}

func (r *Route) ActorPostOutbox(path, scheme string) *Route {
	r.route = r.route.Path(path).Schemes(scheme).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContext(scheme, r.host, req, r.db)
			if err != nil {
				ErrorLogger.Errorf("Error building context for ActorPostOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.PostOutbox(c.Context, w, req)
			if err != nil {
				ErrorLogger.Errorf("Error in ActorPostOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			} else if !isApRequest {
				r.badRequestHandler.ServeHTTP(w, req)
				return
			}
			return
		})
	return r
}

func (r *Route) ActorGetInbox(path, scheme string, web func(http.ResponseWriter, *http.Request)) *Route {
	r.route = r.route.Path(path).Schemes(scheme).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContext(scheme, r.host, req, r.db)
			if err != nil {
				ErrorLogger.Errorf("Error building context for ActorGetInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.GetInbox(c.Context, w, req)
			if err != nil {
				ErrorLogger.Errorf("Error in ActorGetInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			} else if !isApRequest {
				web(w, req)
				return
			}
			return
		})
	return r
}

func (r *Route) ActorGetOutbox(path, scheme string, web func(http.ResponseWriter, *http.Request)) *Route {
	r.route = r.route.Path(path).Schemes(scheme).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContext(scheme, r.host, req, r.db)
			if err != nil {
				ErrorLogger.Errorf("Error building context for ActorGetOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.GetOutbox(c.Context, w, req)
			if err != nil {
				ErrorLogger.Errorf("Error in ActorGetOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			} else if !isApRequest {
				web(w, req)
				return
			}
			return
		})
	return r
}

func (r *Route) ActivityPubOnlyHandleFunc(path string, authFn pub.AuthenticateFunc) *Route {
	apHandler := pub.NewActivityStreamsHandler(authFn, r.db, r.clock)
	r.route = r.route.Path(path).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			isASRequest, err := apHandler(req.Context(), w, req)
			if err != nil {
				ErrorLogger.Errorf("Error in ActivityPubOnlyHandleFunc: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			if !isASRequest && r.notFoundHandler != nil {
				r.notFoundHandler.ServeHTTP(w, req)
				return
			}
			return
		})
	return r
}

func (r *Route) ActivityPubAndWebHandleFunc(path string, authFn pub.AuthenticateFunc, f func(http.ResponseWriter, *http.Request)) *Route {
	apHandler := pub.NewActivityStreamsHandler(authFn, r.db, r.clock)
	r.route = r.route.Path(path).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			isASRequest, err := apHandler(req.Context(), w, req)
			if err != nil {
				ErrorLogger.Errorf("Error in ActivityPubAndWebHandleFunc: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			if !isASRequest {
				f(w, req)
				return
			}
			return
		})
	return r
}

func (r *Route) HandleAuthorizationRequest(path string) *Route {
	r.route = r.route.Path(path).HandlerFunc(r.oauth.HandleAuthorizationRequest)
	return r
}

func (r *Route) HandleAccessTokenRequest(path string) *Route {
	r.route = r.route.Path(path).HandlerFunc(r.oauth.HandleAccessTokenRequest)
	return r
}

func (r *Route) WebOnlyHandle(path string, handler http.Handler) *Route {
	r.route = r.route.Path(path).Handler(handler)
	return r
}

func (r *Route) WebOnlyHandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route {
	r.route = r.route.Path(path).HandlerFunc(f)
	return r
}

func (r *Route) Headers(pairs ...string) *Route {
	r.route = r.route.Headers(pairs...)
	return r
}

func (r *Route) Host(tpl string) *Route {
	r.route = r.route.Host(tpl)
	return r
}

func (r *Route) Methods(methods ...string) *Route {
	r.route = r.route.Methods(methods...)
	return r
}

func (r *Route) Name(name string) *Route {
	r.route = r.route.Name(name)
	return r
}

func (r *Route) Path(tpl string) *Route {
	r.route = r.route.Path(tpl)
	return r
}

func (r *Route) PathPrefix(tpl string) *Route {
	r.route = r.route.PathPrefix(tpl)
	return r
}

func (r *Route) Queries(pairs ...string) *Route {
	r.route = r.route.Queries(pairs...)
	return r
}

func (r *Route) Schemes(schemes ...string) *Route {
	r.route = r.route.Schemes(schemes...)
	return r
}
