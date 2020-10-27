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

package db

import (
	"github.com/go-fed/apcore/models"
)

var _ models.SqlDialect = &pgV0{}

type pgV0 struct {
	schema string
}

func NewPgV0(schema string) *pgV0 {
	p := &pgV0{
		schema: schema,
	}
	if p.schema == "" {
		p.schema = "public"
	}
	p.schema += "."
	return p
}

// TODO
func (p *pgV0) indexTokenCode() string {
	return `CREATE INDEX IF NOT EXISTS oauth_tokens_code_index ON ` + p.schema + `oauth_tokens (code);`
}

// TODO
func (p *pgV0) indexTokenAccess() string {
	return `CREATE INDEX IF NOT EXISTS oauth_tokens_access_index ON ` + p.schema + `oauth_tokens (access);`
}

// TODO
func (p *pgV0) indexTokenRefresh() string {
	return `CREATE INDEX IF NOT EXISTS oauth_tokens_refresh_index ON ` + p.schema + `oauth_tokens (refresh);`
}

/*
func (p *pgV0) FollowersByUserUUID() string {
	return `SELECT local_data.payload FROM ` + p.schema + `local_data
INNER JOIN` + p.schema + `users
ON users.actor->>'followers' = local_data.payload->>'id'
WHERE users.id = $1`
}

func (p *pgV0) RemoveTokenByAccess() string {
	return "DELETE FROM " + p.schema + "oauth_tokens WHERE access = $1"
}

func (p *pgV0) RemoveTokenByRefresh() string {
	return "DELETE FROM " + p.schema + "oauth_tokens WHERE refresh = $1"
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

func (p *pgV0) InsertUserPrivileges() string {
	return `INSERT INTO ` + p.schema + `user_privileges (user_id, admin) VALUES ($1, $2)`
}

func (p *pgV0) InsertUserPreferences() string {
	return `INSERT INTO ` + p.schema + `user_preferences (user_id, on_follow) VALUES ($1, $2)`
}
*/

/* SqlDialect */

func (p *pgV0) CreateUsersTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone NOT NULL DEFAULT current_timestamp,
  email text NOT NULL,
  hashpass bytea NOT NULL,
  salt bytea NOT NULL,
  actor jsonb NOT NULL,
  privileges jsonb NOT NULL,
  preferences jsonb NOT NULL
);`
}

func (p *pgV0) InsertUser() string {
	return `INSERT INTO ` + p.schema + `users (email, hashpass, salt, actor, privileges, preferences) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
}

func (p *pgV0) SensitiveUserByEmail() string {
	return "SELECT id, hashpass, salt FROM " + p.schema + "users WHERE email = $1"
}

func (p *pgV0) UserByID() string {
	return "SELECT id, email, actor, privileges, preferences FROM " + p.schema + "users WHERE id = $1"
}

func (p *pgV0) ActorIDForOutbox() string {
	return `SELECT actor->>'id' FROM ` + p.schema + `users
WHERE actor->'outbox' ? $1`
}

func (p *pgV0) ActorIDForInbox() string {
	return `SELECT actor->>'id' FROM ` + p.schema + `users
WHERE actor->'inbox' ? $1`
}

func (p *pgV0) CreateFedDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `fed_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
  payload jsonb NOT NULL
);`
}

func (p *pgV0) FedCreate() string {
	return `INSERT INTO ` + p.schema + `fed_data (payload) VALUES ($1)`
}

func (p *pgV0) FedUpdate() string {
	return `UPDATE ` + p.schema + `fed_data SET payload = $2 WHERE payload->>'id' = $1`
}

func (p *pgV0) FedDelete() string {
	return `DELETE FROM ` + p.schema + `fed_data WHERE payload->>'id' = $1`
}

func (p *pgV0) CreateLocalDataTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `local_data
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone NOT NULL DEFAULT current_timestamp,
  payload jsonb NOT NULL
);`
}

