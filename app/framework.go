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
	"github.com/go-fed/oauth2"
)

// Framework provides request-time hooks for use in handlers.
type Framework interface {
	// ValidateOAuth2AccessToken attempts to obtain and validate the OAuth
	// token presented in the request. This can be called in your handlers
	// at request-handing time.
	//
	// If an error is returned, both the token and authentication values
	// should be ignored.
	//
	// It is possible a token is returned when the user has not passed
	// authentication.
	//
	// You may use the token to further enforce the scope, depending on the
	// application and handler's use case.
	ValidateOAuth2AccessToken(w http.ResponseWriter, r *http.Request) (token oauth2.TokenInfo, authenticated bool, err error)

	// Send will send an Activity or Object on behalf of the user
	// represented by the outbox IRI.
	//
	// Note that a new ID is not needed on the activity and/or objects that
	// are being sent; they will be generated as needed.
	Send(c context.Context, outbox *url.URL, toSend vocab.Type) error
}
