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
)

// Model handles managing a single database type.
type Model interface {
	Prepare(*sql.DB, SqlDialect) error
	CreateTable(*sql.Tx, SqlDialect) error
	Close()
}

// stmtPair make a pair of **sql.Stmt and its associated SQL string.
//
// The goal is to populate *stmt based on the associated sqlStr.
type stmtPair struct {
	stmt   **sql.Stmt
	sqlStr string
}

// prepareStmtPair is a mapper that populates the stmtPair.stmt.
func prepareStmtPair(db *sql.DB, s stmtPair) (err error) {
	*s.stmt, err = db.Prepare(s.sqlStr)
	return err
}

// stmtPairs are a list of stmtPair.
type stmtPairs []stmtPair

// prepareStmtPairs turns stmtPairs into a single error, with a side effect of
// populating all stmt.
func prepareStmtPairs(db *sql.DB, s stmtPairs) (err error) {
	doIfNoErr := func(p stmtPair, fn func(*sql.DB, stmtPair) error) error {
		if err == nil {
			return fn(db, p)
		}
		return err
	}
	for _, p := range s {
		err = doIfNoErr(p, prepareStmtPair)
	}
	return
}
