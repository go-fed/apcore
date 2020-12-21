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
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

var _ pub.FederatingProtocol = &FederatingBehavior{}

type FederatingBehavior struct {
	maxInboxForwardingDepth int
	maxDeliveryDepth        int
	app                     app.S2SApplication
	db                      *Database
	po                      *services.Policies
	pk                      *services.PrivateKeys
	f                       *services.Followers
	u                       *services.Users
	tc                      *conn.Controller
}

func NewFederatingBehavior(c *config.Config,
	a app.S2SApplication,
	db *Database,
	po *services.Policies,
	pk *services.PrivateKeys,
	f *services.Followers,
	u *services.Users,
	tc *conn.Controller) *FederatingBehavior {
	return &FederatingBehavior{
		maxInboxForwardingDepth: c.ActivityPubConfig.MaxInboxForwardingRecursionDepth,
		maxDeliveryDepth:        c.ActivityPubConfig.MaxDeliveryRecursionDepth,
		app:                     a,
		db:                      db,
		po:                      po,
		pk:                      pk,
		f:                       f,
		u:                       u,
		tc:                      tc,
	}
}

func (f *FederatingBehavior) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (out context.Context, err error) {
	ctx := &util.Context{c}
	ctx.WithActivity(activity)
	out = ctx.Context
	return
}

func (f *FederatingBehavior) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	authenticated, err = verifyHttpSignatures(c, r, f.db, f.pk, f.tc)
	out = c
	return
}

func (f *FederatingBehavior) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	ctx := util.Context{c}
	var activity pub.Activity
	if activity, err = ctx.Activity(); err != nil {
		return
	}
	var actorID *url.URL
	if actorID, err = ctx.ActorIRI(); err != nil {
		return
	}
	blocked, err = f.po.IsBlocked(ctx, actorID, activity)
	return
}

func (f *FederatingBehavior) FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	ctx := util.Context{c}
	var uuid paths.UUID
	uuid, err = ctx.UserPathUUID()
	if err != nil {
		return
	}
	var prefs *services.Preferences
	prefs, err = f.u.Preferences(ctx, uuid, nil)
	if err != nil {
		return
	}
	wrapped = pub.FederatingWrappedCallbacks{
		OnFollow: prefs.OnFollow,
	}
	other = f.app.ApplyFederatingCallbacks(&wrapped)
	return
}

func (f *FederatingBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	activityIRI, err := pub.GetId(activity)
	if err != nil {
		return err
	}
	util.InfoLogger.Infof("Nothing to do for federated Activity of type %q: %s", activity.GetTypeName(), activityIRI)
	return nil
}

func (f *FederatingBehavior) MaxInboxForwardingRecursionDepth(c context.Context) int {
	return f.maxInboxForwardingDepth
}

func (f *FederatingBehavior) MaxDeliveryRecursionDepth(c context.Context) int {
	return f.maxDeliveryDepth
}

func (f *FederatingBehavior) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	ctx := util.Context{c}
	var actorIRI *url.URL
	actorIRI, err = ctx.ActorIRI()
	if err != nil {
		return
	}
	// Here we limit to only allow forwarding to the target user's
	// followers.
	var fc vocab.ActivityStreamsCollection
	fc, err = f.f.GetAllForActor(ctx, actorIRI)
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

func (f *FederatingBehavior) GetInbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := util.Context{c}
	// IfChange
	var inboxIRI *url.URL
	if inboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	if ctx.HasPrivateScope() {
		ocp, err = f.db.GetInbox(c, inboxIRI)
	} else {
		ocp, err = f.db.GetPublicInbox(c, inboxIRI)
	}
	// ThenChange(router.go)
	return
}
