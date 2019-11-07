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
	"time"
)

// MyConfig is an example struct highlighting how an implementation can have
// automatic support in the apcore configuration file.
//
// Note that values put into the configuration are ones that remain constant
// throughout the running lifetime of the apcore.Application.
type MyConfig struct {
	FieldS string    `ini:"my_test_app_s" comment:"First test field"`
	FieldT int       `ini:"my_test_app_t" comment:"Second test field"`
	FieldU time.Time `ini:"my_test_app_u" comment:"Third test field"`
}
