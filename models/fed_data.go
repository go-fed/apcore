package models

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/util"
)

var _ Model = &FedData{}

// FedData is a Model that provides additional database methods for
// ActivityStreams data received from federated peers.
type FedData struct {
	fedCreate *sql.Stmt
	fedUpdate *sql.Stmt
	fedDelete *sql.Stmt
}

func (f *FedData) Prepare(db *sql.DB, s SqlDialect) error {
	var err error
	f.fedCreate, err = db.Prepare(s.FedCreate())
	if err != nil {
		return err
	}
	f.fedUpdate, err = db.Prepare(s.FedUpdate())
	if err != nil {
		return err
	}
	f.fedDelete, err = db.Prepare(s.FedDelete())
	if err != nil {
		return err
	}
	return nil
}

func (f *FedData) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateFedDataTable())
	return err
}

func (f *FedData) Close() {
	f.fedCreate.Close()
	f.fedUpdate.Close()
	f.fedDelete.Close()
}

// Create inserts the federated data into the table.
func (f *FedData) Create(c util.Context, tx *sql.Tx, v vocab.Type) error {
	_, err := tx.Stmt(f.fedCreate).ExecContext(c, ActivityStreams{v})
	return err
}

// Update replaces the federated data for the specified IRI.
func (f *FedData) Update(c util.Context, tx *sql.Tx, fedIDIRI *url.URL, v vocab.Type) error {
	_, err := tx.Stmt(f.fedUpdate).ExecContext(c, fedIDIRI.String(), ActivityStreams{v})
	return err
}

// Delete removes the federated data with the specified IRI.
func (f *FedData) Delete(c util.Context, tx *sql.Tx, fedIDIRI *url.URL) error {
	_, err := tx.Stmt(f.fedDelete).ExecContext(c, fedIDIRI.String())
	return err
}
