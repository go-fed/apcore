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

type Actor string

const (
	InstanceActor Actor = "instance"
)

var AllActors []Actor = []Actor{InstanceActor}

type PathKey string

const (
	UserPathKey           PathKey = "users"
	InboxPathKey                  = "inbox"
	InboxFirstPathKey             = "inboxFirst"
	InboxLastPathKey              = "inboxLast"
	OutboxPathKey                 = "outbox"
	OutboxFirstPathKey            = "outboxFirst"
	OutboxLastPathKey             = "outboxLast"
	FollowersPathKey              = "followers"
	FollowersFirstPathKey         = "followersFirst"
	FollowersLastPathKey          = "followersLast"
	FollowingPathKey              = "following"
	FollowingFirstPathKey         = "followingFirst"
	FollowingLastPathKey          = "followingLast"
	LikedPathKey                  = "liked"
	LikedFirstPathKey             = "likedFirst"
	LikedLastPathKey              = "likedLast"
	HttpSigPubKeyKey              = "httpsigPubKey"
)

var knownPaths map[PathKey]string = map[PathKey]string{
	UserPathKey:           "{user}",
	InboxPathKey:          "{user}/inbox",
	InboxFirstPathKey:     "{user}/inbox",
	InboxLastPathKey:      "{user}/inbox",
	OutboxPathKey:         "{user}/outbox",
	OutboxFirstPathKey:    "{user}/outbox",
	OutboxLastPathKey:     "{user}/outbox",
	FollowersPathKey:      "{user}/followers",
	FollowersFirstPathKey: "{user}/followers",
	FollowersLastPathKey:  "{user}/followers",
	FollowingPathKey:      "{user}/following",
	FollowingFirstPathKey: "{user}/following",
	FollowingLastPathKey:  "{user}/following",
	LikedPathKey:          "{user}/liked",
	LikedFirstPathKey:     "{user}/liked",
	LikedLastPathKey:      "{user}/liked",
	HttpSigPubKeyKey:      "{user}/publicKeys/httpsig",
}

func knownPath(prefix string, k PathKey) string {
	var b strings.Builder
	b.WriteRune('/')
	b.WriteString(prefix)
	b.WriteRune('/')
	b.WriteString(knownPaths[k])
	return b.String()
}

func knownUserPaths(k PathKey) string {
	return knownPath("users", k)
}

func knownActorsPaths(k PathKey) string {
	return knownPath("actors", k)
}

func ActorPathFor(k PathKey, c Actor) string {
	return strings.ReplaceAll(knownActorsPaths(k), "{user}", string(c))
}

func ActorIRIFor(scheme, host string, k PathKey, c Actor) *url.URL {
	u := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     ActorPathFor(k, c),
		RawQuery: uuidPathQueryFor(k),
	}
	return u
}

var knownUserPathQuery map[PathKey]string = map[PathKey]string{
	InboxFirstPathKey:     fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	InboxLastPathKey:      fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
	OutboxFirstPathKey:    fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	OutboxLastPathKey:     fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
	FollowersFirstPathKey: fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	FollowersLastPathKey:  fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
	FollowingFirstPathKey: fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	FollowingLastPathKey:  fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
	LikedFirstPathKey:     fmt.Sprintf("%s=%s", queryCollectionPage, queryTrue),
	LikedLastPathKey:      fmt.Sprintf("%s=%s&%s=%s", queryCollectionPage, queryTrue, queryCollectionEnd, queryTrue),
}

type UUID string

func UUIDFromUserPath(path string) (UUID, error) {
	s := strings.Split(path, "/")
	if len(s) < 3 {
		return UUID(""), fmt.Errorf("known user path does not contain uuid: %s", path)
	}
	return UUID(s[2]), nil
}

func UUIDPathFor(k PathKey, uuid UUID) string {
	return strings.ReplaceAll(knownUserPaths(k), "{user}", string(uuid))
}

func uuidPathQueryFor(k PathKey) string {
	pq, ok := knownUserPathQuery[k]
	if !ok {
		return ""
	}
	return pq
}

func UUIDIRIFor(scheme string, host string, k PathKey, uuid UUID) *url.URL {
	u := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     UUIDPathFor(k, uuid),
		RawQuery: uuidPathQueryFor(k),
	}
	return u
}

func uuidFromActorID(actorID *url.URL) (UUID, error) {
	return UUIDFromUserPath(actorID.Path)
}

func IRIForActorID(k PathKey, actorID *url.URL) (*url.URL, error) {
	uuid, err := uuidFromActorID(actorID)
	if err != nil {
		return nil, err
	}
	// Must first check special cases
	var pFn func(k PathKey) string = knownUserPaths
	if string(uuid) == string(InstanceActor) {
		pFn = knownActorsPaths
	}
	return &url.URL{
		Scheme:   actorID.Scheme,
		Host:     actorID.Host,
		Path:     strings.ReplaceAll(pFn(k), "{user}", string(uuid)),
		RawQuery: uuidPathQueryFor(k),
	}, nil
}

func Route(k PathKey) string {
	return knownUserPaths(k)
}

func IsUserPath(id *url.URL) bool {
	s := strings.Split(id.Path, "/")
	return len(s) == 3 &&
		(strings.Contains(id.Path, "users") || strings.Contains(id.Path, "actors"))
}

func IsFollowersPath(id *url.URL) bool {
	return isSubPath(id, "followers")
}

func IsFollowingPath(id *url.URL) bool {
	return isSubPath(id, "following")
}

func IsLikedPath(id *url.URL) bool {
	return isSubPath(id, "liked")
}

func isSubPath(id *url.URL, sub string) bool {
	s := strings.Split(id.Path, "/")
	return len(s) > 3 &&
		(strings.Contains(id.Path, "users") || strings.Contains(id.Path, "actors")) &&
		strings.Contains(s[3], sub)
}
