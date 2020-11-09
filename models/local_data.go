package models

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/apcore/util"
)

var _ Model = &LocalData{}

// LocalData is a Model that provides additional database methods for
// ActivityStreams data generated by this instance.
type LocalData struct {
	exists      *sql.Stmt
	get         *sql.Stmt
	localCreate *sql.Stmt
	localUpdate *sql.Stmt
	localDelete *sql.Stmt
}

func (f *LocalData) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(f.exists), s.LocalExists()},
			{&(f.get), s.LocalGet()},
			{&(f.localCreate), s.LocalCreate()},
			{&(f.localUpdate), s.LocalUpdate()},
			{&(f.localDelete), s.LocalDelete()},
		})
}

func (f *LocalData) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateLocalDataTable())
	return err
}

func (f *LocalData) Close() {
	f.exists.Close()
	f.get.Close()
	f.localCreate.Close()
	f.localUpdate.Close()
	f.localDelete.Close()
}

// Exists determines if the ID is stored in the local table.
func (f *LocalData) Exists(c util.Context, tx *sql.Tx, id *url.URL) (exists bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(f.exists).QueryContext(c, id.String())
	if err != nil {
		return
	}
	defer rows.Close()
	err = enforceOneRow(rows, "LocalData.Exists", func(r singleRow) error {
		return r.Scan(&exists)
	})
	return
}

// Get retrieves the ID from the local table.
func (f *LocalData) Get(c util.Context, tx *sql.Tx, id *url.URL) (v ActivityStreams, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(f.get).QueryContext(c, id.String())
	if err != nil {
		return
	}
	defer rows.Close()
	err = enforceOneRow(rows, "LocalData.Get", func(r singleRow) error {
		return r.Scan(&v)
	})
	return
}

// Create inserts the local data into the table.
func (f *LocalData) Create(c util.Context, tx *sql.Tx, v ActivityStreams) error {
	r, err := tx.Stmt(f.localCreate).ExecContext(c, v)
	return mustChangeOneRow(r, err, "LocalData.Create")
}

// Update replaces the local data for the specified IRI.
func (f *LocalData) Update(c util.Context, tx *sql.Tx, localIDIRI *url.URL, v ActivityStreams) error {
	r, err := tx.Stmt(f.localUpdate).ExecContext(c, localIDIRI.String(), v)
	return mustChangeOneRow(r, err, "LocalData.Update")
}

// Delete removes the local data with the specified IRI.
func (f *LocalData) Delete(c util.Context, tx *sql.Tx, localIDIRI *url.URL) error {
	r, err := tx.Stmt(f.localDelete).ExecContext(c, localIDIRI.String())
	return mustChangeOneRow(r, err, "LocalData.Delete")
}