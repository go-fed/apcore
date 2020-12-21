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

package app

import (
	"github.com/go-fed/apcore/util"
)

type Database interface {
	Begin() TxBuilder
}

type TxBuilder interface {
	QueryOneRow(sql string, cb func(r SingleRow) error, args ...interface{})
	Query(sql string, cb func(r SingleRow) error, args ...interface{})
	ExecOneRow(sql string, args ...interface{})
	Exec(sql string, args ...interface{})
	Do(c util.Context) error
}

type SingleRow interface {
	Scan(dest ...interface{}) error
}
