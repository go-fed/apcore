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
	err = p.maybeLogExecute(t, p.deliveryAttemptTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.privateKeyTable())
	if err != nil {
		return
	}
	// OAuth information
	err = p.maybeLogExecute(t, p.tokenTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.clientTable())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.indexTokenCode())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.indexTokenAccess())
	if err != nil {
		return
	}
	err = p.maybeLogExecute(t, p.indexTokenRefresh())
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
  create_time timestamp with time zone NOT NULL DEFAULT current_timestamp,
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
  create_time timestamp with time zone NOT NULL DEFAULT current_timestamp,
  email text NOT NULL,
  hashpass bytea NOT NULL,
  salt bytea NOT NULL,
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
  user_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE CASCADE,
  on_follow text NOT NULL
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
  permitted string NOT NULL,
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

func (p *pgV0) deliveryAttemptTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `delivery_attempts
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
  from_id uuid REFERENCES ` + p.schema + `users (id) NOT NULL ON DELETE CASCADE,
  to text NOT NULL,
  payload bytea NOT NULL,
  state text NOT NULL
);`
}

func (p *pgV0) privateKeyTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `private_keys
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES ` + p.schema + `users(id) NOT NULL ON DELETE CASCADE,
  create_time timestamp with time zone DEFAULT current_timestamp,
  priv_key bytea NOT NULL
);`
}

func (p *pgV0) tokenTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `oauth_tokens
(
  client_id text NOT NULL,
  user_id text NOT NULL,
  redirect_uri text NOT NULL,
  scope text NOT NULL,
  code text NOT NULL,
  code_create_at timestamp with time zone NOT NULL,
  code_expires_in bigint NOT NULL,
  access text NOT NULL,
  access_create_at timestamp with time zone NOT NULL,
  access_expires_in bigint NOT NULL,
  refresh text NOT NULL,
  refresh_create_at timestamp with time zone NOT NULL,
  refresh_expires_in bigint NOT NULL
);`
}

func (p *pgV0) indexTokenCode() string {
	return `CREATE INDEX IF NOT EXISTS oauth_tokens_code_index ON ` + p.schema + `oauth_tokens (code);`
}

func (p *pgV0) indexTokenAccess() string {
	return `CREATE INDEX IF NOT EXISTS oauth_tokens_access_index ON ` + p.schema + `oauth_tokens (access);`
}

func (p *pgV0) indexTokenRefresh() string {
	return `CREATE INDEX IF NOT EXISTS oauth_tokens_refresh_index ON ` + p.schema + `oauth_tokens (refresh);`
}

func (p *pgV0) clientTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `oauth_clients
(
  id text PRIMARY KEY,
  secret text NOT NULL,
  domain text NOT NULL,
  user_id uuid REFERENCES ` + p.schema + `users(id) NOT NULL ON DELETE CASCADE
);`
}

func (p *pgV0) HashPassForUserID() string {
	return "SELECT hashpass, salt FROM " + p.schema + "users WHERE id = $1"
}

func (p *pgV0) UserIdForEmail() string {
	return "SELECT id FROM " + p.schema + "users WHERE email = $1"
}

func (p *pgV0) UserIdForBoxPath() string {
	return "SELECT id FROM " + p.schema + "users WHERE (actor->>'inbox' = $1 OR actor->>'outbox' = $2)"
}

