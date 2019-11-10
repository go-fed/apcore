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
	var a apcore.Application
	var e error
	a, e = newApplication(
		/*tmpls=*/ []string{
			"templates/nav.html",
			"templates/inline_css.html",
			"templates/footer.html",
			"templates/header.html",
			"templates/home.html",
			"templates/login.html",
			"templates/users.html",
		},
	)
	if e != nil {
		panic(e)
	}
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
