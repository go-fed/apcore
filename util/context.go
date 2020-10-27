package util

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

const (
	activityIRIContextKey  = "activityIRI"
	activityTypeContextKey = "activityType"
	userPathUUIDContextKey = "userPathUUID"
)

type Context struct {
	context.Context
}

func (c *Context) WithActivityStreamsValue(t vocab.Type) {
	if id, err := pub.GetId(t); err != nil {
		c.WithActivityIRI(id)
	}
	c.WithActivityType(t.GetTypeName())
}

func (c *Context) WithActivityIRI(u *url.URL) {
	c.Context = context.WithValue(c.Context, activityIRIContextKey, u)
}

func (c *Context) WithActivityType(s string) {
	c.Context = context.WithValue(c.Context, activityTypeContextKey, s)
}

func (c *Context) WithUserPathUUID(s string) {
	c.Context = context.WithValue(c.Context, userPathUUIDContextKey, s)
}

func (c Context) ActivityIRI() (u *url.URL, err error) {
	v := c.Value(activityIRIContextKey)
	var ok bool
	if v == nil {
		err = errors.New("no activity id in context")
	} else if u, ok = v.(*url.URL); !ok {
		err = errors.New("activity id in context is not a *url.URL")
	}
	return
}

func (c Context) ActivityType() (s string, err error) {
	return c.toStringValue("activity type", activityTypeContextKey)
}

func (c Context) UserPathUUID() (s string, err error) {
	return c.toStringValue("user path UUID", userPathUUIDContextKey)
}

func (c Context) toStringValue(name, key string) (s string, err error) {
	v := c.Value(key)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no %s in context", name)
	} else if s, ok = v.(string); !ok {
		err = fmt.Errorf("%s in context is not a string", name)
	}
	return
}
