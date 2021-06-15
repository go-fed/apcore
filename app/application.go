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

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

// Application is an ActivityPub application built on top of apcore's
// infrastructure. Your application must also implement C2SApplication,
// S2SApplication, or both interfaces in order to gain the benefits of
// federating using ActivityPub's Social or Federating Protocols.
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
	// Provides a read-only interface for some of APCore's config fields.
	//
	// This configuration object is intended to be stable for the lifetime
	// of a running application. When the command to serve, is given, this
	// function is only called once during application initialization.
	SetConfiguration(interface{}, APCoreConfig) error

	// The handler for the application's "404 Not Found" webpage.
	NotFoundHandler(Framework) http.Handler
	// The handler when a request makes an unsupported HTTP method against
	// a URI.
	MethodNotAllowedHandler(Framework) http.Handler
	// The handler for an internal server error.
	InternalServerErrorHandler(Framework) http.Handler
	// The handler for a bad request.
	BadRequestHandler(Framework) http.Handler

	// Web handlers for the application server

	// Web handler for a GET call to the login page.
	//
	// It should render a login page that POSTs to the "/login" endpoint.
	//
	// If the URL contains a query parameter "login_error" with a value of
	// "true", then it should convey to the user that the email or password
	// previously entered was incorrect.
	GetLoginWebHandlerFunc(Framework) http.HandlerFunc
	// Web handler for a GET call to the OAuth2 authorization page.
	//
	// It should render UX that informs the user that the other application
	// is requesting to be authorized as that user to obtain certain scopes.
	//
	// See the OAuth2 RFC 6749 for more information.
	GetAuthWebHandlerFunc(Framework) http.HandlerFunc

	// Web handlers for ActivityPub related data

	// Web handler for a call to GET an actor's outbox. The framework
	// applies OAuth2 authorizations to fetch a public-only or private
	// snapshot of the outbox, and passes it to this handler function.
	//
	// The builtin ActivityPub handler will use the OAuth authorization.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetOutboxWebHandlerFunc(Framework) func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage)
	// Web handler for a call to GET an actor's followers collection. The
	// framework has no authorization requirements to view a user's
	// followers collection.
	//
	// Also returns for the corresponding AuthorizeFunc handler, which will
	// be applied to both ActivityPub and web requests.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served. Returning a nil AuthorizeFunc
	// results in public access.
	GetFollowersWebHandlerFunc(Framework) (CollectionPageHandlerFunc, AuthorizeFunc)
	// Web handler for a call to GET an actor's following collection. The
	// framework has no authorization requirements to view a user's
	// following collection.
	//
	// Also returns for the corresponding AuthorizeFunc handler, which will
	// be applied to both ActivityPub and web requests.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served. Returning a nil AuthorizeFunc
	// results in public access.
	GetFollowingWebHandlerFunc(Framework) (CollectionPageHandlerFunc, AuthorizeFunc)
	// Web handler for a call to GET an actor's liked collection. The
	// framework has no authorization requirements to view a user's
	// liked collection.
	//
	// Also returns for the corresponding AuthorizeFunc handler, which will
	// be applied to both ActivityPub and web requests.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served. Returning a nil AuthorizeFunc
	// results in public access.
	GetLikedWebHandlerFunc(Framework) (CollectionPageHandlerFunc, AuthorizeFunc)
	// Web handler for a call to GET an actor. The framework has no
	// authorization requirements to view a user, like a profile.
	//
	// Also returns for the corresponding AuthorizeFunc handler, which will
	// be applied to both ActivityPub and web requests.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served. Returning a nil AuthorizeFunc
	// results in public access.
	GetUserWebHandlerFunc(Framework) (VocabHandlerFunc, AuthorizeFunc)

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
	BuildRoutes(r Router, db Database, f Framework) error

	// CALLS MADE AT SERVING TIME
	//
	// These calls are made when the server is handling requests, but are
	// not called during server initialization.

	// NewIDPath creates a new id IRI path component for the content being
	// created.
	//
	// A peer making a GET request to this path on this server should then
	// serve the ActivityPub value provided in this call. For example:
	//   "/notes/abcd0123-4567-890a-bcd0-1234567890ab"
	//
	// Ensure the route returned by NewIDPath will be servable by a handler
	// created in the BuildRoutes call.
	NewIDPath(c context.Context, t vocab.Type) (path string, err error)

	// ScopePermitsPrivateGetInbox determines if an OAuth token scope
	// permits the bearer to view private (non-Public) messages in an
	// actor's inbox.
	ScopePermitsPrivateGetInbox(scope string) (permitted bool, err error)
	// ScopePermitsPrivateGetOutbox determines if an OAuth token scope
	// permits the bearer to view private (non-Public) messages in an
	// actor's outbox.
	ScopePermitsPrivateGetOutbox(scope string) (permitted bool, err error)

	// DefaultUserPreferences returns an application-specific preferences
	// struct to be serialized into JSON and used as initial user app
	// preferences.
	DefaultUserPreferences() interface{}
	// DefaultUserPrivileges returns an application-specific privileges
	// struct to be serialized into JSON and used as initial user app
	// privileges.
	DefaultUserPrivileges() interface{}
	// DefaultAdminPrivileges returns an application-specific privileges
	// struct to be serialized into JSON and used as initial user app
	// privileges for new admins.
	DefaultAdminPrivileges() interface{}

	// CALLS MADE BOTH AT STARTUP AND SERVING TIME
	//
	// These calls are made at least once during server initialization, and
	// are called when the server is handling requests.

	// Information about this application's software. This will be shown at
	// the command line and used for NodeInfo statistics, as well as for
	// user agent information.
	Software() Software
}

// C2SApplication is an Application with additional methods required to support
// the C2S, or Social, ActivityPub protocol.
type C2SApplication interface {
	// ScopePermitsPostOutbox determines if an OAuth token scope permits the
	// bearer to post to an actor's outbox.
	ScopePermitsPostOutbox(scope string) (permitted bool, err error)

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
	ApplySocialCallbacks(swc *pub.SocialWrappedCallbacks) (others []interface{})
}

// S2SApplication is an Application with the additional methods required to
// support the S2S, or Federating, ActivityPub protocol.
type S2SApplication interface {
	// Web handler for a call to GET an actor's inbox. The framework applies
	// OAuth2 authorizations to fetch a public-only or private snapshot of
	// the inbox, and passes it into this handler function.
	//
	// The builtin ActivityPub handler will use the OAuth authorization.
	//
	// Returning a nil handler is allowed, and doing so results in only
	// ActivityStreams content being served.
	GetInboxWebHandlerFunc(Framework) func(w http.ResponseWriter, r *http.Request, outbox vocab.ActivityStreamsOrderedCollectionPage)

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
	ApplyFederatingCallbacks(fwc *pub.FederatingWrappedCallbacks) (others []interface{})
}

// APCoreConfig allows the application to reuse common fields set in apcore's config.
type APCoreConfig interface {
	// Hostname of the application set in the config
	Host() string
	// Clock timezone set in the config
	ClockTimezone() string
}
