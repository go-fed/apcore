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
	"fmt"
	"net/http"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/util"
	oa2 "gopkg.in/oauth2.v3"
)

var _ pub.SocialProtocol = &SocialBehavior{}

type SocialBehavior struct {
	app app.Application
	o   *oauth2.Server
}

func NewSocialBehavior(app app.Application, o *oauth2.Server) *SocialBehavior {
	return &SocialBehavior{
		app: app,
		o:   o,
	}
}

func (s *SocialBehavior) PostOutboxRequestBodyHook(c context.Context, r *http.Request, data vocab.Type) (out context.Context, err error) {
	ctx := util.Context{c}
	ctx.WithActivityStream(data)
	out = ctx.Context
	return
}

func (s *SocialBehavior) AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	out = c
	var t oa2.TokenInfo
	t, authenticated, err = s.o.ValidateOAuth2AccessToken(w, r)
	if err != nil || !authenticated {
		return
	}
	// Authenticated, but must determine if permitted by the granted scope.
	authenticated, err = s.app.ScopePermitsPostOutbox(t.GetScope())
	return
}

func (s *SocialBehavior) SocialCallbacks(c context.Context) (wrapped pub.SocialWrappedCallbacks, other []interface{}, err error) {
	wrapped = pub.SocialWrappedCallbacks{}
	other = s.app.ApplySocialCallbacks(&wrapped)
	return
}

func (s *SocialBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	return fmt.Errorf("Unhandled client Activity of type: %s", activity.GetTypeName())
}
