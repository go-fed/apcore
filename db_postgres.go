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
	UpdateUserPolicy() string
	UpdateInstancePolicy() string
	InsertUserPolicy() string
	InsertInstancePolicy() string
	InstancePolicies() string
	UserPolicies() string
	InsertResolutions() string
	UserResolutions() string
	InboxContains() string
	GetInbox() string
	SetInboxUpdate() string
	SetInboxInsert() string
	SetInboxDelete() string
	ActorForOutbox() string
	ActorForInbox() string
	OutboxForInbox() string
	Exists() string
	Get() string
	LocalCreate() string
	FedCreate() string
	LocalUpdate() string
	FedUpdate() string
	LocalDelete() string
	FedDelete() string
	GetOutbox() string
	SetOutboxUpdate() string
	SetOutboxInsert() string
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
	err = p.maybeLogExecute(t, p.instancePolicyTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.userPolicyTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.resolutionTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.resolutionUserPolicyJoinTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.resolutionInstancePolicyJoinTable())
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
		InfoLogger.Infof("SQL exec: %s", s)
	}
	_, err = t.Exec(s)
	return
}

func (p *pgV0) fedDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `fed_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
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
  create_time timestamp with time zone DEFAULT current_timestamp,
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
  create_time timestamp with time zone DEFAULT current_timestamp,
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
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE RESTRICT,
  federated_id uuid REFERENCES ` + p.schema + `fed_data (id) NOT NULL ON DELETE CASCADE,
);`
}

func (p *pgV0) usersOutboxTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users_inbox
(
  id bigserial PRIMARY KEY,
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE RESTRICT,
  local_id uuid REFERENCES ` + p.schema + `local_data (id) NOT NULL ON DELETE CASCADE,
);`
}

func (p *pgV0) userPrivilegesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_privileges
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) userPreferencesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_preferences
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) instancePolicyTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `instance_policies
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
  order integer NOT NULL CONSTRAINT unique_order UNIQUE DEFERRABLE INITIALLY DEFERRED,
  description text NOT NULL,
  subject text NOT NULL,
  kind text NOT NULL
);`
}

func (p *pgV0) userPolicyTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `user_policies
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE CASCADE,
  order integer NOT NULL,
  description text NOT NULL,
  subject text NOT NULL,
  kind text NOT NULL,
  CONSTRAINT user_unique_order UNIQUE (user_id, order) DEFERRABLE INITIALLY DEFERRED
);`
}

func (p *pgV0) resolutionTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `resolutions
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
  order integer NOT NULL,
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE CASCADE,
  permitted integer NOT NULL,
  activity_iri text NOT NULL,
  is_public boolean NOT NULL,
  reason text NOT NULL
  CONSTRAINT activity_unique_order UNIQUE (activity_iri, order) DEFERRABLE INITIALLY DEFERRED
);`
}

func (p *pgV0) resolutionInstancePolicyJoinTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `resolutions_instance_policies
(
  resolution_id uuid REFERENCES ` + p.schema + `resolutions (id) NOT NULL ON DELETE CASCADE,
  instance_policy_id uuid REFERENCES ` + p.schema + `instance_policies (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) resolutionUserPolicyJoinTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `resolutions_user_policies
(
  resolution_id uuid REFERENCES ` + p.schema + `resolutions (id) NOT NULL ON DELETE CASCADE,
  user_policy_id uuid REFERENCES ` + p.schema + `user_policies (id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) UpdateUserPolicy() string {
	// TODO
	return ""
}

func (p *pgV0) UpdateInstancePolicy() string {
	// TODO
	return ""
}

func (p *pgV0) InsertUserPolicy() string {
	// TODO
	return ""
}

func (p *pgV0) InsertInstancePolicy() string {
	// TODO
	return ""
}

func (p *pgV0) InstancePolicies() string {
	// TODO
	return ""
}

func (p *pgV0) UserPolicies() string {
	// TODO
	return ""
}

func (p *pgV0) InsertResolutions() string {
	// TODO
	return ""
}

func (p *pgV0) UserResolutions() string {
	// TODO
	return ""
}

func (p *pgV0) InboxContains() string {
	// TODO
	return ""
}

func (p *pgV0) GetInbox() string {
	// TODO
	return ""
}

func (p *pgV0) SetInboxUpdate() string {
	// TODO
	return ""
}

func (p *pgV0) SetInboxInsert() string {
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

func (p *pgV0) LocalCreate() string {
	// TODO
	return ""
}

func (p *pgV0) FedCreate() string {
	// TODO
	return ""
}

func (p *pgV0) LocalUpdate() string {
	// TODO
	return ""
}

func (p *pgV0) FedUpdate() string {
	// TODO
	return ""
}

func (p *pgV0) LocalDelete() string {
	// TODO
	return ""
}

func (p *pgV0) FedDelete() string {
	// TODO
	return ""
}

func (p *pgV0) GetOutbox() string {
	// TODO
	return ""
}

func (p *pgV0) SetOutboxUpdate() string {
	// TODO
	return ""
}

func (p *pgV0) SetOutboxInsert() string {
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
