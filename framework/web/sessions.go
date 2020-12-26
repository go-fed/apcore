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

package web

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/util"
	gs "github.com/gorilla/sessions"
)

type Sessions struct {
	name    string
	cookies *gs.CookieStore
}

func NewSessions(c *config.Config, scheme string) (s *Sessions, err error) {
	var authKey, encKey []byte
	var keys [][]byte
	authKey, err = ioutil.ReadFile(c.ServerConfig.CookieAuthKeyFile)
	if err != nil {
		return
	}
	if len(c.ServerConfig.CookieEncryptionKeyFile) > 0 {
		util.InfoLogger.Info("Cookie encryption key file detected")
		encKey, err = ioutil.ReadFile(c.ServerConfig.CookieEncryptionKeyFile)
		if err != nil {
			return
		}
		keys = [][]byte{authKey, encKey}
	} else {
		util.InfoLogger.Info("No cookie encryption key file detected")
		keys = [][]byte{authKey}
	}
	if len(c.ServerConfig.CookieSessionName) <= 0 {
		err = fmt.Errorf("no cookie session name provided")
		return
	}
	s = &Sessions{
		name:    c.ServerConfig.CookieSessionName,
		cookies: gs.NewCookieStore(keys...),
	}
	opt := &gs.Options{
		Path:     "/",
		Domain:   c.ServerConfig.Host,
		MaxAge:   c.ServerConfig.CookieMaxAge,
		Secure:   scheme != "http",
		HttpOnly: true,
	}
	s.cookies.Options = opt
	s.cookies.MaxAge(opt.MaxAge)
	return
}

func (s *Sessions) Get(r *http.Request) (ses *Session, err error) {
	var gs *gs.Session
	gs, err = s.cookies.Get(r, s.name)
	ses = &Session{
		gs: gs,
	}
	return
}

type Session struct {
	gs *gs.Session
}

const (
	userIDSessionKey = "userid"
)

func (s *Session) SetUserID(uuid string) {
	s.gs.Values[userIDSessionKey] = uuid
	return
}

func (s *Session) UserID() (uuid string, err error) {
	if v, ok := s.gs.Values[userIDSessionKey]; !ok {
		err = fmt.Errorf("no user id in session")
		return
	} else if uuid, ok = v.(string); !ok {
		err = fmt.Errorf("user id in session is not a string")
		return
	}
	return
}

func (s *Session) DeleteUserID() {
	delete(s.gs.Values, userIDSessionKey)
}

const (
	firstPartyCredentialKey = "fpckey"
)

func (s *Session) SetFirstPartyCredentialID(id string) {
	s.gs.Values[firstPartyCredentialKey] = id
	return
}

func (s *Session) HasFirstPartyCredentialID() bool {
	_, ok := s.gs.Values[firstPartyCredentialKey]
	return ok
}

func (s *Session) FirstPartyCredentialID() (id string, err error) {
	if v, ok := s.gs.Values[firstPartyCredentialKey]; !ok {
		err = fmt.Errorf("no first party credential in session")
		return
	} else if id, ok = v.(string); !ok {
		err = fmt.Errorf("first party credential in session is not a string")
		return
	}
	return
}

func (s *Session) DeleteFirstPartyCredentialID() {
	delete(s.gs.Values, firstPartyCredentialKey)
}

func (s *Session) Clear() {
	s.DeleteUserID()
	s.DeleteFirstPartyCredentialID()
}

func (s *Session) Save(r *http.Request, w http.ResponseWriter) error {
	return s.gs.Save(r, w)
}
