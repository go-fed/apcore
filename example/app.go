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
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/util"
	"github.com/google/uuid"
)

const (
	notFoundTemplate         = "not_found.tmpl"
	notAllowedTemplate       = "not_allowed.tmpl"
	internalErrorTemplate    = "internal_error.tmpl"
	badRequestTemplate       = "bad_request.tmpl"
	loginTemplate            = "login.tmpl"
	authTemplate             = "auth.tmpl"
	inboxTemplate            = "inbox.tmpl"
	outboxTemplate           = "outbox.tmpl"
	followersTemplate        = "followers.tmpl"
	followingTemplate        = "following.tmpl"
	userTemplate             = "user.tmpl"
	listUsersTemplate        = "list_users.tmpl"
	homeTemplate             = "home.tmpl"
	createNoteTemplate       = "create_note.tmpl"
	listNotesTemplate        = "list_notes.tmpl"
	noteTemplate             = "note.tmpl"
	followersRequestTemplate = "followers_request.tmpl"
	followingCreateTemplate  = "following_create.tmpl"
)

var _ app.Application = &App{}
var _ app.S2SApplication = &App{}
var _ app.C2SApplication = &App{}

var fm template.FuncMap = map[string]interface{}{
	"seq": func(n int) []int {
		v := make([]int, n)
		for i := 1; i <= n; i++ {
			v[i-1] = i
		}
		return v
	},
	"isString": func(i interface{}) bool {
		_, ok := i.(string)
		return ok
	},
}

// App is an example application that minimally implements the
// app.Application interface.
type App struct {
	// startTime is set when Start is called
	startTime time.Time
	templates *template.Template
}

// newApplication creates a new App for the framework to use.
func newApplication(glob string) (*App, error) {
	t, err := template.New("").Funcs(fm).ParseGlob(glob)
	if err != nil {
		return nil, err
	}
	util.InfoLogger.Infof("Templates found:")
	for _, tp := range t.Templates() {
		util.InfoLogger.Infof("%s", tp.Name())
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
func (a *App) NotFoundHandler(f app.Framework) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusNotFound, notFoundTemplate, nil, "NotFoundHandler")
	})
}

// MethodNotAllowedHandler would scold the user for choosing unorthodox methods
// that resulted in this error, but in this instance of the universe only sends
// a boring reply.
func (a *App) MethodNotAllowedHandler(f app.Framework) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusMethodNotAllowed, notAllowedTemplate, nil, "MethodNotAllowedHandler")
	})
}

// InternalServerErrorHandler puts the underlying operating system into the time
// out corner. Haha, just kidding, that was a joke. Laugh.
func (a *App) InternalServerErrorHandler(f app.Framework) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		err = a.templates.ExecuteTemplate(w, internalErrorTemplate, a.getTemplateData(s, nil))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving InternalServerErrorHandler: %v", err)
		}
	})
}

// BadRequestHandler is "I ran out of witty things to say, so I hope you
// understand the pattern by now" level of error handling. Don't let this
// limited example and snarky commentary demotivate you. If you wanted cold,
// soulless enterprisey software, you came to the wrong place.
func (a *App) BadRequestHandler(f app.Framework) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusBadRequest, badRequestTemplate, nil, "BadRequestHandler")
	})
}

// GetLoginWebHandlerFunc returns a handler that renders the login page for
// the user.
//
// The form should POST to "/login", and if the query parameter "login_error"
// is "true" then it should also render the "email or password incorrect" error
// message.
func (a *App) GetLoginWebHandlerFunc(f app.Framework) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, loginTemplate, nil, "GetLoginWebHandlerFunc")
	}
}

// GetAuthWebHandlerFunc returns a handler that renders the authorization page
// for the user to approve in the OAuth2 flow.
func (a *App) GetAuthWebHandlerFunc(f app.Framework) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, authTemplate, nil, "GetAuthWebHandlerFunc")
	}
}

// GetInboxWebHandlerFunc returns a function rendering the outbox. The framework
// passes in a public-only or private view of the outbox, depending on the
// authorization of the incoming request.
func (a *App) GetInboxWebHandlerFunc(f app.Framework) func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, inbox vocab.ActivityStreamsOrderedCollectionPage) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, inboxTemplate, inbox, "GetInboxWebHandlerFunc")
	}
}

