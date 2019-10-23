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
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

const (
	userPreferencesContextKey    = "userPreferences"
	targetUserUUIDContextKey     = "targetUserUUID"
	activityIRIContextKey        = "activityIRI"
	activityTypeContextKey       = "activityType"
	completeRequestURLContextKey = "completeRequestURL"
	privateScopeContextKey       = "privateScope"
)

type ctx struct {
	context.Context
}

func newRequestContext(scheme, host string, r *http.Request, db *apdb, usernameFromPathFn func(string) (string, error)) (c ctx, err error) {
	pc := &ctx{r.Context()}
	var username string
	if username, err = usernameFromPathFn(r.URL.Path); err != nil {
		return
	}
	var userId string
	if userId, err = db.UserIdForUsername(c.Context, username); err != nil {
		return
	}
	pc.withTargetUserUUID(userId)
	var u userPreferences
	if u, err = db.UserPreferences(c.Context, userId); err != nil {
		return
	}
	pc.withUserPreferences(u)
	pc.withCompleteRequestURL(r, scheme, host)
	c = *pc
	return
}

func newRequestContextForBox(scheme, host string, r *http.Request, db *apdb) (c ctx, err error) {
	pc := &ctx{r.Context()}
	var userId string
	if userId, err = db.UserIdForBoxPath(c.Context, r.URL.Path); err != nil {
		return
	}
	pc.withTargetUserUUID(userId)
	var u userPreferences
	if u, err = db.UserPreferences(c.Context, userId); err != nil {
		return
	}
	pc.withUserPreferences(u)
	pc.withCompleteRequestURL(r, scheme, host)
	c = *pc
	return
}

func (c *ctx) withActivityStreamsValue(t vocab.Type) {
	if id, err := pub.GetId(t); err != nil {
		c.withActivityIRI(id)
	}
	c.withActivityType(t.GetTypeName())
}

func (c *ctx) withUserPreferences(u userPreferences) {
	c.Context = context.WithValue(c.Context, userPreferencesContextKey, u)
}

func (c *ctx) withTargetUserUUID(s string) {
	c.Context = context.WithValue(c.Context, targetUserUUIDContextKey, s)
}

func (c *ctx) withActivityIRI(u *url.URL) {
	c.Context = context.WithValue(c.Context, activityIRIContextKey, u)
}

func (c *ctx) withActivityType(s string) {
	c.Context = context.WithValue(c.Context, activityTypeContextKey, s)
}

func (c *ctx) withCompleteRequestURL(r *http.Request, scheme, host string) {
	u := *r.URL // Copy
	u.Host = host
	u.Scheme = scheme
	c.Context = context.WithValue(c.Context, completeRequestURLContextKey, &u)
}

func (c *ctx) SetPrivateScope(b bool) {
	c.Context = context.WithValue(c.Context, privateScopeContextKey, b)
}

func (c ctx) UserPreferences() (u userPreferences, err error) {
	v := c.Value(userPreferencesContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no user preferences in context")
	} else if u, ok = v.(userPreferences); !ok {
		err = fmt.Errorf("user preferences in context is not of type userPreferences")
	}
	return
}

func (c ctx) TargetUserUUID() (s string, err error) {
	v := c.Value(targetUserUUIDContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no target user UUID in context")
	} else if s, ok = v.(string); !ok {
		err = fmt.Errorf("target user UUID in context is not a string")
	}
	return
}

func (c ctx) ActivityIRI() (u *url.URL, err error) {
	v := c.Value(activityIRIContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no activity id in context")
	} else if u, ok = v.(*url.URL); !ok {
		err = fmt.Errorf("activity id in context is not a *url.URL")
	}
	return
}

func (c ctx) ActivityType() (s string, err error) {
	v := c.Value(activityTypeContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no activity type in context")
	} else if s, ok = v.(string); !ok {
		err = fmt.Errorf("activity type in context is not a string")
	}
	return
}

func (c ctx) CompleteRequestURL() (u *url.URL, err error) {
	v := c.Value(completeRequestURLContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no complete request URL in context")
	} else if u, ok = v.(*url.URL); !ok {
		err = fmt.Errorf("complete request URL in context is not a *url.URL")
	}
	return
}

func (c *ctx) HasPrivateScope() bool {
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
