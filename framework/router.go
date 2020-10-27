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
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
	"github.com/gorilla/mux"
)

var _ app.Router = &Router{}

type Router struct {
	router            *mux.Router
	inboxes           services.Inboxes
	outboxes          services.Outboxes
	oauth             *oauth2.Server
	actor             pub.Actor
	clock             pub.Clock
	host              string
	scheme            string
	errorHandler      http.Handler
	badRequestHandler http.Handler
}

func newRouter(router *mux.Router,
	inboxes services.Inboxes,
	outboxes services.Outboxes,
	oauth *oauth2.Server,
	actor pub.Actor,
	clock pub.Clock,
	host string,
	scheme string,
	errorHandler http.Handler,
	badRequestHandler http.Handler) *Router {
	return &Router{
		router:            router,
		inboxes:           inboxes,
		outboxes:          outboxes,
		oauth:             oauth,
		actor:             actor,
		clock:             clock,
		host:              host,
		scheme:            scheme,
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
		scheme:            r.scheme,
		errorHandler:      r.errorHandler,
		badRequestHandler: r.badRequestHandler,
		notFoundHandler:   r.router.NotFoundHandler,
	}
}

func (r *Router) actorPostInbox(path, scheme string) *Route {
	return r.wrap(r.router.NewRoute()).actorPostInbox(path, scheme)
}

func (r *Router) actorPostOutbox(path, scheme string) *Route {
	return r.wrap(r.router.NewRoute()).actorPostOutbox(path, scheme)
}

func (r *Router) actorGetInbox(path, scheme string, web func(http.ResponseWriter, *http.Request, vocab.ActivityStreamsOrderedCollectionPage)) *Route {
	return r.wrap(r.router.NewRoute()).actorGetInbox(path, scheme, web)
}

func (r *Router) actorGetOutbox(path, scheme string, web func(http.ResponseWriter, *http.Request, vocab.ActivityStreamsOrderedCollectionPage)) *Route {
	return r.wrap(r.router.NewRoute()).actorGetOutbox(path, scheme, web)
}

func (r *Router) ActivityPubOnlyHandleFunc(path string, authFn app.AuthorizeFunc) app.Route {
	return r.wrap(r.router.NewRoute()).ActivityPubOnlyHandleFunc(path, authFn)
}

func (r *Router) ActivityPubAndWebHandleFunc(path string, authFn app.AuthorizeFunc, f func(http.ResponseWriter, *http.Request)) app.Route {
	return r.wrap(r.router.NewRoute()).ActivityPubAndWebHandleFunc(path, authFn, f)
}

func (r *Router) HandleAuthorizationRequest(path string) app.Route {
	return r.wrap(r.router.NewRoute()).HandleAuthorizationRequest(path)
}

func (r *Router) HandleAccessTokenRequest(path string) app.Route {
	return r.wrap(r.router.NewRoute()).HandleAccessTokenRequest(path)
}

func (r *Router) Get(name string) app.Route {
	return r.wrap(r.router.Get(name))
}

func (r *Router) WebOnlyHandle(path string, handler http.Handler) app.Route {
	return r.wrap(r.router.Handle(path, handler))
}

func (r *Router) WebOnlyHandleFunc(path string, f func(http.ResponseWriter, *http.Request)) app.Route {
	return r.wrap(r.router.HandleFunc(path, f))
}

func (r *Router) Handle(path string, handler http.Handler) app.Route {
	return r.wrap(r.router.Handle(path, handler))
}

func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) app.Route {
	return r.wrap(r.router.HandleFunc(path, f))
}

func (r *Router) Headers(pairs ...string) app.Route {
	return r.wrap(r.router.Headers(pairs...))
}

func (r *Router) Host(tpl string) app.Route {
	return r.wrap(r.router.Host(tpl))
}

func (r *Router) Methods(methods ...string) app.Route {
	return r.wrap(r.router.Methods(methods...))
}

func (r *Router) Name(name string) app.Route {
	return r.wrap(r.router.Name(name))
}

