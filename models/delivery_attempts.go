// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
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

package models

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/apcore/util"
)

// These constants are used to mark the simple state of the delivery attempt.
const (
	newDeliveryAttempt     = "new"
	successDeliveryAttempt = "success"
	failedDeliveryAttempt  = "failed"
)

var _ Model = &DeliveryAttempts{}

// DeliveryAttempts is a Model that provides additional database methods for
// delivery attempts.
type DeliveryAttempts struct {
	insertDeliveryAttempt         *sql.Stmt
	markDeliveryAttemptSuccessful *sql.Stmt
	markDeliveryAttemptFailed     *sql.Stmt
}

func (d *DeliveryAttempts) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(d.insertDeliveryAttempt), s.InsertAttempt()},
			{&(d.markDeliveryAttemptSuccessful), s.MarkSuccessfulAttempt()},
			{&(d.markDeliveryAttemptFailed), s.MarkFailedAttempt()},
		})
}

func (d *DeliveryAttempts) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateDeliveryAttemptsTable())
	return err
}

func (d *DeliveryAttempts) Close() {
	d.insertDeliveryAttempt.Close()
	d.markDeliveryAttemptSuccessful.Close()
	d.markDeliveryAttemptFailed.Close()
}

// Create a new delivery attempt.
func (d *DeliveryAttempts) Create(c util.Context, tx *sql.Tx, from string, toActor *url.URL, payload []byte) (id string, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(d.insertDeliveryAttempt).QueryContext(c,
		from,
		toActor.String(),
		payload,
		newDeliveryAttempt)
	if err != nil {
		return
	}
	defer rows.Close()
	return id, enforceOneRow(rows, "DeliveryAttempts.Create", func(r singleRow) error {
		return r.Scan(&(id))
	})
}

// MarkSuccessful marks a delivery attempt as successful.
func (d *DeliveryAttempts) MarkSuccessful(c util.Context, tx *sql.Tx, id string) error {
	r, err := tx.Stmt(d.markDeliveryAttemptSuccessful).ExecContext(c,
		id,
		successDeliveryAttempt)
	return mustChangeOneRow(r, err, "DeliveryAttempts.MarkSuccessful")
}

// MarkFailed marks a delivery attempt as failed.
func (d *DeliveryAttempts) MarkFailed(c util.Context, tx *sql.Tx, id string) error {
	r, err := tx.Stmt(d.markDeliveryAttemptFailed).ExecContext(c,
		id,
		failedDeliveryAttempt)
	return mustChangeOneRow(r, err, "DeliveryAttempts.MarkFailed")
}
