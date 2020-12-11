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
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework"
	"github.com/go-fed/apcore/util"
)

const (
	notFoundTemplate      = "not_found.tmpl"
	notAllowedTemplate    = "not_allowed.tmpl"
	internalErrorTemplate = "internal_error.tmpl"
	badRequestTemplate    = "bad_request.tmpl"
	loginTemplate         = "login.tmpl"
	authTemplate          = "auth.tmpl"
	inboxTemplate         = "inbox.tmpl"
	outboxTemplate        = "outbox.tmpl"
	followersTemplate     = "followers.tmpl"
	followingTemplate     = "following.tmpl"
	usersTemplate         = "users.tmpl"
	listUsersTemplate     = "list_users.tmpl"
	homeTemplate          = "home.tmpl"
)

var _ app.Application = &App{}

// App is an example application that minimally implements the
// app.Application interface.
type App struct {
	// startTime is set when Start is called
	startTime time.Time
	templates *template.Template
}

// newApplication creates a new App for the framework to use.
func newApplication(glob string) (*App, error) {
	t, err := template.ParseGlob(glob)
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
	return nil
}

// SetConfiguration is called with the same type that is returned in
// NewConfiguration, and allows us to save a copy of the values that an
// administrator has configured for our software.
//
// Note we don't do anything with the configuration values in this example
// application. But don't let that stop your imagination from taking off!
func (a *App) SetConfiguration(i interface{}) error {
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
		w.WriteHeader(http.StatusNotFound)
		a.templates.ExecuteTemplate(w, notFoundTemplate, a.getTemplateData(nil))
	})
}

// MethodNotAllowedHandler would scold the user for choosing unorthodox methods
// that resulted in this error, but in this instance of the universe only sends
// a boring reply.
func (a *App) MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		a.templates.ExecuteTemplate(w, notAllowedTemplate, a.getTemplateData(nil))
	})
}

// InternalServerErrorHandler puts the underlying operating system into the time
// out corner. Haha, just kidding, that was a joke. Laugh.
func (a *App) InternalServerErrorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		a.templates.ExecuteTemplate(w, internalErrorTemplate, a.getTemplateData(nil))
	})
}

// BadRequestHandler is "I ran out of witty things to say, so I hope you
// understand the pattern by now" level of error handling. Don't let this
// limited example and snarky commentary demotivate you. If you wanted cold,
// soulless enterprisey software, you came to the wrong place.
func (a *App) BadRequestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		a.templates.ExecuteTemplate(w, badRequestTemplate, a.getTemplateData(nil))
	})
}

// GetLoginWebHandlerFunc returns a handler that renders the login page for
// the user.
//
// The form should POST to "/login", and if the query parameter "login_error"
// is "true" then it should also render the "email or password incorrect" error
// message.
func (a *App) GetLoginWebHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.templates.ExecuteTemplate(w, loginTemplate, a.getTemplateData(nil))
	}
}

// GetAuthWebHandlerFunc returns a handler that renders the authorization page
// for the user to approve in the OAuth2 flow.
func (a *App) GetAuthWebHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.templates.ExecuteTemplate(w, authTemplate, a.getTemplateData(nil))
	}
}

// GetInboxWebHandlerFunc returns a function rendering the outbox. The framework
// passes in a public-only or private view of the outbox, depending on the
// authorization of the incoming request.
func (a *App) GetInboxWebHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, inbox vocab.ActivityStreamsOrderedCollectionPage) {
		a.templates.ExecuteTemplate(w, inboxTemplate, a.getTemplateData(inbox))
	}
}

// GetOutboxWebHandlerFunc returns a function rendering the outbox. The
// framework passes in a public-only or private view of the outbox, depending on
// the authorization of the incoming request.
func (a *App) GetOutboxWebHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
		a.templates.ExecuteTemplate(w, outboxTemplate, a.getTemplateData(outbox))
	}
}

func (a *App) GetFollowersWebHandlerFunc() (http.HandlerFunc, app.AuthorizeFunc) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Pass in followers
			a.templates.ExecuteTemplate(w, followersTemplate, a.getTemplateData(nil))
		}), func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
			return true, nil
		}
}

func (a *App) GetFollowingWebHandlerFunc() (http.HandlerFunc, app.AuthorizeFunc) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Pass in following
			a.templates.ExecuteTemplate(w, followingTemplate, a.getTemplateData(nil))
		}), func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
			return true, nil
		}
}

