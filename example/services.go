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

package main

import (
	"context"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

func getLatestPublicNotes(ctx context.Context, db app.Database) (notes []vocab.ActivityStreamsNote, err error) {
	c := util.Context{ctx}
	var rz *streams.TypeResolver
	rz, err = streams.NewTypeResolver(func(c context.Context, note vocab.ActivityStreamsNote) error {
		notes = append(notes, note)
		return nil
	})
	if err != nil {
		return
	}
	txb := db.Begin()
	txb.Query(`SELECT payload FROM %[1]slocal_data
WHERE payload->>'type' = 'Note' AND (
  payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
  OR payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public')
ORDER BY create_time DESC LIMIT 10`,
		func(r app.SingleRow) error {
			var v models.ActivityStreams
			if err := r.Scan(&v); err != nil {
				return err
			}
			return rz.Resolve(c.Context, v.Type)
		})
	err = txb.Do(c)
	return
}

func getUsers(ctx context.Context, db app.Database) (ppl []vocab.ActivityStreamsPerson, err error) {
	c := util.Context{ctx}
	var rz *streams.TypeResolver
	rz, err = streams.NewTypeResolver(func(c context.Context, pn vocab.ActivityStreamsPerson) error {
		ppl = append(ppl, pn)
		return nil
	})
	if err != nil {
		return
	}
	txb := db.Begin()
	txb.Query(`SELECT actor FROM %[1]susers
WHERE actor->>'type' = 'Person'
ORDER BY create_time DESC`,
		func(r app.SingleRow) error {
			var v models.ActivityStreams
			if err := r.Scan(&v); err != nil {
				return err
			}
			return rz.Resolve(c.Context, v.Type)
		})
	err = txb.Do(c)
	return
}