func (p *pgV0) LocalCreate() string {
	return `INSERT INTO ` + p.schema + `local_data (payload) VALUES ($1)`
}

func (p *pgV0) LocalUpdate() string {
	return `UPDATE ` + p.schema + `local_data SET payload = $2 WHERE payload->>'id' = $1`
}

func (p *pgV0) LocalDelete() string {
	return `DELETE FROM ` + p.schema + `local_data WHERE payload->>'id' = $1`
}

func (p *pgV0) CreateInboxesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `inboxes
(
  id bigserial PRIMARY KEY,
  actor_id text NOT NULL,
  inbox jsonb NOT NULL
);`
}

func (p *pgV0) CreateOutboxesTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `outboxes
(
  id bigserial PRIMARY KEY,
  actor_id text NOT NULL,
  outbox jsonb NOT NULL
);`
}

func (p *pgV0) InsertInbox() string {
	return `INSERT INTO ` + p.schema + `inboxes (actor_id, inbox) VALUES ($1, $2)`
}

func (p *pgV0) InsertOutbox() string {
	return `INSERT INTO ` + p.schema + `outboxes (actor_id, outbox) VALUES ($1, $2)`
}

func (p *pgV0) InboxContainsForActor() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `inboxes
  WHERE actor_id = $1 AND inbox->'orderedItems' ? $2
  LIMIT 1
)`
}

func (p *pgV0) InboxContains() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `inboxes
  WHERE inbox->'id' ? $1 AND inbox->'orderedItems' ? $2
  LIMIT 1
)`
}

func (p *pgV0) OutboxContainsForActor() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `outboxes
  WHERE actor_id = $1 AND outbox->'orderedItems' ? $2
  LIMIT 1
)`
}

func (p *pgV0) OutboxContains() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `outboxes
  WHERE outbox->'id' ? $1 AND outbox->'orderedItems' ? $2
  LIMIT 1
)`
}

func (p *pgV0) GetInbox() string {
	return `WITH page AS (
  SELECT
    inbox,
    jsonb_path_query_array(
      inbox,
      '$.orderedItems[$min to $max]',
      jsonb_build_object(
        'min',
	$2::jsonb,
        'max',
	$3::jsonb)) AS page
    FROM ` + p.schema + `inboxes
    WHERE inbox->'id' ? $1
)
SELECT
  inbox ||
    jsonb_build_object(
      'orderedItems',
      page,
      'totalItems',
      jsonb_path_query(page, '$.size()'),
      'type',
      'OrderedCollectionPage')
  FROM page`
}

func (p *pgV0) GetOutbox() string {
	return `WITH page AS (
  SELECT
    outbox,
    jsonb_path_query_array(
      outbox,
      '$.orderedItems[$min to $max]',
      jsonb_build_object(
        'min',
	$2::jsonb,
        'max',
	$3::jsonb)) AS page
    FROM ` + p.schema + `outboxes
    WHERE outbox->'id' ? $1
)
SELECT
  outbox ||
    jsonb_build_object(
      'orderedItems',
      page,
      'totalItems',
      jsonb_path_query(page, '$.size()'),
      'type',
      'OrderedCollectionPage')
  FROM page`
}

func (p *pgV0) GetPublicInbox() string {
	return `WITH inbox AS (
  SELECT inbox
  FROM modeltest.inboxes
  WHERE inbox->'id' ? $1
),
page_elements AS (
  SELECT
    jsonb_array_elements(
      jsonb_path_query_array(
        inbox,
        '$.orderedItems[*]')) AS page
  FROM inbox
),
fed_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.local_data AS ld
  ON pd.page = ld.payload->'id'
  WHERE
    ld.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR ld.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
only_public AS (
  SELECT
    jsonb_path_query_array(
      jsonb_agg(i.page),
      '$[$min to $max]',
      jsonb_build_object(
        'min',
        $2::jsonb,
        'max',
        $3::jsonb)) AS page
  FROM (
    SELECT
      *
    FROM fed_public
    UNION ALL
    SELECT
      *
    FROM local_public) AS i
)
SELECT
  i.inbox ||
    jsonb_build_object(
      'orderedItems',
      op.page,
      'totalItems',
      jsonb_path_query(op.page, '$.size()'),
      'type',
      'OrderedCollectionPage')
  FROM inbox AS i, only_public AS op`
}

