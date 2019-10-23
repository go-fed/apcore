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
	"net/url"
	"strings"
)

func normalize(i *url.URL) *url.URL {
	c := *i
	c.RawQuery = ""
	c.Fragment = ""
	return &c
}

func normalizeAsIRI(s string) (*url.URL, error) {
	c, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return normalize(c), nil
}

type pathKey string

const (
	userPathKey      pathKey = "users"
	inboxPathKey             = "inbox"
	outboxPathKey            = "outbox"
	followersPathKey         = "followers"
	followingPathKey         = "following"
	likedPathKey             = "liked"
	pubKeyKey                = "pubKey"
)

var knownUserPaths map[pathKey]string = map[pathKey]string{
	userPathKey:      "/users/{user}",
	inboxPathKey:     "/users/{user}/inbox",
	outboxPathKey:    "/users/{user}/outbox",
	followersPathKey: "/users/{user}/followers",
	followingPathKey: "/users/{user}/following",
	likedPathKey:     "/users/{user}/liked",
	pubKeyKey:        "/users/{user}/publicKeys/1",
}

func usernameFromKnownUserPath(path string) (string, error) {
	s := strings.Split(path, "/")
	if len(s) < 3 {
		return "", fmt.Errorf("known user path does not contain username: %s", path)
	}
	return s[2], nil
}

func knownUserPathFor(k pathKey, username string) string {
	return strings.ReplaceAll(knownUserPaths[k], "{user}", username)
}

func knownUserIRIFor(scheme string, host string, k pathKey, username string) *url.URL {
	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   knownUserPathFor(k, username),
	}
	return u
}
