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

type Following struct {
	DB        *sql.DB
	Following *models.Following
}

func (f *Following) ContainsForActor(c util.Context, actor, id *url.URL) (has bool, err error) {
	return has, doInTx(c, f.DB, func(tx *sql.Tx) error {
		has, err = f.Following.ContainsForActor(c, tx, actor, id)
		return err
	})
}

func (f *Following) Contains(c util.Context, following, id *url.URL) (has bool, err error) {
	return has, doInTx(c, f.DB, func(tx *sql.Tx) error {
		has, err = f.Following.Contains(c, tx, following, id)
		return err
	})
}

func (f *Following) GetPage(c util.Context, following *url.URL, min, n int) (page vocab.ActivityStreamsCollectionPage, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var isEnd bool
		var mp models.ActivityStreamsCollectionPage
		mp, isEnd, err = f.Following.GetPage(c, tx, following, min, min+n)
		if err != nil {
			return err
		}
		page = mp.ActivityStreamsCollectionPage
		return addNextPrevCol(page, min, n, isEnd)
	})
	return
}

func (f *Following) GetLastPage(c util.Context, following *url.URL, n int) (page vocab.ActivityStreamsCollectionPage, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var startIdx int
		var mp models.ActivityStreamsCollectionPage
		mp, startIdx, err = f.Following.GetLastPage(c, tx, following, n)
		if err != nil {
			return err
		}
		page = mp.ActivityStreamsCollectionPage
		return addNextPrevCol(page, startIdx, n, true)
	})
	return
}

func (f *Following) PrependItem(c util.Context, following, item *url.URL) error {
	return doInTx(c, f.DB, func(tx *sql.Tx) error {
		return f.Following.PrependItem(c, tx, following, item)
	})
}

func (f *Following) DeleteItem(c util.Context, following, item *url.URL) error {
	return doInTx(c, f.DB, func(tx *sql.Tx) error {
		return f.Following.DeleteItem(c, tx, following, item)
	})
}

func (f *Following) GetAllForActor(c util.Context, actor *url.URL) (col vocab.ActivityStreamsCollection, err error) {
	err = doInTx(c, f.DB, func(tx *sql.Tx) error {
		var mc models.ActivityStreamsCollection
		mc, err = f.Following.GetAllForActor(c, tx, actor)
		if err != nil {
			return err
		}
		col = mc.ActivityStreamsCollection
		return err
	})
	return
}
