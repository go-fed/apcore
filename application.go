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
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

// Application is an ActivityPub application built on top of apcore's
// infrastructure.
type Application interface {
	// CALLS MADE AT SERVER STARTUP
	//
	// These calls are made at least once, during server initialization, but
	// are not called when the server is handling requests.

	// Start is called at the beginning of a server's lifecycle, after
	// configuration processing and after the database connection is opened
	// but before web traffic is being served.
	//
	// If an error is returned, then the startup process fails.
	Start() error
	// Stop is called at the end of a server's lifecycle, after the web
	// servers have stopped serving traffic but before the database is
	// closed.
	//
	// If an error is returned, shutdown continues but an error is reported.
	Stop() error

	// Returns a pointer to the configuration struct used by the specific
	// application. It will be used to save and load from configuration
	// files. This object will be passed to SetConfiguration after it is
	// loaded from file.
	//
	// It is expected the Application will return an object with sane
	// defaults. The object's struct definition may have struct tags
	// supported by gopkg.in/ini.v1 for additional customization. For
	// example, the "comment" struct tag is much appreciated by admins.
	// Also, it is very important that keys to not collide, so prefix your
	// configuration options with a common prefix:
	//
	//     type MyAppConfig struct {
	//         SomeKey string `ini:"my_app_some_key" comment:"Description of this key"`
	//     }
	//
	// This configuration object is intended to be stable for the lifetime
	// of a running application. When the command to "serve" is given, this
	// function is only called once during application initialization.
	//
	// The command to "configure" will append these defaults to the guided
	// flow. Admins will then be able to inspect the file and modify the
	// configuration if desired.
	//
	// However, sane defaults for an application are incredibly important,
	// as the "new" command guides an admin through the creation process
	// all the way to serving without interruption. So have sane defaults!
	NewConfiguration() interface{}
	// Sets the configuration. The parameter's type is the same type that
	// is returned by NewConfiguration. Return an error if the configuration
	// is invalid.
	//
	// This configuration object is intended to be stable for the lifetime
	// of a running application. When the command to serve, is given, this
	// function is only called once during application initialization.
	SetConfiguration(interface{}) error

	// Whether this application supports ActivityPub's C2S protocol, or the
	// Social API.
	//
	// This and S2SEnabled may both be true. If C2SEnabled and S2SEnabled
	// both return false, an error will arise at startup.
	//
	// This is only checked at startup time. Attempting to enable or disable
	// C2S at runtime has no effect.
	C2SEnabled() bool
	// Whether this application supports ActivityPub's S2S protocol, or the
	// Federating API.
	//
	// This and C2SEnabled may both be true. If C2SEnabled and S2SEnabled
	// both return false, an error will arise at startup.
	//
	// This is only checked at startup time. Attempting to enable or disable
	// S2S at runtime has no effect.
	S2SEnabled() bool

	// The handler for the application's "404 Not Found" webpage.
	NotFoundHandler() http.Handler
	// The handler when a request makes an unsupported HTTP method against
	// a URI.
	MethodNotAllowedHandler() http.Handler
	// The handler for an internal server error.
	InternalServerErrorHandler() http.Handler
	// The handler for a bad request.
	BadRequestHandler() http.Handler
	// Web handler for a call to GET an actor's inbox. The framework applies
	// OAuth2 authorizations to fetch a public-only or private snapshot of
	// the inbox, and passes it into this handler function.
	//
	// The builtin ActivityPub handler will use the OAuth authorization.
	//
	// Only called if S2SEnabled is true.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetInboxWebHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage)
	// Web handler for a call to GET an actor's outbox. The framework
	// applies OAuth2 authorizations to fetch a public-only or private
	// snapshot of the outbox, and passes it to this handler function.
	//
	// The builtin ActivityPub handler will use the OAuth authorization.
	//
	// Always called regardless whether C2SEnabled or S2SEnabled is true.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetOutboxWebHandlerFunc() func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage)
	// Web handler for a call to GET an actor's followers collection. The
	// framework has no authorization requirements to view a user's
	// followers collection.
	//
	// Always called regardless whether C2SEnabled or S2SEnabled is true.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetFollowersWebHandlerFunc() http.HandlerFunc
	// Web handler for a call to GET an actor's following collection. The
	// framework has no authorization requirements to view a user's
	// following collection.
	//
	// Always called regardless whether C2SEnabled or S2SEnabled is true.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetFollowingWebHandlerFunc() http.HandlerFunc
	// Web handler for a call to GET an actor's liked collection. The
	// framework has no authorization requirements to view a user's
	// liked collection.
	//
	// Always called regardless whether C2SEnabled or S2SEnabled is true.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetLikedWebHandlerFunc() http.HandlerFunc
	// Web handler for a call to GET an actor. The framework has no
	// authorization requirements to view a user, like a profile.
	//
	// Always called regardless whether C2SEnabled or S2SEnabled is true.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetUserWebHandlerFunc() http.HandlerFunc

	// Builds the HTTP and ActivityPub routes specific for this application.
	//
	// The database is provided so custom handlers can access application
	// data directly, allowing clients to create the custom Fediverse
	// behavior their application desires.
	//
	// The Framework provided allows handlers to use common behaviors
	// provided by the apcore server framework.
	//
	// The bulk of typical HTTP application logic is in the handlers created
	// by the Router. The apcore.Router also supports creating routes that
	// process and serve ActivityStreams data, but the processing of the
	// ActivityPub data itself is handled elsewhere in
	// ApplyFederatingCallbacks and/or ApplySocialCallbacks.
	BuildRoutes(r *Router, db Database, f Framework) error

	// CALLS MADE AT SERVING TIME
	//
	// These calls are made when the server is handling requests, but are
	// not called during server initialization.

	// NewId creates a new id IRI for the content being created.
	//
	// A peer making a GET request to this IRI on this server should then
	// serve the ActivityPub value provided in this call.
	//
	// Ensure the route returned by NewId will be servable by a handler
	// created in the BuildRoutes call.
	NewId(c context.Context, t vocab.Type) (id *url.URL, err error)
	// ApplyFederatingCallbacks injects ActivityPub specific behaviors for
	// federated data.
	//
	// Additional behavior for out-of-the-box supported types, such as the
	// Create type, can be set by directly defining a function on the
	// callback passed in:
	//
	//     func (m *myImpl) ApplyFederatingCallbacks(fwc *pub.FederatingWrappedCallbacks) (others []interface{}) {
	//       fwc.Create = func(c context.Context, as vocab.ActivityStreamsCreate) error {
	//         // Additional application behavior for the Create activity.
	//       }
	//     }
	//
	// To use less common types that do no have out-of-the-box behavior,
	// such as the Listen type, return the functions in `others` that
	// implement the behavior. The functions in `others` must be in the
	// form:
	//
	//     func(c context.Context, as vocab.ActivityStreamsListen) error {
	//       // Application behavior for the Listen activity.
	//     }
	//
	// Caution: returning an out-of-the-box supported type in `others` will
	// override the framework-provided default behavior for that type. For
	// example, the "Create" behavior's default behavior of creating
	// ActivityStreams types in the database can be overridden by:
	//
	//     func (m *myImpl) ApplyFederatingCallbacks(fwc *pub.FederatingWrappedCallbacks) (others []interface{}) {
	//       others = []interface{}{
	//         func(c context.Context, as vocab.ActivityStreamsCreate) error {
	//           // New behavior for the Create activity that overrides the
	//           // framework provided defaults.
	//         },
	//       }
	//       return
	//     }
	//
	// Note: The `OnFollow` value will already be populated by the user's
	// preferred behavior upon receiving a Follow request.
	//
	// Only called if S2SEnabled returned true at startup time.
	ApplyFederatingCallbacks(fwc *pub.FederatingWrappedCallbacks) (others []interface{})
	// ApplySocialCallbacks injects ActivityPub specific behaviors for
	// social, or C2S, data.
	//
	// Additional behavior for out-of-the-box supported types, such as the
	// Create type, can be set by directly defining a function on the
	// callback passed in:
	//
	//     func (m *myImpl) ApplySocialCallbacks(swc *pub.SocialWrappedCallbacks) (others []interface{}) {
	//       swc.Create = func(c context.Context, as vocab.ActivityStreamsCreate) error {
	//         // Additional application behavior for the Create activity.
	//       }
	//     }
	//
	// To use less common types that do no have out-of-the-box behavior,
	// such as the Listen type, return the functions in `others` that
	// implement the behavior. The functions in `others` must be in the
	// form:
	//
	//     func(c context.Context, as vocab.ActivityStreamsListen) error {
	//       // Application behavior for the Listen activity.
	//     }
	//
	// Caution: returning an out-of-the-box supported type in `others` will
	// override the framework-provided default behavior for that type. For
	// example, the "Create" behavior's default behavior of creating
	// ActivityStreams types in the database can be overridden by:
	//
	//     func (m *myImpl) ApplySocialCallbacks(swc *pub.SocialWrappedCallbacks) (others []interface{}) {
	//       others = []interface{}{
	//         func(c context.Context, as vocab.ActivityStreamsCreate) error {
	//           // New behavior for the Create activity that overrides the
	//           // framework provided defaults.
	//         },
	//       }
	//       return
	//     }
	//
	// Only called if C2SEnabled returned true at startup time.
	ApplySocialCallbacks(swc *pub.SocialWrappedCallbacks) (others []interface{})

	// ScopePermitsPostOutbox determines if an OAuth token scope permits the
	// bearer to post to an actor's outbox. It is only called if C2S is
	// enabled.
	ScopePermitsPostOutbox(scope string) (permitted bool, err error)
	// ScopePermitsPrivateGetInbox determines if an OAuth token scope
	// permits the bearer to view private (non-Public) messages in an
	// actor's inbox. It is always called, regardless whether C2S or S2S is
	// enabled.
	ScopePermitsPrivateGetInbox(scope string) (permitted bool, err error)
	// ScopePermitsPrivateGetOutbox determines if an OAuth token scope
	// permits the bearer to view private (non-Public) messages in an
	// actor's outbox. It is always called, regardless whether C2S or S2S is
	// enabled.
	ScopePermitsPrivateGetOutbox(scope string) (permitted bool, err error)

	// CALLS MADE BOTH AT STARTUP AND SERVING TIME
	//
	// These calls are made at least once during server initialization, and
	// are called when the server is handling requests.

	// Information about this application's software. This will be shown at
	// the command line and used for NodeInfo statistics, as well as for
	// user agent information.
	Software() Software
}