// GetOutboxWebHandlerFunc returns a function rendering the outbox. The
// framework passes in a public-only or private view of the outbox, depending on
// the authorization of the incoming request.
func (a *App) GetOutboxWebHandlerFunc(f app.Framework) func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
		a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, outboxTemplate, outbox, "GetOutboxWebHandlerFunc")
	}
}

func (a *App) GetFollowersWebHandlerFunc(f app.Framework) (app.CollectionPageHandlerFunc, app.AuthorizeFunc) {
	return func(w http.ResponseWriter, r *http.Request, followers vocab.ActivityStreamsCollectionPage) {
			a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, followersTemplate, followers, "GetFollowersWebHandlerFunc")
		}, func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
			return true, nil
		}
}

func (a *App) GetFollowingWebHandlerFunc(f app.Framework) (app.CollectionPageHandlerFunc, app.AuthorizeFunc) {
	return func(w http.ResponseWriter, r *http.Request, following vocab.ActivityStreamsCollectionPage) {
			a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, followingTemplate, following, "GetFollowingWebHandlerFunc")
		}, func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
			return true, nil
		}
}

// GetLikedWebHandlerFunc would have us fetch the user's liked collection and
// then display it in a webpage. Instead, we return null so there's no way to
// view the content as a webpage, but instead it is only obtainable as public
// ActivityStreams data.
func (a *App) GetLikedWebHandlerFunc(f app.Framework) (app.CollectionPageHandlerFunc, app.AuthorizeFunc) {
	return nil, nil
}

