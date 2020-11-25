// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
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
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/util"
)

// Run will launch the apcore server.
func Run(a app.Application) {
	if !flag.Parsed() {
		flag.Parse()
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Check and prepare debug mode
	if *devFlag {
		util.InfoLogger.Info("Debug mode enabled")
		if len(*infoLogFileFlag) > 0 {
			util.InfoLogger.Warning("info_log_file flag ignored in debug mode")
		}
		if len(*errorLogFileFlag) > 0 {
			util.InfoLogger.Warning("error_log_file flag ignored in debug mode")
		}
	} else {
		// Prepare production logging
		var il, el io.Writer = os.Stdout, os.Stderr
		var err error
		if len(*infoLogFileFlag) > 0 {
			il, err = os.OpenFile(
				*infoLogFileFlag,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0660)
			if err != nil {
				util.ErrorLogger.Errorf("cannot open %s: %s", *infoLogFileFlag, err)
				os.Exit(1)
			}
		}
		if len(*errorLogFileFlag) > 0 {
			el, err = os.OpenFile(
				*errorLogFileFlag,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0660)
			if err != nil {
				util.ErrorLogger.Errorf("cannot open %s: %s", *infoLogFileFlag, err)
				os.Exit(1)
			}
		}
		util.LogInfoTo(*systemLogFlag, il)
		defer util.LogInfoToStdout()
		util.LogErrorTo(*systemLogFlag, el)
		defer util.LogErrorToStderr()
	}

	// Conduct the action
	var action cmdAction
	found := false
	for _, v := range allActions {
		if v.Name == flag.Arg(0) {
			action = v
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", flag.Arg(0))
		fmt.Fprintf(os.Stderr, "Available actions:\n%s", allActionsUsage())
		os.Exit(1)
	} else if err := action.Action(a); err != nil {
		util.ErrorLogger.Errorf("error running %s: %s", flag.Arg(0), err)
		os.Exit(1)
	}
}
