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
	"github.com/go-fed/apcore/framework"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

func doCreateTables(configFilePath string, a app.Application, debug bool, scheme string) error {
	db, d, ms, err := newModels(configFilePath, a, debug, scheme)
	if err != nil {
		return err
	}
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

func doInitAdmin(configFilePath string, a app.Application, debug bool, scheme string) error {
	db, users, c, err := newUserService(configFilePath, a, debug, scheme)
	if err != nil {
		return err
	}

	// Prompt for admin information
	p := services.CreateUserParameters{
		Scheme:     scheme,
		Host:       c.ServerConfig.Host,
		RSAKeySize: c.ServerConfig.RSAKeySize,
		HashParams: services.HashPasswordParameters{
			SaltSize:       c.ServerConfig.SaltSize,
			BCryptStrength: c.ServerConfig.BCryptStrength,
		},
	}
	var password string
	p.Username, p.Email, password, err = framework.PromptAdminUser()

	// Create the user in the database
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = users.CreateAdminUser(util.Context{context.Background()}, p, password)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func doInitData(configFilePath string, a app.Application, debug bool, scheme string) error {
	db, users, c, err := newUserService(configFilePath, a, debug, scheme)
	if err != nil {
		return err
	}

	// Create the server actor in the database
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = users.CreateInstanceActorSingleton(util.Context{context.Background()}, scheme, c.ServerConfig.Host, c.ServerConfig.RSAKeySize)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func doInitServerProfile(configFilePath string, a app.Application, debug bool, scheme string) error {
	db, users, c, err := newUserService(configFilePath, a, debug, scheme)
	if err != nil {
		return err
	}
	defer db.Close()

	sp, err := framework.PromptServerProfile(scheme, c.ServerConfig.Host)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = users.SetServerPreferences(util.Context{context.Background()}, sp)
	if err != nil {
		return err
	}
	return tx.Commit()
}
