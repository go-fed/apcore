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
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

var _ pub.FederatingProtocol = &federatingBehavior{}

type federatingBehavior struct {
	app Application
	db  *database
}

func newFederatingBehavior(a Application, db *database) *federatingBehavior {
	return &federatingBehavior{
		app: a,
		db:  db,
	}
}

func (f *federatingBehavior) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	// 1. Validate HTTP Signatures
	return
}

func (f *federatingBehavior) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	ctx := ctx{c}
	var targetUserId string
	if targetUserId, err = ctx.TargetUserUUID(); err != nil {
		return
	}
	var activityIRI *url.URL
	if activityIRI, err = ctx.ActivityIRI(); err != nil {
		return
	}
	var activityType string
	if activityType, err = ctx.ActivityType(); err != nil {
		return
	}
	// 1. Get Policies For Instance
	var ip policies
	if ip, err = f.db.InstancePolicies(c); err != nil {
		return
	}
	// 2. Get This Actor's Policies
	var ap policies
	if ap, err = f.db.UserPolicies(c, targetUserId); err != nil {
		return
	}
	// 3. Apply policies -- instance first
	p := append(ip, ap...)
	blocked, err = p.IsBlocked(c, f.db, targetUserId, actorIRIs, activityIRI, activityType)
	return
}

func (f *federatingBehavior) Callbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}) {
	// TODO
	return
}

func (f *federatingBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	// TODO
	return nil
}

func (f *federatingBehavior) MaxInboxForwardingRecursionDepth(c context.Context) int {
	return f.app.MaxInboxForwardingRecursionDepth(c)
}

func (f *federatingBehavior) MaxDeliveryRecursionDepth(c context.Context) int {
	return f.app.MaxDeliveryRecursionDepth(c)
}

func (f *federatingBehavior) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	// TODO
	return
}

func (f *federatingBehavior) GetInbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := ctx{c}
	var inboxIRI *url.URL
	if inboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	ocp, err = f.db.GetInbox(c, inboxIRI)
	return
}
