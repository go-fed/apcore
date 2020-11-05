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

type Liked struct {
	DB    *sql.DB
	Liked *models.Liked
}

func (f *Liked) ContainsForActor(c util.Context, actor, id *url.URL) (has bool, err error) {
	return has, doInTx(c, f.DB, func(tx *sql.Tx) error {
		has, err = f.Liked.ContainsForActor(c, tx, actor, id)
		return err
	})
}

func (f *Liked) Contains(c util.Context, liked, id *url.URL) (has bool, err error) {
	return has, doInTx(c, f.DB, func(tx *sql.Tx) error {
		has, err = f.Liked.Contains(c, tx, liked, id)
		return err
	})
}

func (f *Liked) GetPage(c util.Context, liked *url.URL, min, n int) (page vocab.ActivityStreamsCollectionPage, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var isEnd bool
		page, isEnd, err = f.Liked.GetPage(c, tx, liked, min, min+n)
		if err != nil {
			return err
		}
		return addNextPrevCol(page, min, n, isEnd)
	})
	return
}

func (f *Liked) GetLastPage(c util.Context, liked *url.URL, n int) (page vocab.ActivityStreamsCollectionPage, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var startIdx int
		page, startIdx, err = f.Liked.GetLastPage(c, tx, liked, n)
		if err != nil {
			return err
		}
		return addNextPrevCol(page, startIdx, n, true)
	})
	return
}

func (f *Liked) PrependItem(c util.Context, liked, item *url.URL) error {
	return doInTx(c, f.DB, func(tx *sql.Tx) error {
		return f.Liked.PrependItem(c, tx, liked, item)
	})
}

func (f *Liked) DeleteItem(c util.Context, liked, item *url.URL) error {
	return doInTx(c, f.DB, func(tx *sql.Tx) error {
		return f.Liked.DeleteItem(c, tx, liked, item)
	})
}

func (f *Liked) GetAllForActor(c util.Context, actor *url.URL) (col vocab.ActivityStreamsCollection, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		col, err = f.Liked.GetAllForActor(c, tx, actor)
		if err != nil {
			return err
		}
		return err
	})
	return
}