func (p *pgV0) GetPublicOutbox() string {
	return `WITH outbox AS (
  SELECT outbox
  FROM modeltest.outboxes
  WHERE outbox->'id' ? $1
),
page_elements AS (
  SELECT
    jsonb_array_elements(
      jsonb_path_query_array(
        outbox,
        '$.orderedItems[*]')) AS page
  FROM outbox
),
fed_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.local_data AS ld
  ON pd.page = ld.payload->'id'
  WHERE
    ld.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR ld.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
only_public AS (
  SELECT
    jsonb_path_query_array(
      jsonb_agg(i.page),
      '$[$min to $max]',
      jsonb_build_object(
        'min',
        $2::jsonb,
        'max',
        $3::jsonb)) AS page
  FROM (
    SELECT
      *
    FROM fed_public
    UNION ALL
    SELECT
      *
    FROM local_public) AS i
)
SELECT
  i.outbox ||
    jsonb_build_object(
      'orderedItems',
      op.page,
      'totalItems',
      jsonb_path_query(op.page, '$.size()'),
      'type',
      'OrderedCollectionPage')
  FROM outbox AS i, only_public AS op`
}

func (p *pgV0) GetInboxLastPage() string {
	return `WITH stats AS (
  SELECT
    inbox,
    GREATEST(0,
      jsonb_path_query(inbox, '$.orderedItems.size()')::numeric - $2) AS startIndex
  FROM modeltest.inboxes
  WHERE inbox->'id' ? $1
),
page AS (
  SELECT
    inbox,
    startIndex,
    jsonb_path_query_array(
      inbox,
      '$.orderedItems[$min to last]',
      jsonb_build_object(
        'min',
        startIndex)) AS page
  FROM stats
)
SELECT
  inbox ||
    jsonb_build_object(
    'orderedItems',
    page,
    'totalItems',
    jsonb_path_query(page, '$.size()'),
    'type',
    'OrderedCollectionPage') AS inbox,
  startIndex
FROM page`
}

func (p *pgV0) GetOutboxLastPage() string {
	return `WITH stats AS (
  SELECT
    outbox,
    GREATEST(0,
      jsonb_path_query(outbox, '$.orderedItems.size()')::numeric - $2) AS startIndex
  FROM modeltest.outboxes
  WHERE outbox->'id' ? $1
),
page AS (
  SELECT
    outbox,
    startIndex,
    jsonb_path_query_array(
      outbox,
      '$.orderedItems[$min to last]',
      jsonb_build_object(
        'min',
        startIndex)) AS page
  FROM stats
)
SELECT
  outbox ||
    jsonb_build_object(
    'orderedItems',
    page,
    'totalItems',
    jsonb_path_query(page, '$.size()'),
    'type',
    'OrderedCollectionPage') AS outbox,
  startIndex
FROM page`
}

func (p *pgV0) GetPublicInboxLastPage() string {
	return `WITH inbox AS (
  SELECT inbox
  FROM modeltest.inboxes
  WHERE inbox->'id' ? $1
),
page_elements AS (
  SELECT
    jsonb_array_elements(
      jsonb_path_query_array(
        inbox,
        '$.orderedItems[*]')) AS page
  FROM inbox
),
fed_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.local_data AS ld
  ON pd.page = ld.payload->'id'
  WHERE
    ld.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR ld.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
merged AS (
  SELECT
    jsonb_agg(i.page) AS page,
	COUNT(i.page) AS n
  FROM (
    SELECT
      *
    FROM fed_public
    UNION ALL
    SELECT
      *
    FROM local_public) AS i
),
only_public AS (
  SELECT
    jsonb_path_query_array(
      page,
      '$[$min to last]',
      jsonb_build_object(
        'min',
        GREATEST(0, n - $2))) AS page,
	GREATEST(0, n - $2) AS startIndex
  FROM merged
)
SELECT
  i.inbox ||
    jsonb_build_object(
      'orderedItems',
      op.page,
      'totalItems',
      jsonb_path_query(op.page, '$.size()'),
      'type',
      'OrderedCollectionPage') AS inbox,
  op.startIndex
FROM inbox AS i, only_public AS op`
}

