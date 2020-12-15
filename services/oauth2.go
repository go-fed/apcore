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

package services

import (
	"context"
	"database/sql"

	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
	"github.com/go-oauth2/oauth2/v4"
)

var _ oauth2.ClientStore = &OAuth2{}
var _ oauth2.TokenStore = &OAuth2{}

// OAuth2 implements services for the oauth2 server package.
type OAuth2 struct {
	DB     *sql.DB
	Client *models.ClientInfos
	Token  *models.TokenInfos
}

func (o *OAuth2) GetByID(ctx context.Context, id string) (ci oauth2.ClientInfo, err error) {
	c := util.Context{ctx}
	return ci, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ci, err = o.Client.GetByID(c, tx, id)
		return err
	})
}

func (o *OAuth2) Create(ctx context.Context, info oauth2.TokenInfo) error {
	c := util.Context{ctx}
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.Create(c, tx, info)
	})
}

func (o *OAuth2) RemoveByCode(ctx context.Context, code string) error {
	c := util.Context{ctx}
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.RemoveByCode(c, tx, code)
	})
}

func (o *OAuth2) RemoveByAccess(ctx context.Context, access string) error {
	c := util.Context{ctx}
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.RemoveByAccess(c, tx, access)
	})
}

func (o *OAuth2) RemoveByRefresh(ctx context.Context, refresh string) error {
	c := util.Context{ctx}
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.RemoveByRefresh(c, tx, refresh)
	})
}

func (o *OAuth2) GetByCode(ctx context.Context, code string) (ti oauth2.TokenInfo, err error) {
	c := util.Context{ctx}
	return ti, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ti, err = o.Token.GetByCode(c, tx, code)
		return err
	})
}

func (o *OAuth2) GetByAccess(ctx context.Context, access string) (ti oauth2.TokenInfo, err error) {
	c := util.Context{ctx}
	return ti, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ti, err = o.Token.GetByAccess(c, tx, access)
		return err
	})
}

func (o *OAuth2) GetByRefresh(ctx context.Context, refresh string) (ti oauth2.TokenInfo, err error) {
	c := util.Context{ctx}
	return ti, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ti, err = o.Token.GetByRefresh(c, tx, refresh)
		return err
	})
}
