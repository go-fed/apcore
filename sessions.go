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
	"io/ioutil"

	gs "github.com/gorilla/sessions"
)

type sessions struct {
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
	s = &sessions{
		cookies: gs.NewCookieStore(keys...),
	}
	return
}
