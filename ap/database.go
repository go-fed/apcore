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
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

var _ app.Database = &Database{}

type Database struct {
	inboxes               *services.Inboxes
	outboxes              *services.Outboxes
	users                 *services.Users
	data                  *services.Data
	followers             *services.Followers
	following             *services.Following
	liked                 *services.Liked
	defaultCollectionSize int
	maxCollectionPageSize int
}

func NewDatabase(c *config.Config,
	inboxes *services.Inboxes,
	outboxes *services.Outboxes,
	users *services.Users,
	data *services.Data,
	followers *services.Followers,
	following *services.Following,
	liked *services.Liked) *Database {
	return &Database{
		inboxes:               inboxes,
		outboxes:              outboxes,
		users:                 users,
		data:                  data,
		followers:             followers,
		following:             following,
		liked:                 liked,
		defaultCollectionSize: c.DatabaseConfig.DefaultCollectionPageSize,
		maxCollectionPageSize: c.DatabaseConfig.MaxCollectionPageSize,
	}
}

func (d *Database) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	return d.inboxes.Contains(util.Context{c}, inbox, id)
}

func (d *Database) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.inboxes.GetPage
	last := d.inboxes.GetLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		inboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionPageSize,
		any,
		last)
}

func (d *Database) GetPublicInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.inboxes.GetPublicPage
	last := d.inboxes.GetPublicLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		inboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionPageSize,
		any,
		last)
}

// NOTE: This only prepends the FIRST item in the orderedItems property.
func (d *Database) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	oi := inbox.GetActivityStreamsOrderedItems()
	if oi == nil || oi.Len() == 0 {
		return nil
	}
	id, err := pub.ToId(oi.At(0))
	if err != nil {
		return err
	}
	inboxIRI, err := pub.GetId(inbox)
	if err != nil {
		return err
	}
	return d.inboxes.PrependItem(util.Context{c}, paths.Normalize(inboxIRI), id)
}

func (d *Database) Owns(c context.Context, id *url.URL) (owns bool, err error) {
	owns = d.data.Owns(id)
	return
}

func (d *Database) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return d.users.ActorIDForOutbox(util.Context{c}, paths.Normalize(outboxIRI))
}

func (d *Database) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return d.users.ActorIDForInbox(util.Context{c}, paths.Normalize(inboxIRI))
}

func (d *Database) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	return d.outboxes.OutboxForInbox(util.Context{c}, paths.Normalize(inboxIRI))
}

func (d *Database) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	return d.data.Exists(util.Context{c}, id)
}

func (d *Database) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	return d.data.Get(util.Context{c}, id)
}

func (d *Database) Create(c context.Context, asType vocab.Type) (err error) {
	return d.data.Create(util.Context{c}, asType)
}

func (d *Database) Update(c context.Context, asType vocab.Type) (err error) {
	return d.data.Update(util.Context{c}, asType)
}

func (d *Database) Delete(c context.Context, id *url.URL) (err error) {
	return d.data.Delete(util.Context{c}, id)
}

func (d *Database) GetOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.outboxes.GetPage
	last := d.outboxes.GetLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		outboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionPageSize,
		any,
		last)
}

func (d *Database) GetPublicOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.outboxes.GetPublicPage
	last := d.outboxes.GetPublicLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		outboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionPageSize,
		any,
		last)
}

// NOTE: This only prepends the FIRST item in the orderedItems property.
func (d *Database) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	oi := outbox.GetActivityStreamsOrderedItems()
	if oi == nil || oi.Len() == 0 {
		return nil
	}
	id, err := pub.ToId(oi.At(0))
	if err != nil {
		return err
	}
	outboxIRI, err := pub.GetId(outbox)
	if err != nil {
		return err
	}
	return d.outboxes.PrependItem(util.Context{c}, paths.Normalize(outboxIRI), id)
}

func (d *Database) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return d.followers.GetAllForActor(util.Context{c}, actorIRI)
}

func (d *Database) Following(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return d.following.GetAllForActor(util.Context{c}, actorIRI)
}

func (d *Database) Liked(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return d.liked.GetAllForActor(util.Context{c}, actorIRI)
}
