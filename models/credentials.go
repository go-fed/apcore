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

package models

import (
	"database/sql"
	"time"

	"github.com/go-fed/apcore/util"
	"github.com/go-fed/oauth2"
)

// Credentials is a Model that provides a first-party proxy to OAuth2 tokens for
// cookies and other first-party storage.
type Credentials struct {
	createCred           *sql.Stmt
	updateCred           *sql.Stmt
	updateCredExpires    *sql.Stmt
	removeCred           *sql.Stmt
	getTokenInfoByCredID *sql.Stmt
}

func (c *Credentials) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(c.createCred), s.CreateFirstPartyCredential()},
			{&(c.updateCred), s.UpdateFirstPartyCredential()},
			{&(c.updateCredExpires), s.UpdateFirstPartyCredentialExpires()},
			{&(c.removeCred), s.RemoveFirstPartyCredential()},
			{&(c.getTokenInfoByCredID), s.GetTokenInfoForCredentialID()},
		})
}

func (c *Credentials) CreateTable(tx *sql.Tx, s SqlDialect) error {
	_, err := tx.Exec(s.CreateFirstPartyCredentialsTable())
	return err
}

func (c *Credentials) Close() {
	c.createCred.Close()
	c.updateCred.Close()
	c.updateCredExpires.Close()
	c.removeCred.Close()
	c.getTokenInfoByCredID.Close()
}

// Create saves the new first party credential.
func (c *Credentials) Create(ctx util.Context, tx *sql.Tx, userID, tokenID string, expires time.Time) (id string, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(c.createCred).QueryContext(ctx, userID, tokenID, expires)
	if err != nil {
		return
	}
	defer rows.Close()
	return id, enforceOneRow(rows, "Credentials.Create", func(r singleRow) error {
		return r.Scan(&(id))
	})
}

func (c *Credentials) Update(ctx util.Context, tx *sql.Tx, id string, info oauth2.TokenInfo) error {
	r, err := tx.Stmt(c.updateCred).ExecContext(ctx,
		id,
		info.GetClientID(),
		info.GetUserID(),
		info.GetRedirectURI(),
		info.GetScope(),
		info.GetCode(),
		info.GetCodeCreateAt(),
		info.GetCodeExpiresIn(),
		info.GetCodeChallenge(),
		info.GetCodeChallengeMethod(),
		info.GetAccess(),
		info.GetAccessCreateAt(),
		info.GetAccessExpiresIn(),
		info.GetRefresh(),
		info.GetRefreshCreateAt(),
		info.GetRefreshExpiresIn(),
	)
	return mustChangeOneRow(r, err, "Credentials.Update")
}

func (c *Credentials) UpdateExpires(ctx util.Context, tx *sql.Tx, id string, expires time.Time) error {
	r, err := tx.Stmt(c.updateCredExpires).ExecContext(ctx, id, expires)
	return mustChangeOneRow(r, err, "Credentials.UpdateExpires")
}

func (c *Credentials) Delete(ctx util.Context, tx *sql.Tx, id string) error {
	r, err := tx.Stmt(c.removeCred).ExecContext(ctx, id)
	return mustChangeOneRow(r, err, "Credentials.Delete")
}

func (c *Credentials) GetTokenInfo(ctx util.Context, tx *sql.Tx, id string) (oauth2.TokenInfo, error) {
	rows, err := tx.Stmt(c.getTokenInfoByCredID).QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ti := &TokenInfo{}
	return ti, enforceOneRow(rows, "Credentials.GetTokenInfo", func(r singleRow) error {
		return ti.scanFromSingleRow(r)
	})
}
