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
	// supported by gopkg.in/ini.v1 for additional customization.
	//
	// This configuration object is intended to be stable for the lifetime
	// of a running application. When the command to serve, is given, this
	// function is only called once during application initialization.
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
	// TODO: Function to construct HTTP routes for the application.
}
