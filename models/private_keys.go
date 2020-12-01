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

	"github.com/go-fed/apcore/util"
)

var _ Model = &PrivateKeys{}

// PrivateKeys is a Model that provides additional database methods for the
// PrivateKey type.
type PrivateKeys struct {
	createPrivateKey *sql.Stmt
	getByUserID      *sql.Stmt
	getInstanceActor *sql.Stmt
}

func (p *PrivateKeys) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(p.createPrivateKey), s.CreatePrivateKey()},
			{&(p.getByUserID), s.GetPrivateKeyByUserID()},
			{&(p.getInstanceActor), s.GetPrivateKeyForInstanceActor()},
		})
}

func (p *PrivateKeys) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreatePrivateKeysTable())
	return err
}

func (p *PrivateKeys) Close() {
	p.createPrivateKey.Close()
	p.getByUserID.Close()
	p.getInstanceActor.Close()
}

// Create a new private key entry in the database.
func (p *PrivateKeys) Create(c util.Context, tx *sql.Tx, userID, purpose string, privKey []byte) error {
	r, err := tx.Stmt(p.createPrivateKey).ExecContext(c, userID, purpose, privKey)
	return mustChangeOneRow(r, err, "PrivateKeys.Create")
}

// GetByUserID fetches a private key by the userID and purpose of the key.
func (p *PrivateKeys) GetByUserID(c util.Context, tx *sql.Tx, userID, purpose string) (b []byte, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(p.getByUserID).QueryContext(c, userID, purpose)
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "PrivateKeys.GetByUserID", func(r singleRow) error {
		return r.Scan(&(b))
	})
}

// GetInstanceActor fetches a private key for the single instance actor.
func (p *PrivateKeys) GetInstanceActor(c util.Context, tx *sql.Tx, purpose string) (b []byte, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(p.getInstanceActor).QueryContext(c, purpose)
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "PrivateKeys.GetInstanceActor", func(r singleRow) error {
		return r.Scan(&(b))
	})
}
