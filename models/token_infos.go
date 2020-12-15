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
	"github.com/go-oauth2/oauth2/v4"
)

var _ oauth2.TokenInfo = &TokenInfo{}

type TokenInfo struct {
	ClientID       string
	UserID         string
	RedirectURI    string
	Scope          string
	Code           sql.NullString
	CodeCreated    sql.NullTime
	CodeExpires    NullDuration
	Access         sql.NullString
	AccessCreated  sql.NullTime
	AccessExpires  NullDuration
	Refresh        sql.NullString
	RefreshCreated sql.NullTime
	RefreshExpires NullDuration
}

func (t *TokenInfo) New() oauth2.TokenInfo {
	return &TokenInfo{}
}

func (t *TokenInfo) GetClientID() string {
	return t.ClientID
}

func (t *TokenInfo) SetClientID(s string) {
	t.ClientID = s
}

func (t *TokenInfo) GetUserID() string {
	return t.UserID
}

func (t *TokenInfo) SetUserID(s string) {
	t.UserID = s
}

func (t *TokenInfo) GetRedirectURI() string {
	return t.RedirectURI
}

func (t *TokenInfo) SetRedirectURI(s string) {
	t.RedirectURI = s
}

func (t *TokenInfo) GetScope() string {
	return t.Scope
}

func (t *TokenInfo) SetScope(s string) {
	t.Scope = s
}

func (t *TokenInfo) GetCode() string {
	if t.Code.Valid {
		return t.Code.String
	}
	return ""
}

func (t *TokenInfo) SetCode(s string) {
	t.Code.String = s
	t.Code.Valid = true
}

func (t *TokenInfo) GetCodeCreateAt() time.Time {
	if t.CodeCreated.Valid {
		return t.CodeCreated.Time
	}
	return time.Time{}
}

func (t *TokenInfo) SetCodeCreateAt(s time.Time) {
	t.CodeCreated.Time = s
	t.CodeCreated.Valid = true
}

func (t *TokenInfo) GetCodeExpiresIn() time.Duration {
	if t.CodeExpires.Valid {
		return t.CodeExpires.Duration
	}
	return 0
}

func (t *TokenInfo) SetCodeExpiresIn(s time.Duration) {
	t.CodeExpires.Duration = s
	t.CodeExpires.Valid = true
}

func (t *TokenInfo) GetAccess() string {
	if t.Access.Valid {
		return t.Access.String
	}
	return ""
}

func (t *TokenInfo) SetAccess(s string) {
	t.Access.String = s
	t.Access.Valid = true
}

func (t *TokenInfo) GetAccessCreateAt() time.Time {
	if t.AccessCreated.Valid {
		return t.AccessCreated.Time
	}
	return time.Time{}
}

func (t *TokenInfo) SetAccessCreateAt(s time.Time) {
	t.AccessCreated.Time = s
	t.AccessCreated.Valid = true
}

func (t *TokenInfo) GetAccessExpiresIn() time.Duration {
	if t.AccessExpires.Valid {
		return t.AccessExpires.Duration
	}
	return 0
}

func (t *TokenInfo) SetAccessExpiresIn(s time.Duration) {
	t.AccessExpires.Duration = s
	t.AccessExpires.Valid = true
}

func (t *TokenInfo) GetRefresh() string {
	if t.Refresh.Valid {
		return t.Refresh.String
	}
	return ""
}

func (t *TokenInfo) SetRefresh(s string) {
	t.Refresh.String = s
	t.Refresh.Valid = true
}

func (t *TokenInfo) GetRefreshCreateAt() time.Time {
	if t.RefreshCreated.Valid {
		return t.RefreshCreated.Time
	}
	return time.Time{}
}

func (t *TokenInfo) SetRefreshCreateAt(s time.Time) {
	t.RefreshCreated.Time = s
	t.RefreshCreated.Valid = true
}

func (t *TokenInfo) GetRefreshExpiresIn() time.Duration {
	if t.RefreshExpires.Valid {
		return t.RefreshExpires.Duration
	}
	return 0
}

func (t *TokenInfo) SetRefreshExpiresIn(s time.Duration) {
	t.RefreshExpires.Duration = s
	t.RefreshExpires.Valid = true
}

func (t *TokenInfo) scanFromSingleRow(r singleRow) error {
	return r.Scan(&(t.ClientID),
		&(t.UserID),
		&(t.RedirectURI),
		&(t.Scope),
		&(t.Code),
		&(t.CodeCreated),
		&(t.CodeExpires),
		&(t.Access),
		&(t.AccessCreated),
		&(t.AccessExpires),
		&(t.Refresh),
		&(t.RefreshCreated),
		&(t.RefreshExpires))
}

