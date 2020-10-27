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

package ap

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/util"
	oa2 "gopkg.in/oauth2.v3"
)

var _ pub.CommonBehavior = &commonBehavior{}

type commonBehavior struct {
	app app.Application
	p   *paths.Paths
	db  *database
	tc  *conn.Controller
	o   *oauth2.Server
}

func newCommonBehavior(
	app app.Application,
	p *paths.Paths,
	db *database,
	tc *conn.Controller,
	o *oauth2.Server) *commonBehavior {
	return &commonBehavior{
		app: app,
		p:   p,
		db:  db,
		tc:  tc,
		o:   o,
	}
}

func (a *commonBehavior) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (newCtx context.Context, authenticated bool, err error) {
	newCtx = c
	var t oa2.TokenInfo
	var oAuthAuthenticated bool
	t, oAuthAuthenticated, err = a.o.ValidateOAuth2AccessToken(w, r)
	if err != nil {
		return
	} else {
		// With or without OAuth, permit public access
		authenticated = true
	}
	// No OAuth2 means guaranteed denial of private access
	if !oAuthAuthenticated {
		return
	}
	// Determine if private access permitted by the granted scope.
	var ok bool
	ok, err = a.app.ScopePermitsPrivateGetInbox(t.GetScope())
	if err != nil {
		return
	} else {
		ctx := &util.Context{c}
		ctx.SetPrivateScope(ok)
		newCtx = ctx.Context
	}
	return
}

func (a *commonBehavior) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (newCtx context.Context, authenticated bool, err error) {
	newCtx = c
	var t oa2.TokenInfo
	var oAuthAuthenticated bool
	t, oAuthAuthenticated, err = a.o.ValidateOAuth2AccessToken(w, r)
	if err != nil {
		return
	} else {
		// With or without OAuth, permit public access
		authenticated = true
	}
	// No OAuth2 means guaranteed denial of private access
	if !oAuthAuthenticated {
		return
	}
	// Determine if private access permitted by the granted scope.
	var ok bool
	ok, err = a.app.ScopePermitsPrivateGetOutbox(t.GetScope())
	if err != nil {
		return
	} else {
		ctx := &util.Context{c}
		ctx.SetPrivateScope(ok)
		newCtx = ctx.Context
	}
	return
}

func (a *commonBehavior) GetOutbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := util.Context{c}
	// IfChange
	var outboxIRI *url.URL
	if outboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	if ctx.HasPrivateScope() {
		ocp, err = a.db.GetOutbox(c, outboxIRI)
	} else {
		ocp, err = a.db.GetPublicOutbox(c, outboxIRI)
	}
	// ThenChange(router.go)
	return
}

func (a *commonBehavior) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	ctx := util.Context{c}
	var userUUID string
	userUUID, err = ctx.UserPathUUID()
	if err != nil {
		return
	}
	var privKey *rsa.PrivateKey
	var kUUID string
	kUUID, privKey, err = a.db.GetUserPKey(c, userUUID)
	if err != nil {
		return
	}
	var pubKeyURL *url.URL
	pubKeyURL, err = a.p.PublicKeyPath(userUUID, kUUID)
	if err != nil {
		return
	}
	pubKeyId := pubKeyURL.String()
	return a.tc.Get(privKey, pubKeyId)
}