func (p *pgV0) UserPreferences() string {
	return "SELECT on_follow FROM " + p.schema + "user_preferences WHERE user_id = $1"
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

func (p *pgV0) InsertUserPKey() string {
	return "INSERT INTO " + p.schema + "private_keys (user_id, priv_key) VALUES ($1, $2)"
}

func (p *pgV0) GetUserPKey() string {
	return "SELECT id, priv_key FROM " + p.schema + "private_keys WHERE user_id = $1"
}

func (p *pgV0) FollowersByUserUUID() string {
	return `SELECT local_data.payload FROM ` + p.schema + `local_data
INNER JOIN` + p.schema + `users
ON users.actor->>'followers' = local_data.payload->>'id'
WHERE users.id = $1`
}

func (p *pgV0) InsertAttempt() string {
	return "INSERT INTO " + p.schema + "delivery_attempts (from_id, to, payload, state) VALUES ($1, $2, $3, 'new')"
}

func (p *pgV0) MarkSuccessfulAttempt() string {
	return "UPDATE " + p.schema + "delivery_attempts SET (state) = ('success') WHERE id = $1"
}

func (p *pgV0) MarkRetryFailureAttempt() string {
	return "UPDATE " + p.schema + "delivery_attempts SET (state) = ('fail') WHERE id = $1"
}

func (p *pgV0) CreateTokenInfo() string {
	return "INSERT INTO " + p.schema + `oauth_tokens
(
  client_id,
  user_id,
  redirect_uri,
  scope,
  code,
  code_create_at,
  code_expires_in,
  access,
  access_create_at,
  access_expires_in,
  refresh,
  refresh_create_at,
  refresh_expires_in
) VALUES
(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13
)`
}

func (p *pgV0) RemoveTokenByCode() string {
	return "DELETE FROM " + p.schema + "oauth_tokens WHERE code = $1"
}

func (p *pgV0) RemoveTokenByAccess() string {
	return "DELETE FROM " + p.schema + "oauth_tokens WHERE access = $1"
}

func (p *pgV0) RemoveTokenByRefresh() string {
	return "DELETE FROM " + p.schema + "oauth_tokens WHERE refresh = $1"
}

func (p *pgV0) GetTokenByCode() string {
	return `SELECT
(
  client_id,
  user_id,
  redirect_uri,
  scope,
  code,
  code_create_at,
  code_expires_in,
  access,
  access_create_at,
  access_expires_in,
  refresh,
  refresh_create_at,
  refresh_expires_in
) 
FROM ` + p.schema + "oauth_tokens WHERE code = $1"
}

func (p *pgV0) GetTokenByAccess() string {
	return `SELECT
(
  client_id,
  user_id,
  redirect_uri,
  scope,
  code,
  code_create_at,
  code_expires_in,
  access,
  access_create_at,
  access_expires_in,
  refresh,
  refresh_create_at,
  refresh_expires_in
) 
FROM ` + p.schema + "oauth_tokens WHERE access = $1"
}

func (p *pgV0) GetTokenByRefresh() string {
	return `SELECT
(
  client_id,
  user_id,
  redirect_uri,
  scope,
  code,
  code_create_at,
  code_expires_in,
  access,
  access_create_at,
  access_expires_in,
  refresh,
  refresh_create_at,
  refresh_expires_in
) 
FROM ` + p.schema + "oauth_tokens WHERE refresh = $1"
}

func (p *pgV0) GetClientById() string {
	return "SELECT (id, secret, domain, user_id) FROM " + p.schema + "oauth_clients WHERE id = $1"
}

func (p *pgV0) InboxContains() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `users AS u
  INNER JOIN ` + p.schema + `users_inbox AS ui
  ON u.id = ui.user_id
  INNER JOIN ` + p.schema + `fed_data AS f
  ON ui.federated_id = f.id
  WHERE u.actor->>'inbox' = $1 AND f.payload->>'id' = $2
);`
}

func (p *pgV0) GetInbox() string {
	return `SELECT ui.id, u.id, f.id, f.payload->>'id'
FROM ` + p.schema + `users AS u
INNER JOIN ` + p.schema + `users_inbox AS ui
ON u.id = ui.user_id
INNER JOIN ` + p.schema + `fed_data AS f
ON ui.federated_id = f.id
WHERE u.actor->>'inbox' = $1
ORDER BY f.create_time DESC
LIMIT $3 OFFSET $2`
}

func (p *pgV0) GetPublicInbox() string {
	return `SELECT ui.id, u.id, f.id, f.payload->>'id'
FROM ` + p.schema + `users AS u
INNER JOIN ` + p.schema + `users_inbox AS ui
ON u.id = ui.user_id
INNER JOIN ` + p.schema + `fed_data AS f
ON ui.federated_id = f.id
WHERE u.actor->>'inbox' = $1 AND
(
  f.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
  OR f.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
)
ORDER BY f.create_time DESC
LIMIT $3 OFFSET $2`
}

func (p *pgV0) SetInboxUpdate() string {
	return `WITH fed_query AS (
  SELECT fed_data.id FROM ` + p.schema + `fed_data WHERE fed_data.payload->>'id' = $3
)
UPDATE ` + p.schema + `users_outbox
SET (federated_id) = (fed_query.id)
FROM fed_query
WHERE id = $1 AND user_id = $2`
}

func (p *pgV0) SetInboxInsert() string {
	return `INSERT INTO ` + p.schema + `users_inbox (user_id, federated_id)
SELECT users.id, fed_data.id FROM ` + p.schema + `users, ` + p.schema + `fed_data
WHERE users.actor->>'inbox' = $1 AND fed_data.payload->>'id' = $2`
}

func (p *pgV0) SetInboxDelete() string {
	return "DELETE FROM " + p.schema + "users_inbox WHERE id = $1"
}