func (p *pgV0) GetPublicOutboxLastPage() string {
	return `WITH outbox AS (
  SELECT outbox
  FROM modeltest.outboxes
  WHERE outbox->'id' ? $1
),
page_elements AS (
  SELECT
    jsonb_array_elements(
      jsonb_path_query_array(
        outbox,
        '$.orderedItems[*]')) AS page
  FROM outbox
),
fed_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN modeltest.local_data AS ld
  ON pd.page = ld.payload->'id'
  WHERE
    ld.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR ld.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
merged AS (
  SELECT
    jsonb_agg(i.page) AS page,
	COUNT(i.page) AS n
  FROM (
    SELECT
      *
    FROM fed_public
    UNION ALL
    SELECT
      *
    FROM local_public) AS i
),
only_public AS (
  SELECT
    jsonb_path_query_array(
      page,
      '$[$min to last]',
      jsonb_build_object(
        'min',
        GREATEST(0, n - $2))) AS page,
	GREATEST(0, n - $2) AS startIndex
  FROM merged
)
SELECT
  i.outbox ||
    jsonb_build_object(
      'orderedItems',
      op.page,
      'totalItems',
      jsonb_path_query(op.page, '$.size()'),
      'type',
      'OrderedCollectionPage') AS outbox,
  op.startIndex
FROM outbox AS i, only_public AS op`
}

func (p *pgV0) PrependInboxItem() string {
	return `UPDATE modeltest.inboxes
SET inbox = inbox || jsonb_build_object(
  'orderedItems',
  jsonb_build_array($2::text) || (inbox->'orderedItems'),
  'totalItems',
  (COALESCE(inbox->>'totalItems','0')::int + 1)::text::jsonb)
WHERE inbox->'id' ? $1`
}

func (p *pgV0) PrependOutboxItem() string {
	return `UPDATE modeltest.outboxes
SET outbox = outbox || jsonb_build_object(
  'orderedItems',
  jsonb_build_array($2::text) || (outbox->'orderedItems'),
  'totalItems',
  (COALESCE(outbox->>'totalItems','0')::int + 1)::text::jsonb)
WHERE outbox->'id' ? $1`
}

func (p *pgV0) DeleteInboxItem() string {
	return `UPDATE modeltest.inboxes
SET inbox = jsonb_set(
  inbox,
  '{orderedItems}',
  (inbox->'orderedItems') - $2) ||
  jsonb_build_object(
  'totalItems',
  (COALESCE(inbox->>'totalItems','0')::int - 1)::text::jsonb)
WHERE inbox->'id' ? $1`
}

func (p *pgV0) DeleteOutboxItem() string {
	return `UPDATE modeltest.outboxes
SET outbox = jsonb_set(
  outbox,
  '{orderedItems}',
  (outbox->'orderedItems') - $2) ||
  jsonb_build_object(
  'totalItems',
  (COALESCE(outbox->>'totalItems','0')::int - 1)::text::jsonb)
WHERE outbox->'id' ? $1`
}

func (p *pgV0) OutboxForInbox() string {
	return `SELECT actor->>'outbox' FROM ` + p.schema + `users
WHERE actor->'inbox' ? $1`
}

