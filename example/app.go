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

package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore"
)

var _ apcore.Application = &App{}

// App is an example application that minimally implements the
// apcore.Application interface.
type App struct {
	// config is populated by SetConfiguration
	config *MyConfig
	// startTime is set when Start is called
	startTime time.Time
	templates *template.Template
}

// newApplication creates a new App for the framework to use.
func newApplication(tmpls []string) (*App, error) {
	t, err := template.ParseFiles(tmpls...)
	if err != nil {
		return nil, err
	}
	return &App{
		templates: t,
	}, nil
}

// Start marks when the uptime began for our application.
func (a *App) Start() error {
	a.startTime = time.Now() // Server timezone.
	return nil
}

// Stop doesn't do anything for the example application.
func (a *App) Stop() error { return nil }

// NewConfiguration returns our custom struct we would like to be populated by
// administrators running our software. These values don't do anything in our
// example application.
func (a *App) NewConfiguration() interface{} {
	return &MyConfig{
		FieldS: "blah",
		FieldT: 5,
		FieldU: time.Now(),
	}
}

// SetConfiguration is called with the same type that is returned in
// NewConfiguration, and allows us to save a copy of the values that an
// administrator has configured for our software.
//
// Note we don't do anything with the configuration values in this example
// application. But don't let that stop your imagination from taking off!
func (a *App) SetConfiguration(i interface{}) error {
	m, ok := i.(*MyConfig)
	if !ok {
		return fmt.Errorf("SetConfiguration not given a *MyConfig: %T", i)
	}
	a.config = m
	return nil
}

// This example server software supports the Social API, C2S. We don't have to.
// But it makes sense to support one of {C2S, S2S}, otherwise what are you
// doing here?
func (a *App) C2SEnabled() bool {
	return true
}

// This example server software supports the Federation API, S2S. We don't have
// to. But it makes sense to support one of {C2S, S2S}, otherwise what are you
// doing here?
func (a *App) S2SEnabled() bool {
	return true
}

// NotFoundHandler returns our spiffy 404 page.
func (a *App) NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	})
}

// MethodNotAllowedHandler would scold the user for choosing unorthodox methods
// that resulted in this error, but in this instance of the universe only sends
// a boring reply.
func (a *App) MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
	})
}

// InternalServerErrorHandler puts the underlying operating system into the time
// out corner. Haha, just kidding, that was a joke. Laugh.
func (a *App) InternalServerErrorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})
}

// BadRequestHandler is "I ran out of witty things to say, so I hope you
// understand the pattern by now" level of error handling. Don't let this
// limited example and snarky commentary demotivate you. If you wanted cold,
// soulless enterprisey software, you came to the wrong place.
func (a *App) BadRequestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	})
}

// GetInboxWebHandlerFunc returns a function rendering the outbox. The framework
// passes in a public-only or private view of the outbox, depending on the
// authorization of the incoming request.
func (a *App) GetInboxWebHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
		// TODO: Write a template and execute it.
	}
}

// GetOutboxWebHandlerFunc returns a function rendering the outbox. The
// framework passes in a public-only or private view of the outbox, depending on
// the authorization of the incoming request.
func (a *App) GetOutboxWebHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
		// TODO: Write a template and execute it.
	}
}

func (a *App) GetFollowersWebHandlerFunc() (http.HandlerFunc, apcore.AuthorizeFunc) {
	// TODO
	return nil, nil
}

func (a *App) GetFollowingWebHandlerFunc() (http.HandlerFunc, apcore.AuthorizeFunc) {
	// TODO
	return nil, nil
}

// GetLikedWebHandlerFunc would have us fetch the user's liked collection and
// then display it in a webpage. Instead, we return null so there's no way to
// view the content as a webpage, but instead it is only obtainable as public
// ActivityStreams data.
func (a *App) GetLikedWebHandlerFunc() (http.HandlerFunc, apcore.AuthorizeFunc) {
	return nil, nil
}

func (a *App) GetUserWebHandlerFunc() (http.HandlerFunc, apcore.AuthorizeFunc) {
	// TODO
	return nil, nil
}

