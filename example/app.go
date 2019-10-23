package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore"
)

var _ apcore.Application = &App{}

// App is an example application that minimally implements the
// apcore.Application interface.
type App struct {
	config    *MyConfig
	startTime time.Time
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

// GetInboxHandler returns a function rendering the outbox. The framework
// passes in a public-only or private view of the outbox, depending on the
// authorization of the incoming request.
func (a *App) GetInboxHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
		// TODO: Write a template and execute it.
	}
}

// GetOutboxHandler returns a function rendering the outbox. The framework
// passes in a public-only or private view of the outbox, depending on the
// authorization of the incoming request.
func (a *App) GetOutboxHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
	return func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage) {
		// TODO: Write a template and execute it.
	}
}

func (a *App) BuildRoutes(r *apcore.Router, db apcore.Database, f apcore.Framework) error {
	// TODO: Build routes
	// TODO: Include login page that sets "loggedin" scope
	return nil
}

func (a *App) NewId(c context.Context, t vocab.Type) (id *url.URL, err error) {
	// TODO
	return
}

func (a *App) ApplyFederatingCallbacks(fwc *pub.FederatingWrappedCallbacks) (others []interface{}) {
	// TODO
	return
}

func (a *App) ApplySocialCallbacks(swc *pub.SocialWrappedCallbacks) (others []interface{}) {
	// TODO
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
