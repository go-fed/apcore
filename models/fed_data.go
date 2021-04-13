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

var _ Model = &FedData{}

// FedData is a Model that provides additional database methods for
// ActivityStreams data received from federated peers.
type FedData struct {
	exists    *sql.Stmt
	get       *sql.Stmt
	fedCreate *sql.Stmt
	fedUpdate *sql.Stmt
	fedDelete *sql.Stmt
}

func (f *FedData) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(f.exists), s.FedExists()},
			{&(f.get), s.FedGet()},
			{&(f.fedCreate), s.FedCreate()},
			{&(f.fedUpdate), s.FedUpdate()},
			{&(f.fedDelete), s.FedDelete()},
		})
}

func (f *FedData) CreateTable(t *sql.Tx, s SqlDialect) error {
	if _, err := t.Exec(s.CreateFedDataTable()); err != nil {
		return err
	}
	_, err := t.Exec(s.CreateIndexIDFedDataTable())
	return err
}

func (f *FedData) Close() {
	f.exists.Close()
	f.get.Close()
	f.fedCreate.Close()
	f.fedUpdate.Close()
	f.fedDelete.Close()
}

// Exists determines if the ID is stored in the federated table.
func (f *FedData) Exists(c util.Context, tx *sql.Tx, id *url.URL) (exists bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(f.exists).QueryContext(c, id.String())
	if err != nil {
		return
	}
	defer rows.Close()
	err = enforceOneRow(rows, "FedData.Exists", func(r SingleRow) error {
		return r.Scan(&exists)
	})
	return
}

// Get retrieves the ID from the federated table.
func (f *FedData) Get(c util.Context, tx *sql.Tx, id *url.URL) (v ActivityStreams, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(f.get).QueryContext(c, id.String())
	if err != nil {
		return
	}
	defer rows.Close()
	err = enforceOneRow(rows, "FedData.Get", func(r SingleRow) error {
		return r.Scan(&v)
	})
	return
}

// Create inserts the federated data into the table.
func (f *FedData) Create(c util.Context, tx *sql.Tx, v ActivityStreams) error {
	v.SanitizeContentSummaryHTML()
	r, err := tx.Stmt(f.fedCreate).ExecContext(c, v)
	return mustChangeOneRow(r, err, "FedData.Create")
}

// Update replaces the federated data for the specified IRI.
func (f *FedData) Update(c util.Context, tx *sql.Tx, fedIDIRI *url.URL, v ActivityStreams) error {
	v.SanitizeContentSummaryHTML()
	r, err := tx.Stmt(f.fedUpdate).ExecContext(c, fedIDIRI.String(), v)
	return mustChangeOneRow(r, err, "FedData.Update")
}

// Delete removes the federated data with the specified IRI.
func (f *FedData) Delete(c util.Context, tx *sql.Tx, fedIDIRI *url.URL) error {
	r, err := tx.Stmt(f.fedDelete).ExecContext(c, fedIDIRI.String())
	return mustChangeOneRow(r, err, "FedData.Delete")
}