// BuildRoutes takes a Router and builds the endpoint http.Handler core.
//
// A database handle and a supplementary Framework object are provided for
// convenience and use in the server's handlers.
func (a *App) BuildRoutes(r *apcore.Router, db apcore.Database, f apcore.Framework) error {
	// When building routes, the framework already provides actors at the
	// endpoint:
	//
	//     /users/{user}
	//
	// And further routes for the inbox, outbox, followers, following, and
	// liked collections. If you want to use web handlers at these
	// endpoints, other App interface functions allow you to do so.
	//
	// The framework also provides out-of-the-box OAuth2 supporting
	// endpoints:
	//
	//     /login (POST)
	//     /logout (GET)
	//     /authorize (GET & POST)
	//     /token (GET)
	//
	// This means your application still needs to specify a web handler for
	// a GET request to "/login" to display a login page.
	//
	// The framework also handles registering webfinger and host-meta
	// routes:
	//
	//     /.well-known/host-meta
	//     /.well-known/webfinger
	//
	// And supports using Webfinger to find actors on this server.

	// This is a helper function to generate common data needed in the web
	// templates.
	getTemplateData := func() map[string]interface{} {
		return map[string]interface{}{
			"Nav": []struct {
				Href string
				Name string
			}{
				{
					Href: "/",
					Name: "home",
				},
				{
					Href: "/login",
					Name: "login",
				},
				{
					Href: "/logout",
					Name: "logout",
				},
				{
					Href: "/users",
					Name: "users",
				},
			},
		}
	}
	// WebOnlyHandleFunc is a convenience function for endpoints with only
	// web content available; no ActivityStreams content exists at this
	// endpoint.
	//
	// It is sugar for Path(...).HandlerFunc(...)
	r.WebOnlyHandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		a.templates.ExecuteTemplate(w, "home.html", getTemplateData())
	})
	// You can use familiar mux methods to route requests appropriately.
	//
	// This handler displays the login page. The rest of the handlers
	// provide some basic navigational functionality.
	r.NewRoute().Path("/login").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.templates.ExecuteTemplate(w, "login.html", getTemplateData())
	})
	r.WebOnlyHandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		// TODO: List users
		d := getTemplateData()
		a.templates.ExecuteTemplate(w, "users.html", d)
	})

	// Finally, add a handler for the new ActivityStream Notes we will
	// be creating.
	r.NewRoute().Path("/notes").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: View list of existing public (and maybe private) notes
		// with pagination.
	})
	// First, we need an authentication function to make sure whoever views
	// the ActivityStreams data has proper credentials to view the web or
	// ActivityStreams data.
	authFn := func(c apcore.Context, w http.ResponseWriter, r *http.Request, db apcore.Database) (permit bool, err error) {
		// TODO: Based on the note and any auth, permit or deny
		return false, nil
	}
	r.ActivityPubAndWebHandleFunc("/notes/{note}", "https", authFn, func(w http.ResponseWriter, r *http.Request) {
		// TODO: View note in web page, if authorized.
	})
	// Next, a webpage to handle creating, updating, and deleting notes.
	// This is NOT via C2S, but is done natively in our application.
	r.NewRoute().Path("/notes/create").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: View form for creating note.
	})
	r.NewRoute().Path("/notes/create").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Send out new federated info.
	})
	return nil
}

func (a *App) NewId(c context.Context, t vocab.Type) (id *url.URL, err error) {
	// TODO
	return
}

// ApplyFederatingCallbacks lets us provide hooks for our application based on
// incoming ActivityStreams data from peer servers.
func (a *App) ApplyFederatingCallbacks(fwc *pub.FederatingWrappedCallbacks) (others []interface{}) {
	// Here, we add additional behavior to our application if we receive a
	// federated Create activity, besides the spec-suggested side effects.
	//
	// The additional behavior of our application is to print out the
	// Create activity to Stdout.
	fwc.Create = func(c context.Context, create vocab.ActivityStreamsCreate) error {
		fmt.Println(streams.Serialize(create))
		return nil
	}
	// Here we add new behavior to our application.
	//
	// The new behavior is to print out Listen activities to Stdout.
	others = []interface{}{
		func(c context.Context, listen vocab.ActivityStreamsListen) error {
			fmt.Println(streams.Serialize(listen))
			return nil
		},
	}
	return
}

// ApplySocialCallbacks lets us provide hooks for our application based on
// incoming ActivityStreams data from a user's ActivityPub client.
func (a *App) ApplySocialCallbacks(swc *pub.SocialWrappedCallbacks) (others []interface{}) {
	// Here we add no new C2S Behavior. Doing nothing in this function will
	// let the framework handle the suggested C2S side effects.
	return
}

// ScopePermitsPostOutbox ensures the OAuth2 token scope is "loggedin", which
// is the only permission. Other applications can have more granular
// authorization systems.
func (a *App) ScopePermitsPostOutbox(scope string) (permitted bool, err error) {
	return scope == "loggedin", nil
}

// ScopePermitsPrivateGetInbox ensures the OAuth2 token scope is "loggedin",
// which is the only permission. Other applications can have more granular
// authorization systems.
func (a *App) ScopePermitsPrivateGetInbox(scope string) (permitted bool, err error) {
	return scope == "loggedin", nil
}

// ScopePermitsPrivateGetOutbox ensures the OAuth2 token scope is "loggedin",
// which is the only permission. Other applications can have more granular
// authorization systems.
func (a *App) ScopePermitsPrivateGetOutbox(scope string) (permitted bool, err error) {
	return scope == "loggedin", nil
}

// Software describes the current running software, based on the code. This
// allows everyone, from users to developers, to make reasonable judgments about
// the state of the Federative ecosystem as a whole.
//
// Warning: Nothing inherently prevents your application from lying and
// attempting to masquerade as another set of software. Don't be that jerk.
func (a *App) Software() apcore.Software {
	return apcore.Software{
		Name:         "apcore example",
		MajorVersion: 0,
		MinorVersion: 1,
		PatchVersion: 0,
	}
}
