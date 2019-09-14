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

package apcore

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

var _ pub.CommonBehavior = &commonBehavior{}

type commonBehavior struct {
	p  *paths
	db *database
	tc *transportController
}

func newCommonBehavior(
	p *paths,
	db *database,
	tc *transportController) *commonBehavior {
	return &commonBehavior{
		p:  p,
		db: db,
		tc: tc,
	}
}

func (a *commonBehavior) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (a *commonBehavior) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (a *commonBehavior) GetOutbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := ctx{c}
	var outboxIRI *url.URL
	if outboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	ocp, err = a.db.GetOutbox(c, outboxIRI)
	return
}

func (a *commonBehavior) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	ctx := ctx{c}
	var userUUID string
	userUUID, err = ctx.TargetUserUUID()
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
