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
	"fmt"
	"net/http"

	"github.com/go-fed/activity/pub"
)

var _ pub.SocialProtocol = &socialBehavior{}

type socialBehavior struct {
	db *database
}

func newSocialBehavior(db *database) *socialBehavior {
	return &socialBehavior{
		db: db,
	}
}

func (s *socialBehavior) AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (s *socialBehavior) Callbacks(c context.Context) (wrapped pub.SocialWrappedCallbacks, other []interface{}, err error) {
	wrapped = pub.SocialWrappedCallbacks{}
	// TODO: Others from Application
	return
}

func (s *socialBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	ctx := ctx{c}
	t, err := ctx.ActivityType()
	if err != nil {
		return err
	}
	return fmt.Errorf("Unhandled client Activity of type: %s", t)
}
