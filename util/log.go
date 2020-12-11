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

package util

import (
	"io"
	"os"

	"github.com/google/logger"
)

var (
	// These loggers will only respect the logging flags while the call to
	// Run is executing. Otherwise, they log to os.Stdout and os.Stderr.
	InfoLogger  *logger.Logger = logger.Init("apcore", false, false, os.Stdout)
	infoClose   bool           = false
	ErrorLogger *logger.Logger = logger.Init("apcore", false, false, os.Stderr)
	errorClose  bool           = false
)

func LogInfoTo(system bool, w io.Writer) {
	maybeCloseAndLogTo(&InfoLogger, system, w, &infoClose)
}

func LogErrorTo(system bool, w io.Writer) {
	maybeCloseAndLogTo(&ErrorLogger, system, w, &errorClose)
}

func LogInfoToStdout() {
	maybeCloseAndLogTo(&InfoLogger, false, os.Stdout, &infoClose)
}

func LogErrorToStderr() {
	maybeCloseAndLogTo(&ErrorLogger, false, os.Stderr, &errorClose)
}

func maybeCloseAndLogTo(l **logger.Logger, system bool, w io.Writer, shouldClose *bool) {
	if *shouldClose {
		(*l).Close()
	}
	*l = logger.Init("apcore", false, system, w)
	*shouldClose = !(w == os.Stdout || w == os.Stderr)
}