func (p *pgV0) CreateDeliveryAttemptsTable() string {
	return `CREATE TABLE IF NOT EXISTS ` + p.schema + `delivery_attempts
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone DEFAULT current_timestamp,
  from_id uuid REFERENCES ` + p.schema + `users (id) ON DELETE CASCADE NOT NULL,
  deliver_to text NOT NULL,
  payload bytea NOT NULL,
  state text NOT NULL
);`
}

func (p *pgV0) InsertAttempt() string {
	return `INSERT INTO ` + p.schema + `delivery_attempts (from_id, deliver_to, payload, state) VALUES ($1, $2, $3, $4) RETURNING id`
}

func (p *pgV0) MarkSuccessfulAttempt() string {
	return `UPDATE ` + p.schema + `delivery_attempts SET state = $2 WHERE id = $1`
}

func (p *pgV0) MarkFailedAttempt() string {
	return `UPDATE ` + p.schema + `delivery_attempts SET state = $2 WHERE id = $1`
}

func (p *pgV0) CreatePrivateKeysTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `private_keys
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES ` + p.schema + `users(id) ON DELETE CASCADE NOT NULL,
  purpose text NOT NULL,
  priv_key bytea NOT NULL
);`
}

func (p *pgV0) CreatePrivateKey() string {
	return `INSERT INTO ` + p.schema + `private_keys (user_id, purpose, priv_key) VALUES ($1, $2, $3)`
}

func (p *pgV0) GetPrivateKeyByUserID() string {
	return `SELECT priv_key FROM ` + p.schema + `private_keys WHERE user_id = $1 AND purpose = $2`
}

func (p *pgV0) CreateClientInfosTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `oauth_clients
(
  id text PRIMARY KEY DEFAULT gen_random_uuid(),
  secret text NOT NULL,
  domain text NOT NULL,
  user_id uuid REFERENCES ` + p.schema + `users(id) ON DELETE CASCADE NOT NULL
);`
}

func (p *pgV0) CreateClientInfo() string {
	return `INSERT INTO ` + p.schema + `oauth_clients (secret, domain, user_id) VALUES ($1, $2, $3) RETURNING id`
}

func (p *pgV0) GetClientInfoByID() string {
	return `SELECT id, secret, domain, user_id FROM ` + p.schema + `oauth_clients WHERE id = $1`
}

func (p *pgV0) CreateTokenInfosTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `oauth_tokens
(
  client_id text REFERENCES ` + p.schema + `oauth_clients(id) ON DELETE CASCADE NOT NULL,
  user_id uuid REFERENCES ` + p.schema + `users(id) ON DELETE CASCADE NOT NULL,
  redirect_uri text NOT NULL,
  scope text NOT NULL,
  code text,
  code_create_at timestamp with time zone,
  code_expires_in bigint,
  access text,
  access_create_at timestamp with time zone,
  access_expires_in bigint,
  refresh text,
  refresh_create_at timestamp with time zone,
  refresh_expires_in bigint
)`
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

func (p *pgV0) RemoveTokenInfoByCode() string {
	return `DELETE FROM ` + p.schema + `oauth_tokens WHERE code = $1`
}

func (p *pgV0) RemoveTokenInfoByAccess() string {
	return `DELETE FROM ` + p.schema + `oauth_tokens WHERE access = $1`
}

func (p *pgV0) RemoveTokenInfoByRefresh() string {
	return `DELETE FROM ` + p.schema + `oauth_tokens WHERE refresh = $1`
}

func (p *pgV0) GetTokenInfoByCode() string {
	return `SELECT
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
FROM ` + p.schema + "oauth_tokens WHERE code = $1"
}

func (p *pgV0) GetTokenInfoByAccess() string {
	return `SELECT
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
FROM ` + p.schema + "oauth_tokens WHERE access = $1"
}

func (p *pgV0) GetTokenInfoByRefresh() string {
	return `SELECT
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
FROM ` + p.schema + "oauth_tokens WHERE refresh = $1"
}
