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
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

var _ pub.FederatingProtocol = &instanceActorFederatingBehavior{}

type instanceActorFederatingBehavior struct {
	maxInboxForwardingDepth int
	maxDeliveryDepth        int
	db                      *Database
	pk                      *services.PrivateKeys
	f                       *services.Followers
	tc                      *conn.Controller
}

func newInstanceActorFederatingBehavior(c *config.Config,
	db *Database,
	pk *services.PrivateKeys,
	f *services.Followers,
	tc *conn.Controller) *instanceActorFederatingBehavior {
	return &instanceActorFederatingBehavior{
		maxInboxForwardingDepth: c.ActivityPubConfig.MaxInboxForwardingRecursionDepth,
		maxDeliveryDepth:        c.ActivityPubConfig.MaxDeliveryRecursionDepth,
		db:                      db,
		pk:                      pk,
		f:                       f,
		tc:                      tc,
	}
}

func (f *instanceActorFederatingBehavior) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (out context.Context, err error) {
	ctx := &util.Context{c}
	ctx.WithActivity(activity)
	out = ctx.Context
	return
}

func (f *instanceActorFederatingBehavior) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	authenticated, err = verifyHttpSignatures(c, r, f.db, f.pk, f.tc)
	out = c
	return
}

func (f *instanceActorFederatingBehavior) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	return
}

func (f *instanceActorFederatingBehavior) FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	wrapped = pub.FederatingWrappedCallbacks{
		OnFollow: pub.OnFollowDoNothing,
	}
	return
}

func (f *instanceActorFederatingBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	activityIRI, err := pub.GetId(activity)
	if err != nil {
		return err
	}
	util.InfoLogger.Infof("Nothing to do for federated Activity of type %q: %s", activity.GetTypeName(), activityIRI)
	return nil
}

func (f *instanceActorFederatingBehavior) MaxInboxForwardingRecursionDepth(c context.Context) int {
	return f.maxInboxForwardingDepth
}

func (f *instanceActorFederatingBehavior) MaxDeliveryRecursionDepth(c context.Context) int {
	return f.maxDeliveryDepth
}

func (f *instanceActorFederatingBehavior) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
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

func (f *instanceActorFederatingBehavior) GetInbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ctx := util.Context{c}
	// IfChange
	var inboxIRI *url.URL
	if inboxIRI, err = ctx.CompleteRequestURL(); err != nil {
		return
	}
	ocp, err = f.db.GetPublicInbox(c, inboxIRI)
	// ThenChange(router.go)
	return
}
