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
	maxInboxForwardingDepth int
	maxDeliveryDepth        int
	app                     Application
	p                       *paths
	db                      *database
	tc                      *transportController
}

func newFederatingBehavior(c *config, a Application, p *paths, db *database, tc *transportController) *federatingBehavior {
	return &federatingBehavior{
		maxInboxForwardingDepth: c.ActivityPubConfig.MaxInboxForwardingRecursionDepth,
		maxDeliveryDepth:        c.ActivityPubConfig.MaxDeliveryRecursionDepth,
		app:                     a,
		p:                       p,
		db:                      db,
		tc:                      tc,
	}
}

func (f *federatingBehavior) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (out context.Context, err error) {
	ctx := &ctx{c}
	ctx.withActivityStreamsValue(activity)
	out = ctx.Context
	return
}

func (f *federatingBehavior) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	authenticated, err = verifyHttpSignatures(c, r, f.p, f.db, f.tc)
	out = c
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

func (f *federatingBehavior) Callbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	ctx := ctx{c}
	var u userPreferences
	if u, err = ctx.UserPreferences(); err != nil {
		return
	}
	wrapped = pub.FederatingWrappedCallbacks{
		OnFollow: u.OnFollow(),
	}
	other = f.app.ApplyFederatingCallbacks(&wrapped)
	return
}

func (f *federatingBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	ctx := ctx{c}
	activityIRI, err := ctx.ActivityIRI()
	if err != nil {
		return err
	}
	activityType, err := ctx.ActivityType()
	if err != nil {
		return err
	}
	InfoLogger.Infof("Nothing to do for federated Activity of type %q: %s", activityType, activityIRI)
	return nil
}

func (f *federatingBehavior) MaxInboxForwardingRecursionDepth(c context.Context) int {
	return f.maxInboxForwardingDepth
}

func (f *federatingBehavior) MaxDeliveryRecursionDepth(c context.Context) int {
	return f.maxDeliveryDepth
}

func (f *federatingBehavior) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	ctx := ctx{c}
	var userUUID string
	userUUID, err = ctx.TargetUserUUID()
	if err != nil {
		return
	}
	// Here we limit to only allow forwarding to the target user's
	// followers.
	var fc vocab.ActivityStreamsCollection
	fc, err = f.db.FollowersByUserUUID(c, userUUID)
	if err != nil {
		return
	}
	allowedRecipients := make(map[*url.URL]bool, 0)
	items := fc.GetActivityStreamsItems()
	if items != nil {
		for iter := items.Begin(); iter != items.End(); iter = iter.Next() {
			var id *url.URL
			id, err = pub.ToId(iter)
			if err != nil {
				return
			}
			allowedRecipients[id] = true
		}
	}
	for _, elem := range potentialRecipients {
		if has, ok := allowedRecipients[elem]; ok && has {
			filteredRecipients = append(filteredRecipients, elem)
		}
	}
	return
}

func (f *federatingBehavior) GetInbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := ctx{c}
	var inboxIRI *url.URL
	if inboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	if ctx.HasPrivateScope() {
		ocp, err = f.db.GetInbox(c, inboxIRI)
	} else {
		ocp, err = f.db.GetPublicInbox(c, inboxIRI)
	}
	return
}