func (r *Router) NewRoute() app.Route {
	return r.wrap(r.router.NewRoute())
}

func (r *Router) Path(tpl string) app.Route {
	return r.wrap(r.router.Path(tpl))
}

func (r *Router) PathPrefix(tpl string) app.Route {
	return r.wrap(r.router.PathPrefix(tpl))
}

func (r *Router) Queries(pairs ...string) app.Route {
	return r.wrap(r.router.Queries(pairs...))
}

func (r *Router) Schemes(schemes ...string) app.Route {
	return r.wrap(r.router.Schemes(schemes...))
}

func (r *Router) Use(mwf ...mux.MiddlewareFunc) {
	r.router.Use(mwf...)
}

func (r *Router) Walk(walkFn mux.WalkFunc) error {
	return r.router.Walk(walkFn)
}

var _ app.Route = &Route{}

type Route struct {
	route             *mux.Route
	db                *apdb
	oauth             *oauth2.Server
	actor             pub.Actor
	clock             pub.Clock
	host              string
	scheme            string
	errorHandler      http.Handler
	badRequestHandler http.Handler
	notFoundHandler   http.Handler
}

func (r *Route) actorPostInbox(path, scheme string) *Route {
	r.route = r.route.Path(path).Schemes(scheme).Methods("POST").HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContextForBox(scheme, r.host, w, req, r.db, r.oauth)
			if err != nil {
				util.ErrorLogger.Errorf("Error building context for ActorPostInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.PostInbox(c.Context, w, req)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActorPostInbox: %s", err)
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

func (r *Route) actorPostOutbox(path, scheme string) *Route {
	r.route = r.route.Path(path).Schemes(scheme).Methods("POST").HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContextForBox(scheme, r.host, w, req, r.db, r.oauth)
			if err != nil {
				util.ErrorLogger.Errorf("Error building context for ActorPostOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.PostOutbox(c.Context, w, req)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActorPostOutbox: %s", err)
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

func (r *Route) actorGetInbox(path, scheme string, web func(w http.ResponseWriter, r *http.Request, inbox vocab.ActivityStreamsOrderedCollectionPage)) *Route {
	r.route = r.route.Path(path).Schemes(scheme).Methods("GET").HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContextForBox(scheme, r.host, w, req, r.db, r.oauth)
			if err != nil {
				util.ErrorLogger.Errorf("Error building context for ActorGetInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.GetInbox(c.Context, w, req)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActorGetInbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			} else if !isApRequest {
				// IfChange
				var inboxIRI *url.URL
				if inboxIRI, err = c.CompleteRequestURL(); err != nil {
					return
				}
				var inbox vocab.ActivityStreamsOrderedCollectionPage
				if c.HasPrivateScope() {
					inbox, err = r.inboxes.GetInbox(c, inboxIRI)
				} else {
					inbox, err = r.inboxes.GetPublicInbox(c, inboxIRI)
				}
				// ThenChange(ap_s2s.go)
				if web != nil {
					web(w, req, inbox)
				}
				return
			}
			return
		})
	return r
}

func (r *Route) actorGetOutbox(path, scheme string, web func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage)) *Route {
	r.route = r.route.Path(path).Schemes(scheme).Methods("GET").HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContextForBox(scheme, r.host, w, req, r.db, r.oauth)
			if err != nil {
				util.ErrorLogger.Errorf("Error building context for ActorGetOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			isApRequest, err := r.actor.GetOutbox(c.Context, w, req)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActorGetOutbox: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			} else if !isApRequest {
				// IfChange
				var outboxIRI *url.URL
				if outboxIRI, err = c.CompleteRequestURL(); err != nil {
					return
				}
				var outbox vocab.ActivityStreamsOrderedCollectionPage
				if c.HasPrivateScope() {
					outbox, err = r.outboxes.GetOutbox(c, outboxIRI)
				} else {
					outbox, err = r.outboxes.GetPublicOutbox(c, outboxIRI)
				}
				// ThenChange(ap_common.go)
				if web != nil {
					web(w, req, outbox)
				}
				return
			}
			return
		})
	return r
}

func (r *Route) ActivityPubOnlyHandleFunc(path string, authFn app.AuthorizeFunc) app.Route {
	apHandler := pub.NewActivityStreamsHandler(r.db, r.clock)
	r.route = r.route.Path(path).Schemes(r.scheme).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContext(r.scheme, r.host, w, req, r.db, r.oauth)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActivityPubOnlyHandleFunc newRequestContext: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			permit := true
			if authFn != nil {
				permit, err = authFn(c, w, req, r.db)
				if err != nil {
					util.ErrorLogger.Errorf("Error in ActivityPubOnlyHandleFunc authFn: %s", err)
					r.errorHandler.ServeHTTP(w, req)
					return
				}
			}
			if !permit {
				r.notFoundHandler.ServeHTTP(w, req)
				return
			}
			isASRequest, err := apHandler(c, w, req)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActivityPubOnlyHandleFunc: %s", err)
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

func (r *Route) ActivityPubAndWebHandleFunc(path string, authFn app.AuthorizeFunc, f func(http.ResponseWriter, *http.Request)) app.Route {
	apHandler := pub.NewActivityStreamsHandler(r.db, r.clock)
	r.route = r.route.Path(path).Schemes(r.scheme).HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			c, err := newRequestContext(r.scheme, r.host, w, req, r.db, r.oauth)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActivityPubAndWebHandleFunc newRequestContext: %s", err)
				r.errorHandler.ServeHTTP(w, req)
				return
			}
			permit := true
			if authFn != nil {
				permit, err = authFn(c, w, req, r.db)
				if err != nil {
					util.ErrorLogger.Errorf("Error in ActivityPubAndWebHandleFunc authFn: %s", err)
					r.errorHandler.ServeHTTP(w, req)
					return
				}
			}
			if !permit {
				r.notFoundHandler.ServeHTTP(w, req)
				return
			}
			isASRequest, err := apHandler(c, w, req)
			if err != nil {
				util.ErrorLogger.Errorf("Error in ActivityPubAndWebHandleFunc: %s", err)
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

func (r *Route) HandleAuthorizationRequest(path string) app.Route {
	r.route = r.route.Path(path).HandlerFunc(r.oauth.HandleAuthorizationRequest)
	return r
}

func (r *Route) HandleAccessTokenRequest(path string) app.Route {
	r.route = r.route.Path(path).HandlerFunc(r.oauth.HandleAccessTokenRequest)
	return r
}

func (r *Route) WebOnlyHandler(path string, handler http.Handler) app.Route {
	r.route = r.route.Path(path).Handler(handler)
	return r
}

func (r *Route) WebOnlyHandlerFunc(path string, f func(http.ResponseWriter, *http.Request)) app.Route {
	r.route = r.route.Path(path).HandlerFunc(f)
	return r
}

func (r *Route) Handler(handler http.Handler) app.Route {
	r.route = r.route.Handler(handler)
	return r
}

func (r *Route) HandlerFunc(f func(http.ResponseWriter, *http.Request)) app.Route {
	r.route = r.route.HandlerFunc(f)
	return r
}

func (r *Route) Headers(pairs ...string) app.Route {
	r.route = r.route.Headers(pairs...)
	return r
}

func (r *Route) Host(tpl string) app.Route {
	r.route = r.route.Host(tpl)
	return r
}

func (r *Route) Methods(methods ...string) app.Route {
	r.route = r.route.Methods(methods...)
	return r
}

func (r *Route) Name(name string) app.Route {
	r.route = r.route.Name(name)
	return r
}

func (r *Route) Path(tpl string) app.Route {
	r.route = r.route.Path(tpl)
	return r
}

func (r *Route) PathPrefix(tpl string) app.Route {
	r.route = r.route.PathPrefix(tpl)
	return r
}

func (r *Route) Queries(pairs ...string) app.Route {
	r.route = r.route.Queries(pairs...)
	return r
}

func (r *Route) Schemes(schemes ...string) app.Route {
	r.route = r.route.Schemes(schemes...)
	return r
}
