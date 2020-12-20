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

package oauth2

import (
	"net/url"
)

const (
	redirQueryKey      = "redir"
	redirQueryQueryKey = "q"
	loginErrorQueryKey = "login_error"
	authErrorQueryKey  = "auth_error"
)

func loginWithFirstPartyRedirPath(u *url.URL) string {
	var v url.Values
	v.Add(redirQueryKey, u.Path)
	v.Add(redirQueryQueryKey, u.RawQuery)
	n := url.URL{
		Path:     "/login",
		RawQuery: v.Encode(),
	}
	return n.String()
}

func isFirstPartyOAuth2Request(u *url.URL) bool {
	return u.Query().Get(redirQueryKey) != ""
}

func FirstPartyOAuth2LoginRedirPath(u *url.URL) (string, error) {
	v := u.Query()
	p, err := url.QueryUnescape(v.Get(redirQueryKey))
	if err != nil {
		return "", err
	}
	rq, err := url.QueryUnescape(v.Get(redirQueryQueryKey))
	if err != nil {
		return "", err
	}
	n := url.URL{
		Path:     p,
		RawQuery: rq,
	}
	return n.String(), nil
}

func AddLoginError(u *url.URL) *url.URL {
	return addKV(u, loginErrorQueryKey, "true")
}

func AddAuthError(u *url.URL) *url.URL {
	return addKV(u, authErrorQueryKey, "true")
}

func addKV(u *url.URL, key, value string) *url.URL {
	v := u.Query()
	v.Add(key, value)
	out := &url.URL{
		Path:     u.Path,
		RawQuery: v.Encode(),
	}
	return out
}
