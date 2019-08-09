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
	"net/url"

	"github.com/go-fed/activity/pub"
)

var _ pub.Transport = &transport{}

// TODO: Implement
type transport struct {
}

func newTransport() (t *transport, err error) {
	return &transport{}, nil
}

func (t *transport) Dereference(c context.Context, iri *url.URL) (b []byte, err error) {
	return
}

func (t *transport) Deliver(c context.Context, b []byte, to *url.URL) (err error) {
	return
}

func (t *transport) BatchDeliver(c context.Context, b []byte, recipients []*url.URL) (err error) {
	return
}
