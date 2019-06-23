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

package apcore

import (
	"time"

	"gopkg.in/oauth2.v3"
)

var _ oauth2.TokenInfo = &tokenInfo{}

// TODO: Postgres table
type tokenInfo struct {
	clientId       string
	userId         string
	redirectURI    string
	scope          string
	code           string
	codeCreated    time.Time
	codeExpires    time.Duration
	access         string
	accessCreated  time.Time
	accessExpires  time.Duration
	refresh        string
	refreshCreated time.Time
	refreshExpires time.Duration
}

func (t *tokenInfo) New() oauth2.TokenInfo {
	return &tokenInfo{}
}

func (t *tokenInfo) GetClientID() string {
	return t.clientId
}

func (t *tokenInfo) SetClientID(s string) {
	t.clientId = s
}

func (t *tokenInfo) GetUserID() string {
	return t.userId
}

func (t *tokenInfo) SetUserID(s string) {
	t.userId = s
}

func (t *tokenInfo) GetRedirectURI() string {
	return t.redirectURI
}

func (t *tokenInfo) SetRedirectURI(s string) {
	t.redirectURI = s
}

func (t *tokenInfo) GetScope() string {
	return t.scope
}

func (t *tokenInfo) SetScope(s string) {
	t.scope = s
}

func (t *tokenInfo) GetCode() string {
	return t.code
}

func (t *tokenInfo) SetCode(s string) {
	t.code = s
}

func (t *tokenInfo) GetCodeCreateAt() time.Time {
	return t.codeCreated
}

func (t *tokenInfo) SetCodeCreateAt(s time.Time) {
	t.codeCreated = s
}

func (t *tokenInfo) GetCodeExpiresIn() time.Duration {
	return t.codeExpires
}

func (t *tokenInfo) SetCodeExpiresIn(s time.Duration) {
	t.codeExpires = s
}

func (t *tokenInfo) GetAccess() string {
	return t.access
}

func (t *tokenInfo) SetAccess(s string) {
	t.access = s
}

func (t *tokenInfo) GetAccessCreateAt() time.Time {
	return t.accessCreated
}

func (t *tokenInfo) SetAccessCreateAt(s time.Time) {
	t.accessCreated = s
}

func (t *tokenInfo) GetAccessExpiresIn() time.Duration {
	return t.accessExpires
}

func (t *tokenInfo) SetAccessExpiresIn(s time.Duration) {
	t.accessExpires = s
}

func (t *tokenInfo) GetRefresh() string {
	return t.refresh
}

func (t *tokenInfo) SetRefresh(s string) {
	t.refresh = s
}

func (t *tokenInfo) GetRefreshCreateAt() time.Time {
	return t.refreshCreated
}

func (t *tokenInfo) SetRefreshCreateAt(s time.Time) {
	t.refreshCreated = s
}

func (t *tokenInfo) GetRefreshExpiresIn() time.Duration {
	return t.refreshExpires
}

func (t *tokenInfo) SetRefreshExpiresIn(s time.Duration) {
	t.refreshExpires = s
}

var _ oauth2.TokenStore = &tokenStore{}

type tokenStore struct {
	d *database
}

func newTokenStore(d *database) (t *tokenStore, err error) {
	t = &tokenStore{
		d: d,
	}
	return
}

// Create and store the new token information
func (t *tokenStore) Create(info oauth2.TokenInfo) error {
	// TODO
	return nil
}

// Delete the authorization code
func (t *tokenStore) RemoveByCode(code string) error {
	// TODO
	return nil
}

// Use the access token to delete the token information
func (t *tokenStore) RemoveByAccess(access string) error {
	// TODO
	return nil
}

// Use the refresh token to delete the token information
func (t *tokenStore) RemoveByRefresh(refresh string) error {
	// TODO
	return nil
}

// Use the authorization code for token information data
func (t *tokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {
	// TODO
	return nil, nil
}

// Use the access token for token information data
func (t *tokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {
	// TODO
	return nil, nil
}

// Use the refresh token for token information data
func (t *tokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	// TODO
	return nil, nil
}

var _ oauth2.ClientInfo = &clientInfo{}

// TODO: Postgres table
type clientInfo struct {
	id     string
	secret string
	domain string
	userId string
}

func (c *clientInfo) GetID() string {
	return c.id
}

func (c *clientInfo) GetSecret() string {
	return c.secret
}

func (c *clientInfo) GetDomain() string {
	return c.domain
}

func (c *clientInfo) GetUserID() string {
	return c.userId
}

var _ oauth2.ClientStore = &clientStore{}

type clientStore struct {
	d *database
}

func newClientStore(d *database) (t *clientStore, err error) {
	t = &clientStore{
		d: d,
	}
	return
}

// According to the ID for the client information
func (c *clientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	// TODO
	return nil, nil
}
