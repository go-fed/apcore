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

package app

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/paths"
)

// Framework provides request-time hooks for use in handlers.
type Framework interface {
	Context(r *http.Request) context.Context

	UserIRI(userUUID paths.UUID) *url.URL

	// CreateUser creates a new unprivileged user with the given username,
	// email, and password.
	//
	// If an error is returned, it can be checked using IsNotUniqueUsername
	// and IsNotUniqueEmail to show the error to the user.
	CreateUser(c context.Context, username, email, password string) (userID string, err error)

	// IsNotUniqueUsername returns true if the error returned from
	// CreateUser is due to the username not being unique.
	IsNotUniqueUsername(error) bool

	// IsNotUniqueEmail returns true if the error returned from CreateUser
	// is due to the email not being unique.
	IsNotUniqueEmail(error) bool

	// Validate attempts to obtain and validate the OAuth token or first
	// party credential in the request. This can be called in your handlers
	// at request-handing time.
	//
	// If an error is returned, both the token and authentication values
	// should be ignored.
	//
	// TODO: Scopes
	Validate(w http.ResponseWriter, r *http.Request) (userID paths.UUID, authenticated bool, err error)

	// Send will send an Activity or Object on behalf of the user.
	//
	// Note that a new ID is not needed on the activity and/or objects that
	// are being sent; they will be generated as needed.
	//
	// Calling Send when federation is disabled results in an error.
	Send(c context.Context, userID paths.UUID, toSend vocab.Type) error

	// SendAcceptFollow accepts the provided Follow on behalf of the user.
	//
	// Calling SendAcceptFollow when federation is disabled results in an
	// error.
	SendAcceptFollow(c context.Context, userID paths.UUID, followIRI *url.URL) error

	// SendRejectFollow rejects the provided Follow on behalf of the user.
	//
	// Calling SendRejectFollow when federation is disabled results in an
	// error.
	SendRejectFollow(c context.Context, userID paths.UUID, followIRI *url.URL) error

	Session(r *http.Request) (Session, error)

	// TODO: Determine if we need this.
	GetByIRI(c context.Context, id *url.URL) (vocab.Type, error)

	// Given a user ID, retrieves all follow requests that have not yet been
	// Accepted nor Rejected.
	OpenFollowRequests(c context.Context, userID paths.UUID) ([]vocab.ActivityStreamsFollow, error)

	// GetPrivileges accepts a pointer to an appPrivileges struct to read
	// from the database for the given user, and also returns whether that
	// user is an admin.
	GetPrivileges(c context.Context, userID paths.UUID, appPrivileges interface{}) (admin bool, err error)
	// SetPrivileges sets the given application privileges and admin status
	// for the given user.
	SetPrivileges(c context.Context, userID paths.UUID, admin bool, appPrivileges interface{}) error
}

type Session interface {
	UserID() (string, error)
	Set(string, interface{})
	Get(string) (interface{}, bool)
	Has(string) bool
	Delete(string)
	Save(*http.Request, http.ResponseWriter) error
}
