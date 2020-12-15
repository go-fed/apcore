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

	"github.com/go-fed/apcore/util"
	"github.com/go-oauth2/oauth2/v4"
)

var _ oauth2.ClientInfo = &ClientInfo{}

type ClientInfo struct {
	ID     string
	Secret string
	Domain string
	UserID string
}

func (c *ClientInfo) GetID() string {
	return c.ID
}

func (c *ClientInfo) GetSecret() string {
	return c.Secret
}

func (c *ClientInfo) GetDomain() string {
	return c.Domain
}

func (c *ClientInfo) GetUserID() string {
	return c.UserID
}

// ClientInfos is a Model that provides additional database methods for OAuth2
// client information.
type ClientInfos struct {
	create  *sql.Stmt
	getByID *sql.Stmt
}

func (c *ClientInfos) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(c.create), s.CreateClientInfo()},
			{&(c.getByID), s.GetClientInfoByID()},
		})
}

func (c *ClientInfos) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateClientInfosTable())
	return err
}

func (c *ClientInfos) Close() {
	c.create.Close()
	c.getByID.Close()
}

// Create adds a ClientInfo into the database.
func (c *ClientInfos) Create(ctx util.Context, tx *sql.Tx, info oauth2.ClientInfo) (id string, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(c.create).QueryContext(ctx,
		info.GetSecret(),
		info.GetDomain(),
		info.GetUserID())
	if err != nil {
		return
	}
	defer rows.Close()
	return id, enforceOneRow(rows, "ClientInfos.Create", func(r singleRow) error {
		return r.Scan(&(id))
	})
}

// GetByID fetches ClientInfo based on its id.
func (c *ClientInfos) GetByID(ctx util.Context, tx *sql.Tx, id string) (oauth2.ClientInfo, error) {
	rows, err := tx.Stmt(c.getByID).QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ci := &ClientInfo{}
	return ci, enforceOneRow(rows, "ClientInfos.GetByID", func(r singleRow) error {
		return r.Scan(&(ci.ID), &(ci.Secret), &(ci.Domain), &(ci.UserID))
	})
}
