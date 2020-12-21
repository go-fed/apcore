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

package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/paths"
)

const (
	activityContextKey           = "activity"
	activityStreamContextKey     = "activityStream"
	userPathUUIDContextKey       = "userPathUUID"
	actorIRIContextKey           = "actorIRI"
	completeRequestURLContextKey = "completeRequestURL"
	privateScopeContextKey       = "privateScope"
)

type Context struct {
	context.Context
}

// WithUserAPHTTPContext sets the UserPathUUID, ActorIRI, CompleteRequestURL,
// and PrivateScope.
func WithUserAPHTTPContext(scheme, host string, r *http.Request, uuid paths.UUID, authdUserID string) Context {
	c := &Context{r.Context()}
	c.WithUserPathUUID(uuid)
	c.WithActorIRI(paths.UUIDIRIFor(scheme, host, paths.UserPathKey, uuid))
	c.WithCompleteRequestURL(r, scheme, host)
	c.WithPrivateScope(len(authdUserID) > 0 && authdUserID == string(uuid))
	return *c
}

// WithAPHTTPContext sets the CompleteRequestURL.
func WithAPHTTPContext(scheme, host string, r *http.Request) Context {
	c := &Context{r.Context()}
	c.WithCompleteRequestURL(r, scheme, host)
	return *c
}

// WithActivity is used for federating contexts.
func (c *Context) WithActivity(t pub.Activity) {
	c.Context = context.WithValue(c.Context, activityContextKey, t)
}

// WithActivityStream is used for social contexts.
func (c *Context) WithActivityStream(t vocab.Type) {
	c.Context = context.WithValue(c.Context, activityStreamContextKey, t)
}

// WithUserID is used for ActivityPub Inbox/Outbox contexts.
func (c *Context) WithUserPathUUID(uuid paths.UUID) {
	c.Context = context.WithValue(c.Context, userPathUUIDContextKey, uuid)
}

// WithActorIRI is used for ActivityPub Inbox/Outbox contexts.
func (c *Context) WithActorIRI(id *url.URL) {
	c.Context = context.WithValue(c.Context, actorIRIContextKey, id)
}

// WithCompleteRequestURL is used for all ActivityPub HTTP contexts.
func (c *Context) WithCompleteRequestURL(r *http.Request, scheme, host string) {
	u := *r.URL // Copy
	u.Host = host
	u.Scheme = scheme
	c.Context = context.WithValue(c.Context, completeRequestURLContextKey, &u)
}

// WithPrivateScope is available in all GET http requests.
func (c *Context) WithPrivateScope(b bool) {
	c.Context = context.WithValue(c.Context, privateScopeContextKey, b)
}

// Activity is available in federating contexts.
func (c Context) Activity() (t pub.Activity, err error) {
	v := c.Value(activityContextKey)
	var ok bool
	if v == nil {
		err = errors.New("no activity in context")
	} else if t, ok = v.(pub.Activity); !ok {
		err = errors.New("activity in context is not a pub.Activity")
	}
	return
}

// ActivityStream is available in social contexts.
func (c Context) ActivityStream() (t vocab.Type, err error) {
	v := c.Value(activityStreamContextKey)
	var ok bool
	if v == nil {
		err = errors.New("no activity stream in context")
	} else if t, ok = v.(vocab.Type); !ok {
		err = errors.New("activity stream in context is not a vocab.Type")
	}
	return
}

// UserPathUUID is used for ActivityPub HTTP contexts.
func (c Context) UserPathUUID() (s paths.UUID, err error) {
	return c.toUUIDValue("user path UUID", userPathUUIDContextKey)
}

// ActorIRI is used for ActivityPub HTTP contexts.
func (c Context) ActorIRI() (s *url.URL, err error) {
	return c.toURLValue("actor IRI", actorIRIContextKey)
}

// CompleteRequestURL is used for ActivityPub HTTP contexts.
func (c Context) CompleteRequestURL() (u *url.URL, err error) {
	return c.toURLValue("complete Request URL", completeRequestURLContextKey)
}

// HasPrivateScope is available in all GET http requests.
func (c *Context) HasPrivateScope() bool {
	v := c.Value(privateScopeContextKey)
	var b, ok bool
	if v == nil {
		return false
	} else if b, ok = v.(bool); !ok {
		return false
	} else {
		return b
	}
}

func (c Context) toUUIDValue(name, key string) (s paths.UUID, err error) {
	v := c.Value(key)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no %s in context", name)
	} else if s, ok = v.(paths.UUID); !ok {
		err = fmt.Errorf("%s in context is not a paths.UUID", name)
	}
	return
}

func (c Context) toURLValue(name, key string) (u *url.URL, err error) {
	v := c.Value(key)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no %s in context", name)
	} else if u, ok = v.(*url.URL); !ok {
		err = fmt.Errorf("%s in context is not a *url.URL", name)
	}
	return
}