// GetLikedWebHandlerFunc would have us fetch the user's liked collection and
// then display it in a webpage. Instead, we return null so there's no way to
// view the content as a webpage, but instead it is only obtainable as public
// ActivityStreams data.
func (a *App) GetLikedWebHandlerFunc() (http.HandlerFunc, app.AuthorizeFunc) {
	return nil, nil
}

func (a *App) GetUserWebHandlerFunc() (http.HandlerFunc, app.AuthorizeFunc) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Pass in users
			a.templates.ExecuteTemplate(w, usersTemplate, a.getTemplateData(nil))
		}), func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
			return true, nil
		}
}

// BuildRoutes takes a Router and builds the endpoint http.Handler core.
//
// A database handle and a supplementary Framework object are provided for
// convenience and use in the server's handlers.
func (a *App) BuildRoutes(r app.Router, db app.Database, f app.Framework) error {
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
	//     /login (GET & POST)
	//     /logout (GET)
	//     /authorize (GET & POST)
	//     /token (GET)
	//
	// The framework also handles registering webfinger and host-meta
	// routes:
	//
	//     /.well-known/host-meta
	//     /.well-known/webfinger
	//
	// And supports using Webfinger to find actors on this server.
	//
	// Here we save copeies of our error handlers.
	internalErrorHandler := a.InternalServerErrorHandler()

	// WebOnlyHandleFunc is a convenience function for endpoints with only
	// web content available; no ActivityStreams content exists at this
	// endpoint.
	//
	// It is sugar for Path(...).HandlerFunc(...)
	r.WebOnlyHandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		a.templates.ExecuteTemplate(w, homeTemplate, a.getTemplateData(nil))
	})
	r.WebOnlyHandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Fetch users
		a.templates.ExecuteTemplate(w, listUsersTemplate, a.getTemplateData(nil))
	})
	// ActivityPubHandleFunc is a convenience function for endpoints with
	// only ActivityPub content; no web content exists at this endpoint.
	r.ActivityPubOnlyHandleFunc("/activities/{activity}", func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
		// TODO: Based on activity and any auth, permit or deny
		return false, nil
	})
	// You can use familiar mux methods to route requests appropriately.
	//
	// Finally, add a handler for the new ActivityStream Notes we will
	// be creating.
	r.NewRoute().Path("/notes").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: View list of existing public (and maybe private) notes
		// with pagination.
	})
	// First, we need an authentication function to make sure whoever views
	// the ActivityStreams data has proper credentials to view the web or
	// ActivityStreams data.
	authFn := func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
		vars := framework.Vars(r)
		_ = vars["note"]
		// TODO: Based on the note and any auth, permit or deny
		return false, nil
	}
	r.ActivityPubAndWebHandleFunc("/notes/{note}", authFn, func(w http.ResponseWriter, r *http.Request) {
		// TODO: View note in web page, if authorized.
	})
	// Next, a webpage to handle creating, updating, and deleting notes.
	// This is NOT via C2S, but is done natively in our application.
	r.NewRoute().Path("/notes/create").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure the user is logged in.
		_, authd, err := f.ValidateOAuth2AccessToken(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in GET /notes/create: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// Render the webpage.
		a.templates.ExecuteTemplate(w, "create_note.html", a.getTemplateData(nil))
	})
	r.NewRoute().Path("/notes/create").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure the user is logged in.
		_, authd, err := f.ValidateOAuth2AccessToken(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in POST /notes/create: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// TODO: Determine the user here
		var outboxURI *url.URL
		// TODO: Determine the user's outbox URI
		var create vocab.ActivityStreamsCreate
		// TODO: Build a new Create activity here
		if err := f.Send(r.Context(), outboxURI, create); err != nil {
			util.ErrorLogger.Errorf("error sending when creating note")
			internalErrorHandler.ServeHTTP(w, r)
		}
		// TODO: Redirect to newly created URI
		http.Redirect(w, r, "/notes", http.StatusFound)
	})
	return nil
}

func (a *App) NewID(c context.Context, t vocab.Type) (id *url.URL, err error) {
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
func (a *App) Software() app.Software {
	return app.Software{
		Name:         "BLand",
		MajorVersion: 0,
		MinorVersion: 1,
		PatchVersion: 0,
	}
}

func (a *App) DefaultUserPreferences() interface{} {
	return nil
}

func (a *App) DefaultUserPrivileges() interface{} {
	return nil
}
func (a *App) DefaultAdminPrivileges() interface{} {
	return nil
}

// This is a helper function to generate common data needed in the web
// templates.
func (a *App) getTemplateData(other interface{}) map[string]interface{} {
	return map[string]interface{}{
		"Other": other,
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
