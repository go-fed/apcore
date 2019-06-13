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
	"net/url"
)

const (
	// TODO: Populate the following items in the context.
	targetUserUUIDContextKey = "targetUserUUID"
	activityIdContextKey     = "activityId"
	activityTypeContextKey   = "activityType"
)

type ctx struct {
	context.Context
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

func (c ctx) ActivityId() (u *url.URL, err error) {
	v := c.Value(activityIdContextKey)
	var s string
	var ok bool
	if v == nil {
		err = fmt.Errorf("no activity id in context")
	} else if s, ok = v.(string); !ok {
		err = fmt.Errorf("activity id in context is not a string")
	}
	u, err = url.Parse(s)
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
