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

package services

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

type Followers struct {
	DB        *sql.DB
	Followers *models.Followers
}

func (f *Followers) ContainsForActor(c util.Context, actor, id *url.URL) (has bool, err error) {
	return has, doInTx(c, f.DB, func(tx *sql.Tx) error {
		has, err = f.Followers.ContainsForActor(c, tx, actor, id)
		return err
	})
}

func (f *Followers) Contains(c util.Context, followers, id *url.URL) (has bool, err error) {
	return has, doInTx(c, f.DB, func(tx *sql.Tx) error {
		has, err = f.Followers.Contains(c, tx, followers, id)
		return err
	})
}

func (f *Followers) GetPage(c util.Context, followers *url.URL, min, n int) (page vocab.ActivityStreamsCollectionPage, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var isEnd bool
		page, isEnd, err = f.Followers.GetPage(c, tx, followers, min, min+n)
		if err != nil {
			return err
		}
		return addNextPrevCol(page, min, n, isEnd)
	})
	return
}

func (f *Followers) GetLastPage(c util.Context, followers *url.URL, n int) (page vocab.ActivityStreamsCollectionPage, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var startIdx int
		page, startIdx, err = f.Followers.GetLastPage(c, tx, followers, n)
		if err != nil {
			return err
		}
		return addNextPrevCol(page, startIdx, n, true)
	})
	return
}

func (f *Followers) PrependItem(c util.Context, followers, item *url.URL) error {
	return doInTx(c, f.DB, func(tx *sql.Tx) error {
		return f.Followers.PrependItem(c, tx, followers, item)
	})
}

func (f *Followers) DeleteItem(c util.Context, followers, item *url.URL) error {
	return doInTx(c, f.DB, func(tx *sql.Tx) error {
		return f.Followers.DeleteItem(c, tx, followers, item)
	})
}

func (f *Followers) GetAllForActor(c util.Context, actor *url.URL) (col vocab.ActivityStreamsCollection, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		col, err = f.Followers.GetAllForActor(c, tx, actor)
		if err != nil {
			return err
		}
		return err
	})
	return
}
