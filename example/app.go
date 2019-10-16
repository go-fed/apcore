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

type App struct {
	config *MyConfig
}

func (a *App) Start() error { return nil }
func (a *App) Stop() error  { return nil }

func (a *App) NewConfiguration() interface{} {
	return &MyConfig{
		FieldS: "blah",
		FieldT: 5,
		FieldU: time.Now(),
	}
}

func (a *App) SetConfiguration(i interface{}) error {
	m, ok := i.(*MyConfig)
	if !ok {
		return fmt.Errorf("SetConfiguration not given a *MyConfig: %T", i)
	}
	a.config = m
	return nil
}

func (a *App) C2SEnabled() bool {
	return true
}

func (a *App) S2SEnabled() bool {
	return true
}

func (a *App) NotFoundHandler() http.Handler {
	// TODO
	return nil
}

func (a *App) MethodNotAllowedHandler() http.Handler {
	// TODO
	return nil
}

func (a *App) InternalServerErrorHandler() http.Handler {
	// TODO
	return nil
}

func (a *App) BadRequestHandler() http.Handler {
	// TODO
	return nil
}

func (a *App) GetInboxHandler() http.Handler {
	// TODO
	return nil
}

func (a *App) GetOutboxHandler() http.Handler {
	// TODO
	return nil
}

func (a *App) BuildRoutes(r *apcore.Router, db apcore.Database, f apcore.Framework) error {
	// TODO
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

func (a *App) ScopePermitsPostOutbox(scope string) (permitted bool, err error) {
	// TODO
	return
}

func (a *App) ScopePermitsPrivateGetInbox(scope string) (permitted bool, err error) {
	// TODO
	return
}

func (a *App) ScopePermitsPrivateGetOutbox(scope string) (permitted bool, err error) {
	// TODO
	return
}

func (a *App) UsernameFromPath(string) string {
	// TODO
	return ""
}

func (a *App) Software() apcore.Software {
	return apcore.Software{
		Name:         "apcore example",
		MajorVersion: 0,
		MinorVersion: 1,
		PatchVersion: 2,
	}
}
