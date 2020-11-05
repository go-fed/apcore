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
	var err error
	d.insertDeliveryAttempt, err = db.Prepare(s.InsertAttempt())
	if err != nil {
		return err
	}
	d.markDeliveryAttemptSuccessful, err = db.Prepare(s.MarkSuccessfulAttempt())
	if err != nil {
		return err
	}
	d.markDeliveryAttemptFailed, err = db.Prepare(s.MarkFailedAttempt())
	if err != nil {
		return err
	}
	return nil
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
