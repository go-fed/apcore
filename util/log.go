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
	ErrorLogger *logger.Logger = logger.Init("apcore", false, false, os.Stderr)
)

func LogInfoTo(system bool, w io.Writer) {
	closeAndLogTo(&InfoLogger, system, w)
}

func LogErrorTo(system bool, w io.Writer) {
	closeAndLogTo(&ErrorLogger, system, w)
}

func LogInfoToStdout() {
	closeAndLogTo(&InfoLogger, false, os.Stdout)
}

func LogErrorToStderr() {
	closeAndLogTo(&ErrorLogger, false, os.Stderr)
}

func closeAndLogTo(l **logger.Logger, system bool, w io.Writer) {
	(*l).Close()
	*l = logger.Init("apcore", false, system, w)
}
