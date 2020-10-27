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

type DeliveryAttempts struct {
	DB               *sql.DB
	DeliveryAttempts *models.DeliveryAttempts
}

func (d *DeliveryAttempts) InsertAttempt(c util.Context, from string, toActor *url.URL, payload []byte) (id string, err error) {
	return id, doInTx(c, d.DB, func(tx *sql.Tx) error {
		id, err = d.DeliveryAttempts.Create(c, tx, from, toActor, payload)
		return err
	})
}

func (d *DeliveryAttempts) MarkSuccessfulAttempt(c util.Context, id string) (err error) {
	return doInTx(c, d.DB, func(tx *sql.Tx) error {
		return d.DeliveryAttempts.MarkSuccessful(c, tx, id)
	})
}

func (d *DeliveryAttempts) MarkRetryFailureAttempt(c util.Context, id string) (err error) {
	return doInTx(c, d.DB, func(tx *sql.Tx) error {
		return d.DeliveryAttempts.MarkFailed(c, tx, id)
	})
}