func (p *pgV0) ActorForOutbox() string {
	return `SELECT actor->>'id' FROM ` + p.schema + `users
WHERE actor->>'outbox' = $1`
}

func (p *pgV0) ActorForInbox() string {
	return `SELECT actor->>'id' FROM ` + p.schema + `users
WHERE actor->>'inbox' = $1`
}

func (p *pgV0) OutboxForInbox() string {
	return `SELECT actor->>'outbox' FROM ` + p.schema + `users
WHERE actor->>'inbox' = $1`
}

func (p *pgV0) Exists() string {
	return `SELECT EXISTS(
SELECT 1 FROM ` + p.schema + `fed_data
WHERE payload->>'id' = $1
)`
}

func (p *pgV0) Get() string {
	return `SELECT payload FROM ` + p.schema + `fed_data WHERE payload->>'id' = $1
UNION
SELECT payload FROM ` + p.schema + `local_data WHERE payload->>'id' = $1
UNION
SELECT actor FROM ` + p.schema + `users WHERE actor->>'id' = $1`
}

func (p *pgV0) LocalCreate() string {
	return `INSERT INTO ` + p.schema + `local_data (payload) VALUES ($1)`
}

func (p *pgV0) FedCreate() string {
	return `INSERT INTO ` + p.schema + `fed_data (payload) VALUES ($1)`
}

func (p *pgV0) LocalUpdate() string {
	return `UPDATE ` + p.schema + `local_data SET (payload) = ($2) WHERE payload->>'id' = $1`
}

func (p *pgV0) FedUpdate() string {
	return `UPDATE ` + p.schema + `fed_data SET (payload) = ($2) WHERE payload->>'id' = $1`
}

func (p *pgV0) LocalDelete() string {
	return `DELETE FROM ` + p.schema + `local_data WHERE payload->>'id' = $1`
}

func (p *pgV0) FedDelete() string {
	return `DELETE FROM ` + p.schema + `fed_data WHERE payload->>'id' = $1`
}

func (p *pgV0) GetOutbox() string {
	return `SELECT ui.id, u.id, l.id, l.payload->>'id'
FROM ` + p.schema + `users AS u
INNER JOIN ` + p.schema + `users_inbox AS ui
ON u.id = ui.user_id
INNER JOIN ` + p.schema + `local_data AS l
ON ui.local_id = l.id
WHERE u.actor->>'outbox' = $1
ORDER BY l.create_time DESC
LIMIT $3 OFFSET $2`
}

func (p *pgV0) GetPublicOutbox() string {
	return `SELECT ui.id, u.id, l.id, l.payload->>'id'
FROM ` + p.schema + `users AS u
INNER JOIN ` + p.schema + `users_inbox AS ui
ON u.id = ui.user_id
INNER JOIN ` + p.schema + `local_data AS l
ON ui.local_id = l.id
WHERE u.actor->>'outbox' = $1 AND
(
  l.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
  OR l.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
)
ORDER BY l.create_time DESC
LIMIT $3 OFFSET $2`
}

func (p *pgV0) SetOutboxUpdate() string {
	return `WITH local_query AS (
  SELECT local_data.id FROM ` + p.schema + `local_data WHERE local_data.payload->>'id' = $3
)
UPDATE ` + p.schema + `users_outbox
SET (local_id) = (local_query.id)
FROM local_query
WHERE id = $1 AND user_id = $2`
}

func (p *pgV0) SetOutboxInsert() string {
	return `INSERT INTO ` + p.schema + `users_outbox (user_id, local_id)
SELECT users.id, local_data.id FROM ` + p.schema + `users, ` + p.schema + `local_data
WHERE users.actor->>'inbox' = $1 AND local_data.payload->>'id' = $2`
}

func (p *pgV0) SetOutboxDelete() string {
	return "DELETE FROM " + p.schema + "users_outbox WHERE id = $1"
}

func (p *pgV0) Followers() string {
	return `SELECT local_data.payload FROM ` + p.schema + `local_data
INNER JOIN` + p.schema + `users
ON users.actor->>'followers' = local_data.payload->>'id'
WHERE users.actor->>'id' = $1`
}

func (p *pgV0) Following() string {
	return `SELECT local_data.payload FROM ` + p.schema + `local_data
INNER JOIN` + p.schema + `users
ON users.actor->>'following' = local_data.payload->>'id'
WHERE users.actor->>'id' = $1`
}

func (p *pgV0) Liked() string {
	return `SELECT local_data.payload FROM ` + p.schema + `local_data
INNER JOIN` + p.schema + `users
ON users.actor->>'liked' = local_data.payload->>'id'
WHERE users.actor->>'id' = $1`
}
