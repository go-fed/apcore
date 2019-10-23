package main

import (
	"github.com/go-fed/apcore"
)

func main() {
	// Build an instance of our struct that satisfies the Application
	// interface.
	//
	// Implementing the Application interface is where most of the work
	// to use the framework lies.
	//
	// go-fed/apcore provides a very quick-to-implement but vanilla
	// ActivityPub framework. It is a convenience layer on top of
	// go-fed/activity, which has the opposite philosophy: assume as little
	// as possible, provide more powerful but time-consuming interfaces to
	// satisfy.
	var a apcore.Application = &App{}
	// Run the apcore framework.
	//
	// Depending on the command line flags chosen, an action can occur:
	// - Configuring the application and generating a config file
	// - Initializing a database
	// - Initializing an admin account in the database
	// - Launching the example App to serve live web & ActivityPub traffic
	//
	// All of these capabilities are supported by the framework out of the
	// box. Refer to the command line help for more details.
	apcore.Run(a)
}
