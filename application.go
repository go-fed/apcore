package apcore

import (
	"net/http"
)

// Application is an ActivityPub application built on top of apcore's
// infrastructure.
type Application interface {
	// Information about this application's software. This will be shown at
	// the command line and used for NodeInfo.
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
	// The handler for the application's "404 Not Found" webpage.
	NotFoundHandler() http.Handler
	// The handler when a request makes an unsupported HTTP method against
	// a URI.
	MethodNotAllowedHandler() http.Handler
	// Builds the HTTP and ActivityPub routes specific for this application.
	//
	// The bulk of the application logic goes here.
	BuildRoutes(r *Router) error
}
