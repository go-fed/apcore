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

package paths

import (
	"fmt"
	"net/url"
	"strings"
)

func Normalize(i *url.URL) *url.URL {
	c := *i
	c.RawQuery = ""
	c.Fragment = ""
	return &c
}

func NormalizeAsIRI(s string) (*url.URL, error) {
	c, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return Normalize(c), nil
}

type PathKey string

const (
	UserPathKey        PathKey = "users"
	InboxPathKey               = "inbox"
	InboxFirstPathKey          = "inboxFirst"
	InboxLastPathKey           = "inboxLast"
	OutboxPathKey              = "outbox"
	OutboxFirstPathKey         = "outboxFirst"
	OutboxLastPathKey          = "outboxLast"
	FollowersPathKey           = "followers"
	FollowingPathKey           = "following"
	LikedPathKey               = "liked"
	HttpSigPubKeyKey           = "httpsigPubKey"
)

var knownUserPaths map[PathKey]string = map[PathKey]string{
	UserPathKey:        "/users/{user}",
	InboxPathKey:       "/users/{user}/inbox",
	InboxFirstPathKey:  "/users/{user}/inbox",
	InboxLastPathKey:   "/users/{user}/inbox",
	OutboxPathKey:      "/users/{user}/outbox",
	OutboxFirstPathKey: "/users/{user}/outbox",
	OutboxLastPathKey:  "/users/{user}/outbox",
	FollowersPathKey:   "/users/{user}/followers",
	FollowingPathKey:   "/users/{user}/following",
	LikedPathKey:       "/users/{user}/liked",
	HttpSigPubKeyKey:   "/users/{user}/publicKeys/httpsig",
}

var knownUserPathQuery map[PathKey]string = map[PathKey]string{
	InboxFirstPathKey:  fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	InboxLastPathKey:   fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
	OutboxFirstPathKey: fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	OutboxLastPathKey:  fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
}

func UsernameFromUserPath(path string) (string, error) {
	s := strings.Split(path, "/")
	if len(s) < 3 {
		return "", fmt.Errorf("known user path does not contain username: %s", path)
	}
	return s[2], nil
}

func UserPathFor(k PathKey, username string) string {
	return strings.ReplaceAll(knownUserPaths[k], "{user}", username)
}

func userPathQueryFor(k PathKey) string {
	pq, ok := knownUserPathQuery[k]
	if !ok {
		return ""
	}
	return pq
}

func UserIRIFor(scheme string, host string, k PathKey, username string) *url.URL {
	u := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     UserPathFor(k, username),
		RawQuery: userPathQueryFor(k),
	}
	return u
}

func usernameFromActorID(actorID *url.URL) (string, error) {
	return UsernameFromUserPath(actorID.Path)
}

func IRIForActorID(k PathKey, actorID *url.URL) (*url.URL, error) {
	username, err := usernameFromActorID(actorID)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Scheme:   actorID.Scheme,
		Host:     actorID.Host,
		Path:     strings.ReplaceAll(knownUserPaths[k], "{user}", username),
		RawQuery: userPathQueryFor(k),
	}, nil

}
