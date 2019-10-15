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

func promptAdminUser() (username, email, password string, err error) {
	username, err = promptStringWithDefault(
		"Enter the new admin account's username",
		"")
	if err != nil {
		return
	}
	email, err = promptStringWithDefault(
		"Enter the new admin account's email address (will NOT be verified)",
		"")
	if err != nil {
		return
	}
	password, err = promptPassword("Enter the new admin account's password")
	return
}