// TokenInfos is a Model that provides additional database methods for OAuth2
// token information.
type TokenInfos struct {
	createTokenInfo *sql.Stmt
	removeByCode    *sql.Stmt
	removeByAccess  *sql.Stmt
	removeByRefresh *sql.Stmt
	getByCode       *sql.Stmt
	getByAccess     *sql.Stmt
	getByRefresh    *sql.Stmt
}

func (t *TokenInfos) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(t.createTokenInfo), s.CreateTokenInfo()},
			{&(t.removeByCode), s.RemoveTokenInfoByCode()},
			{&(t.removeByAccess), s.RemoveTokenInfoByAccess()},
			{&(t.removeByRefresh), s.RemoveTokenInfoByRefresh()},
			{&(t.getByCode), s.GetTokenInfoByCode()},
			{&(t.getByAccess), s.GetTokenInfoByAccess()},
			{&(t.getByRefresh), s.GetTokenInfoByRefresh()},
		})
}

func (t *TokenInfos) CreateTable(tx *sql.Tx, s SqlDialect) error {
	_, err := tx.Exec(s.CreateTokenInfosTable())
	return err
}

func (t *TokenInfos) Close() {
	t.createTokenInfo.Close()
	t.removeByCode.Close()
	t.removeByAccess.Close()
	t.removeByRefresh.Close()
	t.getByCode.Close()
	t.getByAccess.Close()
	t.getByRefresh.Close()
}

// Create saves the new token information.
func (t *TokenInfos) Create(c util.Context, tx *sql.Tx, info oauth2.TokenInfo) error {
	r, err := tx.Stmt(t.createTokenInfo).ExecContext(c,
		info.GetClientID(),
		info.GetUserID(),
		info.GetRedirectURI(),
		info.GetScope(),
		info.GetCode(),
		info.GetCodeCreateAt(),
		info.GetCodeExpiresIn(),
		info.GetAccess(),
		info.GetAccessCreateAt(),
		info.GetAccessExpiresIn(),
		info.GetRefresh(),
		info.GetRefreshCreateAt(),
		info.GetRefreshExpiresIn(),
	)
	return mustChangeOneRow(r, err, "TokenInfos.Create")
}

// RemoveByCode deletes the token information based on the authorization code.
func (t *TokenInfos) RemoveByCode(c util.Context, tx *sql.Tx, code string) error {
	r, err := tx.Stmt(t.removeByCode).ExecContext(c, code)
	return mustChangeOneRow(r, err, "TokenInfos.RemoveByCode")
}

// RemoveByAccess deletes the token information based on the access token.
func (t *TokenInfos) RemoveByAccess(c util.Context, tx *sql.Tx, access string) error {
	r, err := tx.Stmt(t.removeByAccess).ExecContext(c, access)
	return mustChangeOneRow(r, err, "TokenInfos.RemoveByAccess")
}

// RemoveByRefresh deletes the token information based on the refresh token.
func (t *TokenInfos) RemoveByRefresh(c util.Context, tx *sql.Tx, refresh string) error {
	r, err := tx.Stmt(t.removeByRefresh).ExecContext(c, refresh)
	return mustChangeOneRow(r, err, "TokenInfos.RemoveByRefresh")
}

// GetByCode fetches tokens based on the authorization code.
func (t *TokenInfos) GetByCode(c util.Context, tx *sql.Tx, code string) (oauth2.TokenInfo, error) {
	rows, err := tx.Stmt(t.getByCode).QueryContext(c, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ti := &TokenInfo{}
	return ti, enforceOneRow(rows, "TokenInfos.GetByCode", func(r singleRow) error {
		return ti.scanFromSingleRow(r)
	})
}

// GetByAccess fetches tokens based on the access token.
func (t *TokenInfos) GetByAccess(c util.Context, tx *sql.Tx, access string) (oauth2.TokenInfo, error) {
	rows, err := tx.Stmt(t.getByAccess).QueryContext(c, access)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ti := &TokenInfo{}
	return ti, enforceOneRow(rows, "TokenInfos.GetByAccess", func(r singleRow) error {
		return ti.scanFromSingleRow(r)
	})
}

// GetByRefresh fetches tokens based on the refresh token.
func (t *TokenInfos) GetByRefresh(c util.Context, tx *sql.Tx, refresh string) (oauth2.TokenInfo, error) {
	rows, err := tx.Stmt(t.getByRefresh).QueryContext(c, refresh)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ti := &TokenInfo{}
	return ti, enforceOneRow(rows, "TokenInfos.GetByRefresh", func(r singleRow) error {
		return ti.scanFromSingleRow(r)
	})
}
