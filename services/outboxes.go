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

	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

type Outboxes struct {
	DB       *sql.DB
	Outboxes *models.Outboxes
}

func (i *Outboxes) GetPage(c util.Context, outbox *url.URL, min, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	return page, doInTx(c, i.DB, func(tx *sql.Tx) error {
		var isEnd bool
		page, isEnd, err = i.Outboxes.GetPage(c, tx, outbox, min, min+n)
		if err != nil {
			return err
		}
		return addNextPrev(page, min, n, isEnd)
	})
}

func (i *Outboxes) GetPublicPage(c util.Context, outbox *url.URL, min, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	return page, doInTx(c, i.DB, func(tx *sql.Tx) error {
		var isEnd bool
		page, isEnd, err = i.Outboxes.GetPublicPage(c, tx, outbox, min, min+n)
		if err != nil {
			return err
		}
		return addNextPrev(page, min, n, isEnd)
	})
}

func (i *Outboxes) GetLastPage(c util.Context, outbox *url.URL, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	return page, doInTx(c, i.DB, func(tx *sql.Tx) error {
		var startIdx int
		page, startIdx, err = i.Outboxes.GetLastPage(c, tx, outbox, n)
		if err != nil {
			return err
		}
		return addNextPrev(page, startIdx, n, true)
	})
}

func (i *Outboxes) GetPublicLastPage(c util.Context, outbox *url.URL, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	return page, doInTx(c, i.DB, func(tx *sql.Tx) error {
		var startIdx int
		page, startIdx, err = i.Outboxes.GetPublicLastPage(c, tx, outbox, n)
		if err != nil {
			return err
		}
		return addNextPrev(page, startIdx, n, true)
	})
}

func (i *Outboxes) OutboxForInbox(c util.Context, inboxIRI *url.URL) (outboxIRI models.URL, err error) {
	return outboxIRI, doInTx(c, i.DB, func(tx *sql.Tx) error {
		outboxIRI, err = i.Outboxes.OutboxForInbox(c, tx, inboxIRI)
		return err
	})
}

func (i *Outboxes) PrependItem(c util.Context, outbox, item *url.URL) error {
	return doInTx(c, i.DB, func(tx *sql.Tx) error {
		return i.Outboxes.PrependOutboxItem(c, tx, outbox, item)
	})
}

func (i *Outboxes) DeleteItem(c util.Context, outbox, item *url.URL) error {
	return doInTx(c, i.DB, func(tx *sql.Tx) error {
		return i.Outboxes.DeleteOutboxItem(c, tx, outbox, item)
	})
}
