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

package framework

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/framework/web"
	oa2 "github.com/go-fed/oauth2"
)

var _ app.Framework = &Framework{}

type Framework struct {
	o                 *oauth2.Server
	s                 *web.Sessions
	actor             pub.Actor
	federationEnabled bool
}

func BuildFramework(fw *Framework, o *oauth2.Server, s *web.Sessions, actor pub.Actor, a app.Application) *Framework {
	_, isS2S := a.(app.S2SApplication)
	fw.o = o
	fw.s = s
	fw.actor = actor
	fw.federationEnabled = isS2S
	return fw
}

func (f *Framework) ValidateOAuth2AccessToken(w http.ResponseWriter, r *http.Request) (token oa2.TokenInfo, authenticated bool, err error) {
	return f.o.ValidateOAuth2AccessToken(w, r)
}

func (f *Framework) Send(c context.Context, outbox *url.URL, t vocab.Type) error {
	if !f.federationEnabled {
		return fmt.Errorf("cannot Send: Framework.Send called when federation is not enabled")
	} else if fa, ok := f.actor.(pub.FederatingActor); !ok {
		return fmt.Errorf("cannot Send: pub.Actor is not a pub.FederatingActor with federation enabled")
	} else {
		_, err := fa.Send(c, outbox, t)
		return err
	}
}

func (f *Framework) Session(r *http.Request) (app.Session, error) {
	return f.s.Get(r)
}
