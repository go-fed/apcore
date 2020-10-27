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
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/services"
)

var _ app.Database = &database{}

type database struct {
	db                    *sql.DB
	app                   app.Application
	inboxes               *services.Inboxes
	outboxes              *services.Outboxes
	users                 *services.Users
	data                  *services.Data
	followers             *services.Followers
	following             *services.Following
	liked                 *services.Liked
	hostname              string
	defaultCollectionSize int
	maxCollectionPageSize int
}

func newDatabase(c *config.Config, a app.Application, db *sql.DB, debug bool) (db *database, err error) {
	db = &database{
		db:                    db,
		app:                   a,
		hostname:              c.ServerConfig.Host,
		defaultCollectionSize: c.DatabaseConfig.DefaultCollectionPageSize,
		maxCollectionPageSize: c.DatabaseConfig.MaxCollectionPageSize,
	}
	return
}

func (d *database) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	return d.inboxes.Contains(util.Context{c}, inbox, id)
}

func (d *database) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.inboxes.GetPage
	last := d.inboxes.GetLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		inboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionSize,
		any,
		last)
}

func (d *database) GetPublicInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.inboxes.GetPublicPage
	last := d.inboxes.GetPublicLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		inboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionSize,
		any,
		last)
}

// NOTE: This only prepends the FIRST item in the orderedItems property.
func (d *database) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	oi := inbox.GetActivityStreamsOrderedItems()
	if oi == nil || oi.Len() == 0 {
		return nil
	}
	id, err := pub.GetId(oi.At(0))
	if err != nil {
		return err
	}
	inboxIRI, err := pub.ToId(inbox)
	if err != nil {
		return err
	}
	d.inboxes.PrependItem(util.Context{u}, paths.Normalize(inboxIRI), id)
}

func (d *database) Owns(c context.Context, id *url.URL) (owns bool, err error) {
	owns = id.Host == d.hostname
	return
}

func (d *database) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return d.users.ActorIDForOutbox(util.Context{c}, paths.Normalize(outboxIRI))
}

func (d *database) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return d.users.ActorIDForInbox(util.Context{c}, paths.Normalize(inboxIRI))
}

func (d *database) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	return d.outboxes.OutboxForInbox(util.Context{c}, paths.Normalize(inboxIRI))
}

func (d *database) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	var r *sql.Rows
	r, err = d.exists.QueryContext(c, id.String())
	if err != nil {
		return
	}
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking exists")
			return
		}
		if err = r.Scan(&exists); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	var r *sql.Rows
	r, err = d.get.QueryContext(c, id.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when getting from db for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	value, err = streams.ToType(c, m)
	return
}

func (d *database) Create(c context.Context, asType vocab.Type) (err error) {
	var m map[string]interface{}
	m, err = streams.Serialize(asType)
	if err != nil {
		return
	}
	var b []byte
	b, err = json.Marshal(m)
	if err != nil {
		return
	}
	var id *url.URL
	id, err = pub.GetId(asType)
	if err != nil {
		return
	}
	var owns bool
	if owns, err = d.Owns(c, id); err != nil {
		return
	} else if owns {
		_, err = d.localCreate.ExecContext(c, string(b))
		return
	} else {
		_, err = d.fedCreate.ExecContext(c, string(b))
		return
	}
}

func (d *database) Update(c context.Context, asType vocab.Type) (err error) {
	var m map[string]interface{}
	m, err = streams.Serialize(asType)
	if err != nil {
		return
	}
	var b []byte
	b, err = json.Marshal(m)
	if err != nil {
		return
	}
	var id *url.URL
	id, err = pub.GetId(asType)
	if err != nil {
		return
	}
	var owns bool
	if owns, err = d.Owns(c, id); err != nil {
		return
	} else if owns {
		_, err = d.localUpdate.ExecContext(c, id.String(), string(b))
		return
	} else {
		_, err = d.fedUpdate.ExecContext(c, id.String(), string(b))
		return
	}
}

func (d *database) Delete(c context.Context, id *url.URL) (err error) {
	var owns bool
	if owns, err = d.Owns(c, id); err != nil {
		return
	} else if owns {
		_, err = d.localDelete.ExecContext(c, id.String())
		return
	} else {
		_, err = d.fedDelete.ExecContext(c, id.String())
		return
	}
}

func (d *database) GetOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.outboxes.GetPage
	last := d.outboxes.GetLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		outboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionSize,
		any,
		last)
}

func (d *database) GetPublicOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	any := d.outboxes.GetPublicPage
	last := d.outboxes.GetPublicLastPage
	return services.DoOrderedCollectionPagination(util.Context{c},
		outboxIRI,
		d.defaultCollectionSize,
		d.maxCollectionSize,
		any,
		last)
}

// NOTE: This only prepends the FIRST item in the orderedItems property.
func (d *database) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	oi := outbox.GetActivityStreamsOrderedItems()
	if oi == nil || oi.Len() == 0 {
		return nil
	}
	id, err := pub.GetId(oi.At(0))
	if err != nil {
		return err
	}
	outboxIRI, err := pub.ToId(outbox)
	if err != nil {
		return err
	}
	d.outboxes.PrependItem(util.Context{u}, paths.Normalize(outboxIRI), id)
}

func (d *database) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	var r *sql.Rows
	r, err = d.followers.QueryContext(c, actorIRI.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when fetching followers for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	var res *streams.JSONResolver
	res, err = streams.NewJSONResolver(func(ctx context.Context, i vocab.ActivityStreamsCollection) error {
		followers = i
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, m)
	return
}

func (d *database) Following(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	var r *sql.Rows
	r, err = d.following.QueryContext(c, actorIRI.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when fetching following for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	var res *streams.JSONResolver
	res, err = streams.NewJSONResolver(func(ctx context.Context, i vocab.ActivityStreamsCollection) error {
		followers = i
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, m)
	return
}

func (d *database) Liked(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	var r *sql.Rows
	r, err = d.liked.QueryContext(c, actorIRI.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when fetching liked for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	var res *streams.JSONResolver
	res, err = streams.NewJSONResolver(func(ctx context.Context, i vocab.ActivityStreamsCollection) error {
		followers = i
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, m)
	return
}
