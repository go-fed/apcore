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

package app

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router interface {
	ActivityPubOnlyHandleFunc(path string, authFn AuthorizeFunc) Route
	ActivityPubAndWebHandleFunc(path string, authFn AuthorizeFunc, f func(http.ResponseWriter, *http.Request)) Route
	HandleAuthorizationRequest(path string) Route
	HandleAccessTokenRequest(path string) Route
	Get(name string) Route
	WebOnlyHandle(path string, handler http.Handler) Route
	WebOnlyHandleFunc(path string, f func(http.ResponseWriter, *http.Request)) Route
	Handle(path string, handler http.Handler) Route
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) Route
	Headers(pairs ...string) Route
	Host(tpl string) Route
	Methods(methods ...string) Route
	Name(name string) Route
	NewRoute() Route
	Path(tpl string) Route
	PathPrefix(tpl string) Route
	Queries(pairs ...string) Route
	Schemes(schemes ...string) Route
	Use(mwf ...mux.MiddlewareFunc)
	Walk(walkFn mux.WalkFunc) error
}

type Route interface {
	ActivityPubOnlyHandleFunc(path string, authFn AuthorizeFunc) Route
	ActivityPubAndWebHandleFunc(path string, authFn AuthorizeFunc, f func(http.ResponseWriter, *http.Request)) Route
	HandleAuthorizationRequest(path string) Route
	HandleAccessTokenRequest(path string) Route
	WebOnlyHandler(path string, handler http.Handler) Route
	WebOnlyHandlerFunc(path string, f func(http.ResponseWriter, *http.Request)) Route
	Handler(handler http.Handler) Route
	HandlerFunc(f func(http.ResponseWriter, *http.Request)) Route
	Headers(pairs ...string) Route
	Host(tpl string) Route
	Methods(methods ...string) Route
	Name(name string) Route
	Path(tpl string) Route
	PathPrefix(tpl string) Route
	Queries(pairs ...string) Route
	Schemes(schemes ...string) Route
}
