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
	"net/http"
)

// Application is an ActivityPub application built on top of apcore's
// infrastructure.
type Application interface {
	// Information about this application's software. This will be shown at
	// the command line and used for NodeInfo statistics.
	Software() Software

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

	// Builds the HTTP and ActivityPub routes specific for this application.
	//
	// The database is provided so custom handlers can access application
	// data directly, allowing clients to create the custom Fediverse
	// behavior their application desires.
	//
	// The bulk of the application logic is in the handlers created by the
	// Router.
	BuildRoutes(r *Router, db Database) error
}
