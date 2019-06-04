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

type sqlManager interface {
	CreateTables(t *sql.Tx) (err error)
	UpgradeTables(t *sql.Tx) (err error)
}

type sqlGenerator interface {
	InboxContains() string
	GetInbox() string
	SetInboxUpsert() string
	SetInboxDelete() string
	ActorForOutbox() string
	ActorForInbox() string
	OutboxForInbox() string
	Exists() string
	Get() string
	Create() string
	Update() string
	Delete() string
	GetOutbox() string
	SetOutboxUpsert() string
	SetOutboxDelete() string
	Followers() string
	Following() string
	Liked() string
}

var _ sqlManager = &pgV0{}
var _ sqlGenerator = &pgV0{}

type pgV0 struct {
	schema string
	log    bool
}

func newPgV0(schema string, log bool) *pgV0 {
	p := &pgV0{
		schema: schema,
		log:    log,
	}
	if p.schema == "" {
		p.schema = "public"
	}
	p.schema += "."
	return p
}

func (p *pgV0) CreateTables(t *sql.Tx) (err error) {
	InfoLogger.Info("Running Postgres create tables v0")

	// Create tables
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
	err = p.maybeLogExecute(t, p.usersInboxTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.usersOutboxTable())
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

	// Create indexes
	err = p.maybeLogExecute(t, p.indexFedDataTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.indexLocalDataTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.indexUsersTable())
	if err != nil {
		return
	}
	return
}

func (p *pgV0) UpgradeTables(t *sql.Tx) (err error) {
	err = fmt.Errorf("cannot upgrade Postgres tables to first version v0")
	return
}

func (p *pgV0) maybeLogExecute(t *sql.Tx, s string) (err error) {
	if p.log {
		InfoLogger.Info("SQL exec: %s", s)
	}
	_, err = t.Exec(s)
	return
}

func (p *pgV0) fedDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `fed_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  payload jsonb NOT NULL
);`
}

func (p *pgV0) indexFedDataTable() string {
	return `CREATE INDEX IF NOT EXISTS fed_data_jsonb_index ON ` + p.schema + `fed_data USING GIN (payload);`
}

func (p *pgV0) localDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `local_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  payload jsonb NOT NULL
);`
}

func (p *pgV0) indexLocalDataTable() string {
	return `CREATE INDEX IF NOT EXISTS local_data_jsonb_index ON ` + p.schema + `local_data USING GIN (payload);`
}

func (p *pgV0) usersTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor jsonb NOT NULL
);`
}

func (p *pgV0) indexUsersTable() string {
	return `CREATE INDEX IF NOT EXISTS users_jsonb_index ON ` + p.schema + `users USING GIN (actor);`
}

func (p *pgV0) usersInboxTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users_inbox
(
  id bigserial PRIMARY KEY,
  user_id uuid REFERENCES users (id) NOT NULL ON DELETE RESTRICT,
  federated_id uuid REFERENCES fed_data (id) NOT NULL ON DELETE CASCADE,
);`
}

func (p *pgV0) usersOutboxTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users_inbox
(
  id bigserial PRIMARY KEY,
  user_id uuid REFERENCES users (id) NOT NULL ON DELETE RESTRICT,
  local_id uuid REFERENCES local_data (id) NOT NULL ON DELETE CASCADE,
);`
}

func (p *pgV0) userPrivilegesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_privileges
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) userPreferencesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_preferences
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) userFedRules() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_fed_rules
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) serverFedRules() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `server_fed_rules
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid()
);`
}

func (p *pgV0) fedDataRuleAnnotationsTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `fed_data_rule_annotations
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  fed_data_id uuid REFERENCES fed_data (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) localDataRuleAnnotationsTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `local_data_rule_annotations
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  local_data_id uuid REFERENCES local_data (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) InboxContains() string {
	// TODO
	return ""
}

func (p *pgV0) GetInbox() string {
	// TODO
	return ""
}

func (p *pgV0) SetInboxUpsert() string {
	// TODO
	return ""
}

func (p *pgV0) SetInboxDelete() string {
	// TODO
	return ""
}

func (p *pgV0) ActorForOutbox() string {
	// TODO
	return ""
}

func (p *pgV0) ActorForInbox() string {
	// TODO
	return ""
}

func (p *pgV0) OutboxForInbox() string {
	// TODO
	return ""
}

func (p *pgV0) Exists() string {
	// TODO
	return ""
}

func (p *pgV0) Get() string {
	// TODO
	return ""
}

func (p *pgV0) Create() string {
	// TODO
	return ""
}

func (p *pgV0) Update() string {
	// TODO
	return ""
}

func (p *pgV0) Delete() string {
	// TODO
	return ""
}

func (p *pgV0) GetOutbox() string {
	// TODO
	return ""
}

func (p *pgV0) SetOutboxUpsert() string {
	// TODO
	return ""
}

func (p *pgV0) SetOutboxDelete() string {
	// TODO
	return ""
}

func (p *pgV0) Followers() string {
	// TODO
	return ""
}

func (p *pgV0) Following() string {
	// TODO
	return ""
}

func (p *pgV0) Liked() string {
	// TODO
	return ""
}
