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
	"time"

	"github.com/go-fed/apcore/util"
)

// These constants are used to mark the simple state of the delivery attempt.
const (
	newDeliveryAttempt       = "new"
	successDeliveryAttempt   = "success"
	failedDeliveryAttempt    = "failed"
	abandonedDeliveryAttempt = "abandoned"
)

var _ Model = &DeliveryAttempts{}

// DeliveryAttempts is a Model that provides additional database methods for
// delivery attempts.
type DeliveryAttempts struct {
	insertDeliveryAttempt         *sql.Stmt
	markDeliveryAttemptSuccessful *sql.Stmt
	markDeliveryAttemptFailed     *sql.Stmt
	markDeliveryAttemptAbandoned  *sql.Stmt
	firstRetryablePage            *sql.Stmt
	nextRetryablePage             *sql.Stmt
}

func (d *DeliveryAttempts) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(d.insertDeliveryAttempt), s.InsertAttempt()},
			{&(d.markDeliveryAttemptSuccessful), s.MarkSuccessfulAttempt()},
			{&(d.markDeliveryAttemptFailed), s.MarkFailedAttempt()},
			{&(d.markDeliveryAttemptAbandoned), s.MarkAbandonedAttempt()},
			{&(d.firstRetryablePage), s.FirstPageRetryableFailures()},
			{&(d.nextRetryablePage), s.NextPageRetryableFailures()},
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
	return id, enforceOneRow(rows, "DeliveryAttempts.Create", func(r SingleRow) error {
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

// MarkAbandoned marks a delivery attempt as abandoned.
func (d *DeliveryAttempts) MarkAbandoned(c util.Context, tx *sql.Tx, id string) error {
	r, err := tx.Stmt(d.markDeliveryAttemptAbandoned).ExecContext(c,
		id,
		abandonedDeliveryAttempt)
	return mustChangeOneRow(r, err, "DeliveryAttempts.Abandoned")
}

type RetryableFailure struct {
	ID          string
	UserID      string
	DeliverTo   URL
	Payload     []byte
	NAttempts   int
	LastAttempt time.Time
}

// FirstPageFailures obtains the first page of retryable failures.
func (d *DeliveryAttempts) FirstPageFailures(c util.Context, tx *sql.Tx, fetchTime time.Time, n int) (rf []RetryableFailure, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(d.firstRetryablePage).QueryContext(c, failedDeliveryAttempt, fetchTime, n)
	if err != nil {
		return
	}
	defer rows.Close()
	return rf, doForRows(rows, "DeliveryAttempts.FirstPageFailures", func(r SingleRow) error {
		var rt RetryableFailure
		if err := r.Scan(&(rt.ID), &(rt.UserID), &(rt.DeliverTo), &(rt.Payload), &(rt.NAttempts), &(rt.LastAttempt)); err != nil {
			return err
		}
		rf = append(rf, rt)
		return nil
	})
}

// NextPageFailures obtains the next page of retryable failures.
func (d *DeliveryAttempts) NextPageFailures(c util.Context, tx *sql.Tx, prevID string, fetchTime time.Time, n int) (rf []RetryableFailure, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(d.nextRetryablePage).QueryContext(c, failedDeliveryAttempt, fetchTime, n, prevID)
	if err != nil {
		return
	}
	defer rows.Close()
	return rf, doForRows(rows, "DeliveryAttempts.NextPageFailures", func(r SingleRow) error {
		var rt RetryableFailure
		if err := r.Scan(&(rt.ID), &(rt.UserID), &(rt.DeliverTo), &(rt.Payload), &(rt.NAttempts), &(rt.LastAttempt)); err != nil {
			return err
		}
		rf = append(rf, rt)
		return nil
	})
}
