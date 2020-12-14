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
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

var _ pub.CommonBehavior = &instanceActorCommonBehavior{}

type instanceActorCommonBehavior struct {
	tc *conn.Controller
	db *Database
	pk *services.PrivateKeys
}

func newInstanceActorCommonBehavior(
	db *Database,
	tc *conn.Controller,
	pk *services.PrivateKeys) *instanceActorCommonBehavior {
	return &instanceActorCommonBehavior{
		tc: tc,
		db: db,
		pk: pk,
	}
}

func (a *instanceActorCommonBehavior) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (newCtx context.Context, authenticated bool, err error) {
	authenticated = true
	newCtx = c
	return
}

func (a *instanceActorCommonBehavior) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (newCtx context.Context, authenticated bool, err error) {
	authenticated = true
	newCtx = c
	return
}

func (a *instanceActorCommonBehavior) GetOutbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := util.Context{c}
	// IfChange
	var outboxIRI *url.URL
	if outboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	ocp, err = a.db.GetPublicOutbox(c, outboxIRI)
	// ThenChange(router.go)
	return
}

func (a *instanceActorCommonBehavior) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	var privKey *rsa.PrivateKey
	var pubKeyURL *url.URL
	privKey, pubKeyURL, err = a.pk.GetUserHTTPSignatureKeyForInstanceActor(util.Context{c})
	if err != nil {
		return
	}
	return a.tc.Get(privKey, pubKeyURL.String())
}
