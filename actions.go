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
	"context"

	"github.com/go-fed/apcore/app"
)

func doCreateTables(configFilePath string, a app.Application, debug bool, scheme string) error {
	db, d, ms, err := newModels(configFilePath, a, debug, scheme)
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, m := range ms {
		if err := m.CreateTable(tx, d); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func doInitAdmin(configFilePath string, a app.Application, debug bool) error {
	// TODO
	return nil
}
