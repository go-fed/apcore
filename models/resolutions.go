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
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/apcore/util"
)

type Resolution struct {
	Time time.Time `json:"time",omitempty`

	// The following are used by Policies
	Matched  bool     `json:"matched",omitempty`
	MatchLog []string `json:"matchLog",omitempty`
}

func (r *Resolution) Logf(s string, i ...interface{}) {
	r.Log(fmt.Sprintf(s, i...))
}

func (r *Resolution) Log(s string) {
	r.MatchLog = append(r.MatchLog, s)
}

type CreateResolution struct {
	PolicyID string
	IRI      *url.URL
	R        Resolution
}

var _ Model = &Resolutions{}

// Resolutions is a Model that provides additional database methods for the
// Resolution type.
type Resolutions struct {
	create *sql.Stmt
}

func (r *Resolutions) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(r.create), s.CreateResolution()},
		})
}

func (r *Resolutions) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateResolutionsTable())
	return err
}

func (r *Resolutions) Close() {
	r.create.Close()
}

// Create a new Resolution
func (r *Resolutions) Create(c util.Context, tx *sql.Tx, cr CreateResolution) error {
	rows, err := tx.Stmt(r.create).ExecContext(c,
		cr.PolicyID,
		cr.IRI.String(),
		cr.R)
	return mustChangeOneRow(rows, err, "Resolutions.Create")
}
