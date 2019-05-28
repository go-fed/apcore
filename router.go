package apcore

import (
	"net/http"

	"github.com/go-fed/activity/pub"
	"github.com/gorilla/mux"
)

type Router struct {
	router *mux.Router
}

func (r *Router) wrap(route *mux.Route) *Route {
	return &Route{
		route: route,
	}
}

// TODO: Actor methods

func (r *Router) ActivityPubOnlyHandleFunc(path string, apHandler pub.HandlerFunc) *Route {
	return r.wrap(r.router.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			isASRequest, err := apHandler(req.Context(), w, req)
			if err != nil {
				ErrorLogger.Error(err)
			}
			if !isASRequest && r.router.NotFoundHandler != nil {
				r.router.NotFoundHandler.ServeHTTP(w, req)
			}
		}))
}

func (r *Router) ActivityPubAndWebHandleFunc(path string, apHandler pub.HandlerFunc, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.wrap(r.router.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			isASRequest, err := apHandler(req.Context(), w, req)
			if err != nil {
				ErrorLogger.Error(err)
			}
			if !isASRequest {
				f(w, req)
			}
		}))
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
	route *mux.Route
}

// TODO: Corresponding Route methods.
