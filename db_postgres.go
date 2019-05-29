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

import (
	"database/sql"
	"fmt"
)

type sqlCreateTables interface {
	CreateTables(t *sql.Tx) (err error)
	UpgradeTables(t *sql.Tx) (err error)
}

var _ sqlCreateTables = &pgCreateTablesV0{}

type pgCreateTablesV0 struct {
	schema string
	log    bool
}

func (p *pgCreateTablesV0) CreateTables(t *sql.Tx) (err error) {
	InfoLogger.Info("Running Postgres create tables v0")
	err = p.maybeLogExecute(t, p.fedDataTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.localDataTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.usersTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.userPrivilegesTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.userPreferencesTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.userFedRules())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.serverFedRules())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.fedDataRuleAnnotationsTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.localDataRuleAnnotationsTable())
	if err != nil {
		return
	}
	return
}

func (p *pgCreateTablesV0) UpgradeTables(t *sql.Tx) (err error) {
	err = fmt.Errorf("cannot upgrade Postgres tables to first version v0")
	return
}

func (p *pgCreateTablesV0) maybeLogExecute(t *sql.Tx, s string) (err error) {
	if p.log {
		InfoLogger.Info("SQL exec: %s", s)
	}
	_, err = t.Exec(s)
	return
}

func (p *pgCreateTablesV0) fedDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `fed_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  payload jsonb NOT NULL
);`
}

func (p *pgCreateTablesV0) localDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `local_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  payload jsonb NOT NULL
);`
}

func (p *pgCreateTablesV0) usersTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid()
);`
}

func (p *pgCreateTablesV0) userPrivilegesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_privileges
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users (id)
);`
}

func (p *pgCreateTablesV0) userPreferencesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_preferences
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users (id)
);`
}

func (p *pgCreateTablesV0) userFedRules() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_fed_rules
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users (id)
);`
}

func (p *pgCreateTablesV0) serverFedRules() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `server_fed_rules
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid()
);`
}

func (p *pgCreateTablesV0) fedDataRuleAnnotationsTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `fed_data_rule_annotations
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  fed_data_id uuid REFERENCES fed_data (id)
);`
}

func (p *pgCreateTablesV0) localDataRuleAnnotationsTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `local_data_rule_annotations
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  local_data_id uuid REFERENCES local_data (id)
);`
}
