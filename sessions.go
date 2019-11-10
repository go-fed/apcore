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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	gs "github.com/gorilla/sessions"
)

type sessions struct {
	name    string
	cookies *gs.CookieStore
}

func newSessions(c *config) (s *sessions, err error) {
	var authKey, encKey []byte
	var keys [][]byte
	authKey, err = ioutil.ReadFile(c.ServerConfig.CookieAuthKeyFile)
	if err != nil {
		return
	}
	if len(c.ServerConfig.CookieEncryptionKeyFile) > 0 {
		InfoLogger.Info("Cookie encryption key file detected")
		encKey, err = ioutil.ReadFile(c.ServerConfig.CookieEncryptionKeyFile)
		if err != nil {
			return
		}
		keys = [][]byte{authKey, encKey}
	} else {
		InfoLogger.Info("No cookie encryption key file detected")
		keys = [][]byte{authKey}
	}
	if len(c.ServerConfig.CookieSessionName) <= 0 {
		err = fmt.Errorf("no cookie session name provided")
		return
	}
	s = &sessions{
		name:    c.ServerConfig.CookieSessionName,
		cookies: gs.NewCookieStore(keys...),
	}
	opt := &gs.Options{
		Path:     "/",
		Domain:   c.ServerConfig.Host,
		MaxAge:   c.ServerConfig.CookieMaxAge,
		Secure:   true,
		HttpOnly: true,
	}
	s.cookies.Options = opt
	s.cookies.MaxAge(opt.MaxAge)
	return
}

func (s *sessions) Get(r *http.Request) (ses *session, err error) {
	var gs *gs.Session
	gs, err = s.cookies.Get(r, s.name)
	ses = &session{
		gs: gs,
	}
	return
}

type session struct {
	gs *gs.Session
}

const (
	userIDSessionKey           = "userid"
	oAuthRedirectFormValuesKey = "oauth_redir"
)

func (s *session) SetUserID(uuid string) {
	s.gs.Values[userIDSessionKey] = uuid
	return
}

func (s *session) UserID() (uuid string, err error) {
	if v, ok := s.gs.Values[userIDSessionKey]; !ok {
		err = fmt.Errorf("no user id in session")
		return
	} else if uuid, ok = v.(string); !ok {
		err = fmt.Errorf("user id in session is not a string")
		return
	}
	return
}

func (s *session) SetOAuthRedirectFormValues(f url.Values) {
	s.gs.Values[oAuthRedirectFormValuesKey] = f
	return
}

func (s *session) OAuthRedirectFormValues() (v url.Values, ok bool) {
	var i interface{}
	if i, ok = s.gs.Values[oAuthRedirectFormValuesKey]; !ok {
		return
	}
	v, ok = i.(url.Values)
	return
}

func (s *session) DeleteOAuthRedirectFormValues() {
	delete(s.gs.Values, oAuthRedirectFormValuesKey)
}

func (s *session) Save(r *http.Request, w http.ResponseWriter) error {
	return s.gs.Save(r, w)
}
