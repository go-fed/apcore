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

type Inboxes struct {
	DB      *sql.DB
	Inboxes *models.Inboxes
}

func (i *Inboxes) GetPage(c util.Context, inbox *url.URL, min, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	err = doInTx(c, i.DB, func(tx *sql.Tx) error {
		page, err = i.Inboxes.GetPage(c, tx, inbox, min, min+n)
		return err
	})
	return
}

func (i *Inboxes) GetPublicPage(c util.Context, inbox *url.URL, min, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	err = doInTx(c, i.DB, func(tx *sql.Tx) error {
		page, err = i.Inboxes.GetPublicPage(c, tx, inbox, min, min+n)
		return err
	})
	return
}

func (i *Inboxes) GetLastPage(c util.Context, inbox *url.URL, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	err = doInTx(c, i.DB, func(tx *sql.Tx) error {
		page, err = i.Inboxes.GetLastPage(c, tx, inbox, n)
		return err
	})
	return
}
func (i *Inboxes) GetPublicLastPage(c util.Context, inbox *url.URL, n int) (page models.ActivityStreamsOrderedCollectionPage, err error) {
	err = doInTx(c, i.DB, func(tx *sql.Tx) error {
		page, err = i.Inboxes.GetPublicLastPage(c, tx, inbox, n)
		return err
	})
	return
}

func (i *Inboxes) ContainsForActor(c util.Context, actor, id *url.URL) (has bool, err error) {
	return has, doInTx(c, i.DB, func(tx *sql.Tx) error {
		has, err = i.Inboxes.ContainsForActor(c, tx, actor, id)
		return err
	})
	return
}

func (i *Inboxes) Contains(c util.Context, inbox, id *url.URL) (has bool, err error) {
	return has, doInTx(c, i.DB, func(tx *sql.Tx) error {
		has, err = i.Inboxes.Contains(c, tx, inbox, id)
		return err
	})
	return
}

func (i *Inboxes) PrependItem(c util.Context, inbox, item *url.URL) error {
	return doInTx(c, i.DB, func(tx *sql.Tx) error {
		return i.Inboxes.PrependInboxItem(c, tx, inbox, item)
	})
}

func (i *Inboxes) DeleteItem(c util.Context, inbox, item *url.URL) error {
	return doInTx(c, i.DB, func(tx *sql.Tx) error {
		return i.Inboxes.DeleteInboxItem(c, tx, inbox, item)
	})
}
