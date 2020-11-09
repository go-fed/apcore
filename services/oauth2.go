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
	"gopkg.in/oauth2.v3"
)

var _ oauth2.ClientStore = &OAuth2{}
var _ oauth2.TokenStore = &OAuth2{}

// OAuth2 implements services for the oauth2 server package.
type OAuth2 struct {
	DB     *sql.DB
	Client *models.ClientInfos
	Token  *models.TokenInfos
}

// TODO: Somehow pass in context instead
func (o *OAuth2) context() util.Context {
	return util.Context{context.Background()}
}

func (o *OAuth2) GetByID(id string) (ci oauth2.ClientInfo, err error) {
	c := o.context()
	return ci, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ci, err = o.Client.GetByID(c, tx, id)
		return err
	})
}

func (o *OAuth2) Create(info oauth2.TokenInfo) error {
	c := o.context()
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.Create(c, tx, info)
	})
}

func (o *OAuth2) RemoveByCode(code string) error {
	c := o.context()
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.RemoveByCode(c, tx, code)
	})
}

func (o *OAuth2) RemoveByAccess(access string) error {
	c := o.context()
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.RemoveByAccess(c, tx, access)
	})
}

func (o *OAuth2) RemoveByRefresh(refresh string) error {
	c := o.context()
	return doInTx(c, o.DB, func(tx *sql.Tx) error {
		return o.Token.RemoveByRefresh(c, tx, refresh)
	})
}

func (o *OAuth2) GetByCode(code string) (ti oauth2.TokenInfo, err error) {
	c := o.context()
	return ti, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ti, err = o.Token.GetByCode(c, tx, code)
		return err
	})
}

func (o *OAuth2) GetByAccess(access string) (ti oauth2.TokenInfo, err error) {
	c := o.context()
	return ti, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ti, err = o.Token.GetByAccess(c, tx, access)
		return err
	})
}

func (o *OAuth2) GetByRefresh(refresh string) (ti oauth2.TokenInfo, err error) {
	c := o.context()
	return ti, doInTx(c, o.DB, func(tx *sql.Tx) error {
		ti, err = o.Token.GetByRefresh(c, tx, refresh)
		return err
	})
}