func (a *App) GetUserWebHandlerFunc(f app.Framework) (app.VocabHandlerFunc, app.AuthorizeFunc) {
	return func(w http.ResponseWriter, r *http.Request, user vocab.Type) {
			a.getSessionWriteTemplateHelper(w, r, f, http.StatusOK, userTemplate, user, "GetUserWebHandlerFunc")
		}, func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
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
	internalErrorHandler := a.InternalServerErrorHandler(f)
	badRequestHandler := a.BadRequestHandler(f)

	// WebOnlyHandleFunc is a convenience function for endpoints with only
	// web content available; no ActivityStreams content exists at this
	// endpoint.
	//
	// It is syntactic sugar for Path(...).HandlerFunc(...)
	r.WebOnlyHandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			// TODO: Use internalErrorHandler instead
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		notes, err := getLatestPublicNotes(r.Context(), db)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting latest notes: %v", err)
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		err = a.templates.ExecuteTemplate(w, homeTemplate, a.getTemplateData(s, notes))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving home template: %v", err)
		}
	})
	r.WebOnlyHandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		users, err := getUsers(r.Context(), db)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting latest notes: %v", err)
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		err = a.templates.ExecuteTemplate(w, listUsersTemplate, a.getTemplateData(s, users))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving list users template: %v", err)
		}
	})
	// ActivityPubHandleFunc is a convenience function for endpoints with
	// only ActivityPub content; no web content exists at this endpoint.
	r.ActivityPubOnlyHandleFunc("/activities/{activity}", func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
		return true, nil
	})

	/* Additional web handlers for Note ActivityStreams types */

	// You can use familiar mux methods to route requests appropriately.
	//
	// Add a handler listing the latest notes we are allowed to see:
	// 1) Public notes.
	// 2) Ones a user has published themselves.
	// 3) Ones a user is in the 'to' or 'cc' of.
	r.NewRoute().Path("/notes").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		// View list of existing public (and maybe private) notes with
		// pagination.
		userID, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating token/creds in GET /notes: %s", err)
			// continue processing request as unauthenticated.
		}
		var notes []vocab.Type
		if err != nil || !authd {
			notes, err = getLatestPublicNotes(r.Context(), db)
		} else {
			userIRI := f.UserIRI(userID)
			notes, err = getLatestNotesAndMyPrivateNotes(r.Context(), db, userIRI.String())
		}
		if err != nil {
			util.ErrorLogger.Errorf("error getting notes: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		err = a.templates.ExecuteTemplate(w, listNotesTemplate, a.getTemplateData(s, notes))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving list notes template: %v", err)
		}
	})
	// Next, a webpage to handle creating notes.
	// This is NOT via C2S, but is done natively in our application.
	r.NewRoute().Path("/notes/create").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		// Ensure the user is logged in.
		_, authd, err := f.Validate(w, r)
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
		err = a.templates.ExecuteTemplate(w, createNoteTemplate, a.getTemplateData(s, nil))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving create note template: %v", err)
		}
	})
	r.NewRoute().Path("/notes/create").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure the user is logged in.
		userID, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in POST /notes/create: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Form == nil {
			err = r.ParseForm()
			if err != nil {
				badRequestHandler.ServeHTTP(w, r)
				return
			}
		}
		// Parse the form
		toV, ok := r.Form["note_to"]
		if !ok || len(toV) != 1 {
			util.ErrorLogger.Errorf("error validating to from form")
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		to := toV[0]
		summaryV, ok := r.Form["note_summary"]
		if !ok || len(summaryV) != 1 {
			util.ErrorLogger.Errorf("error validating summary from form")
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		summary := summaryV[0]
		contentV, ok := r.Form["note_content"]
		if !ok || len(contentV) != 1 {
			util.ErrorLogger.Errorf("error validating content from form")
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		content := contentV[0]
		_, public := r.Form["note_public"]

		// Build a new Note
		note := streams.NewActivityStreamsNote()
		toProp := streams.NewActivityStreamsToProperty()
		tos := strings.Split(to, ",")
		for _, t := range tos {
			toIRI, err := url.Parse(t)
			if err != nil {
				util.ErrorLogger.Errorf("error validating an address in to")
				badRequestHandler.ServeHTTP(w, r)
				return
			}
			toProp.AppendIRI(toIRI)
		}
		if public {
			publicIRI, err := url.Parse(pub.PublicActivityPubIRI)
			if err != nil {
				util.ErrorLogger.Errorf("error validating public ActivityStreams address")
				badRequestHandler.ServeHTTP(w, r)
				return
			}
			toProp.AppendIRI(publicIRI)
		}
		note.SetActivityStreamsTo(toProp)
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(summary)
		note.SetActivityStreamsSummary(summaryProp)
		contentProp := streams.NewActivityStreamsContentProperty()
		contentProp.AppendXMLSchemaString(content)
		note.SetActivityStreamsContent(contentProp)
		ctx := util.Context{r.Context()}
		// Send the note -- a Create will automatically be created
		if err := f.Send(ctx, paths.UUID(userID), note); err != nil {
			util.ErrorLogger.Errorf("error sending when creating note: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		iri := note.GetJSONLDId().GetIRI()
		http.Redirect(w, r, iri.String(), http.StatusFound)
	})
	// First, we need an authentication function to make sure whoever views
	// the ActivityStreams data has proper credentials to view the web or
	// ActivityStreams data.
	authFn := func(c util.Context, w http.ResponseWriter, r *http.Request, db app.Database) (permit bool, err error) {
		// Determine who, if any, is logged-in.
		userID, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating token/creds in GET /notes: %s", err)
			// continue processing request as unauthenticated.
		}
		ctx := f.Context(r)
		noteID, err := ctx.CompleteRequestURL()
		if err != nil {
			util.ErrorLogger.Errorf("error sending when creating note: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if err == nil && authd {
			userIRI := f.UserIRI(userID)
			// Authenticated request
			permit, err = getNoteIsReadable(ctx, db, noteID, userIRI)
		} else {
			// Unauthenticated request
			permit, err = getNoteIsPublic(ctx, db, noteID)
		}
		return
	}
	// Next, we use the auth function to protect the note.
	r.ActivityPubAndWebHandleFunc("/notes/{note}", authFn, func(w http.ResponseWriter, r *http.Request) {
		// View note in web page.
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			a.InternalServerErrorHandler(f).ServeHTTP(w, r)
			return
		}
		ctx := f.Context(r)
		noteID, err := ctx.CompleteRequestURL()
		if err != nil {
			util.ErrorLogger.Errorf("error sending when creating note: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		vt, err := f.GetByIRI(ctx, noteID)
		if err != nil {
			util.ErrorLogger.Errorf("error fetching note: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		err = a.templates.ExecuteTemplate(w, noteTemplate, a.getTemplateData(s, vt))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving note template: %v", err)
		}
	})

	/* Followers */

	// These are pages that require a user to be signed-in to manage their
	// followers.
	r.NewRoute().Path("/followers/requests").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in GET /followers/requests: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			// TODO: Better redirection.
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		ctx := f.Context(r)
		frs, err := f.OpenFollowRequests(ctx, userID)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting follow requests: %v", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}

		// Build data for the template
		type tmplData struct {
			ID    string
			Actor string
		}
		var d []tmplData
		for _, v := range frs {
			id, err := pub.GetId(v)
			if err != nil {
				util.ErrorLogger.Errorf("Error getting follow requests: %v", err)
				internalErrorHandler.ServeHTTP(w, r)
				return
			}
			var actor string
			actorsProp := v.GetActivityStreamsActor()
			if actorsProp == nil || actorsProp.Len() == 0 {
				util.ErrorLogger.Errorf("Error getting follow requests: empty actor property")
				internalErrorHandler.ServeHTTP(w, r)
				return
			} else {
				actor = actorsProp.At(0).GetIRI().String()
			}
			t := tmplData{
				ID:    id.String(),
				Actor: actor,
			}
			d = append(d, t)
		}

		err = a.templates.ExecuteTemplate(w, followersRequestTemplate, a.getTemplateData(s, d))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving follower request template: %v", err)
		}
	})
	r.NewRoute().Path("/followers/requests").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in POST /followers/requests: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		err = r.ParseForm()
		if err != nil {
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		ctx := util.Context{r.Context()}
		for k, v := range r.Form {
			if len(v) == 0 {
				continue
			}
			action := v[0]
			if action != "accept" && action != "reject" {
				continue
			}
			followID, err := url.Parse(k)
			if err != nil {
				badRequestHandler.ServeHTTP(w, r)
				return
			}
			if action == "accept" {
				err = f.SendAcceptFollow(ctx, userID, followID)
			} else if action == "reject" {
				err = f.SendRejectFollow(ctx, userID, followID)
			} else {
				// TODO: continue processing instead of failing the bulk processing
				util.ErrorLogger.Errorf("missing accept/reject follow parameter")
				badRequestHandler.ServeHTTP(w, r)
				return
			}
			if err != nil {
				// TODO: continue processing instead of failing the bulk processing
				util.ErrorLogger.Errorf("error when sending %s follow: %s", action, err)
				internalErrorHandler.ServeHTTP(w, r)
				return
			}
		}
		http.Redirect(w, r, "/followers/requests", http.StatusFound)
	})

	/* Following */

	// These are pages that require a user to be signed-in to manage who
	// they wish to follow.
	r.NewRoute().Path("/following/create").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in GET /following/create: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		s, err := f.Session(r)
		if err != nil {
			util.ErrorLogger.Errorf("Error getting session: %v", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		err = a.templates.ExecuteTemplate(w, followingCreateTemplate, a.getTemplateData(s, nil))
		if err != nil {
			util.ErrorLogger.Errorf("Error serving follower request template: %v", err)
		}
	})
	r.NewRoute().Path("/following/create").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, authd, err := f.Validate(w, r)
		if err != nil {
			util.ErrorLogger.Errorf("error validating oauth2 token in POST /following/create: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if !authd {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Parse the form
		err = r.ParseForm()
		if err != nil {
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		toFollowV, ok := r.Form["actor"]
		if !ok || len(toFollowV) != 1 {
			util.ErrorLogger.Errorf("error validating follow form")
			badRequestHandler.ServeHTTP(w, r)
			return
		}
		toFollow, err := url.Parse(toFollowV[0])
		if err != nil {
			util.ErrorLogger.Errorf("error validating follow form: %s", err)
			badRequestHandler.ServeHTTP(w, r)
			return
		}

		// Create follow request
		follow := streams.NewActivityStreamsFollow()

		actorProp := streams.NewActivityStreamsActorProperty()
		actorProp.AppendIRI(f.UserIRI(paths.UUID(userID)))
		follow.SetActivityStreamsActor(actorProp)

		toProp := streams.NewActivityStreamsToProperty()
		toProp.AppendIRI(toFollow)
		follow.SetActivityStreamsTo(toProp)

		objProp := streams.NewActivityStreamsObjectProperty()
		objProp.AppendIRI(toFollow)
		follow.SetActivityStreamsObject(objProp)

		// Send the follow request
		ctx := util.Context{r.Context()}
		if err := f.Send(ctx, paths.UUID(userID), follow); err != nil {
			util.ErrorLogger.Errorf("error sending when sending follow request: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}

		// TODO: Add to following collection

		iri := follow.GetJSONLDId().GetIRI()
		http.Redirect(w, r, iri.String(), http.StatusFound)
	})

	/*We could do something similar for "liked" as we did above for
	"following" or "followers", but we made the choice not to expose the
	"liked" collection at all, so we will just not introduce that concept.*/

	return nil
}

func (a *App) NewIDPath(c context.Context, t vocab.Type) (path string, err error) {
	switch t.GetTypeName() {
	case "Note":
		// This path matches the route created above to serve the data.
		path = fmt.Sprintf("/notes/%s", uuid.New().String())
	case "Create":
		fallthrough
	case "Accept":
		fallthrough
	case "Reject":
		fallthrough
	case "Follow":
		// This path matches the route created above to serve the data.
		path = fmt.Sprintf("/activities/%s", uuid.New().String())
	default:
		err = fmt.Errorf("NewID unhandled type name: %s", t.GetTypeName())
	}
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
	// Here we add new C2S Behavior to cover the Accepting and Rejecting of follow
	// requests.
	//
	// The new behavior is to print out Accept & reject activities to Stdout.
	others = []interface{}{
		func(c context.Context, accept vocab.ActivityStreamsAccept) error {
			fmt.Println(streams.Serialize(accept))
			return nil
		},
		func(c context.Context, reject vocab.ActivityStreamsReject) error {
			fmt.Println(streams.Serialize(reject))
			return nil
		},
	}
	return
}

// ScopePermitsPostOutbox ensures the OAuth2 token scope is "loggedin", which
// is the only permission. Other applications can have more granular
// authorization systems.
func (a *App) ScopePermitsPostOutbox(scope string) (permitted bool, err error) {
	return scope == "postOutbox" || scope == "all", nil
}

// ScopePermitsPrivateGetInbox ensures the OAuth2 token scope is "loggedin",
// which is the only permission. Other applications can have more granular
// authorization systems.
func (a *App) ScopePermitsPrivateGetInbox(scope string) (permitted bool, err error) {
	return scope == "getInbox" || scope == "all", nil
}

// ScopePermitsPrivateGetOutbox ensures the OAuth2 token scope is "loggedin",
// which is the only permission. Other applications can have more granular
// authorization systems.
func (a *App) ScopePermitsPrivateGetOutbox(scope string) (permitted bool, err error) {
	return scope == "getOutbox" || scope == "all", nil
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
func (a *App) getTemplateData(s app.Session, other interface{}) map[string]interface{} {
	if vt, ok := other.(vocab.Type); ok {
		svt, err := streams.Serialize(vt)
		if err == nil {
			other = svt
		} else {
			util.ErrorLogger.Errorf("error serializing ActivityStreams for rendering: %s", err)
		}
	}
	if vts, ok := other.([]vocab.Type); ok {
		var svts []map[string]interface{}
		for _, vt := range vts {
			svt, err := streams.Serialize(vt)
			if err == nil {
				svts = append(svts, svt)
			} else {
				util.ErrorLogger.Errorf("error serializing ActivityStreams for rendering: %s", err)
			}
		}
		other = svts
	}

	m := map[string]interface{}{
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
				Href: "/users",
				Name: "users",
			},
			{
				Href: "/notes",
				Name: "notes",
			},
		},
	}
	if s != nil {
		user, err := s.UserID()
		if err == nil && len(user) > 0 {
			m["User"] = user
		}
	}
	return m
}

func (a *App) getSessionWriteTemplateHelper(w http.ResponseWriter, r *http.Request, f app.Framework, code int, tmpl string, data interface{}, debug string) {
	s, err := f.Session(r)
	if err != nil {
		util.ErrorLogger.Errorf("Error getting session: %v", err)
		a.InternalServerErrorHandler(f).ServeHTTP(w, r)
		return
	}
	w.WriteHeader(code)
	err = a.templates.ExecuteTemplate(w, tmpl, a.getTemplateData(s, data))
	if err != nil {
		util.ErrorLogger.Errorf("Error serving %s: %v", debug, err)
	}
}
