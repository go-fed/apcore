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

// Package apcore implements a generic, extensible ActivityPub server using
// the go-fed libraries.
package apcore

import (
	"github.com/go-fed/apcore/app"
)

const (
	apcoreName         = "apcore"
	apcoreMajorVersion = 0
	apcoreMinorVersion = 1
	apcorePatchVersion = 0
)

func apCoreSoftware() app.Software {
	return app.Software{
		Name:         apcoreName,
		MajorVersion: apcoreMajorVersion,
		MinorVersion: apcoreMinorVersion,
		PatchVersion: apcorePatchVersion,
	}
}
