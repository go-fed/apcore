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

/* SqlDialect */

func (p *pgV0) CreateUsersTable() string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + `users
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  create_time timestamp with time zone NOT NULL DEFAULT current_timestamp,
  last_seen timestamp with time zone NOT NULL DEFAULT current_timestamp,
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

func (p *pgV0) UpdateUserActor() string {
	return `UPDATE ` + p.schema + `users SET actor = $2 WHERE id = $1`
}

func (p *pgV0) SensitiveUserByEmail() string {
	return "SELECT id, hashpass, salt FROM " + p.schema + "users WHERE email = $1"
}

func (p *pgV0) UserByID() string {
	return "SELECT id, email, actor, privileges, preferences FROM " + p.schema + "users WHERE id = $1"
}

func (p *pgV0) UserByPreferredUsername() string {
	return "SELECT id, email, actor, privileges, preferences FROM " + p.schema + "users WHERE actor->'preferredUsername' ? $1"
}

func (p *pgV0) ActorIDForOutbox() string {
	return `SELECT actor->>'id' FROM ` + p.schema + `users
WHERE actor->'outbox' ? $1`
}

func (p *pgV0) ActorIDForInbox() string {
	return `SELECT actor->>'id' FROM ` + p.schema + `users
WHERE actor->'inbox' ? $1`
}

func (p *pgV0) UpdateUserPreferences() string {
	return `UPDATE ` + p.schema + `users SET preferences = $2 WHERE id = $1`
}

func (p *pgV0) UpdateUserPrivileges() string {
	return `UPDATE ` + p.schema + `users SET privileges = $2 WHERE id = $1`
}

func (p *pgV0) InstanceUser() string {
	return "SELECT id, email, actor, privileges, preferences FROM " + p.schema + "users WHERE privileges->>'InstanceActor' = 'true'"
}

func (p *pgV0) GetInstanceActorPreferences() string {
	return `SELECT preferences
FROM ` + p.schema + `users
WHERE privileges->>'InstanceActor' = 'true'`
}

func (p *pgV0) SetInstanceActorPreferences() string {
	return `UPDATE ` + p.schema + `users
SET preferences = $1
WHERE privileges->>'InstanceActor' = 'true'`
}

func (p *pgV0) GetUserActivityStats() string {
	return `SELECT
  COUNT(*),
  COUNT(*) FILTER (WHERE current_timestamp - last_seen < '180 DAY'),
  COUNT(*) FILTER (WHERE current_timestamp - last_seen < '30 DAY'),
  COUNT(*) FILTER (WHERE current_timestamp - last_seen < '7 DAY')
FROM ` + p.schema + `users`
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

func (p *pgV0) CreateIndexIDFedDataTable() string {
	return `CREATE INDEX IF NOT EXISTS fed_data_id_index ON ` + p.schema + `fed_data USING GIN ((payload->'id'));`
}

func (p *pgV0) FedExists() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `fed_data
  WHERE payload->'id' ? $1
  LIMIT 1
)`
}

func (p *pgV0) FedGet() string {
	return `SELECT payload
FROM ` + p.schema + `fed_data
WHERE payload->'id' ? $1`
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

func (p *pgV0) CreateIndexIDLocalDataTable() string {
	return `CREATE INDEX IF NOT EXISTS local_data_id_index ON ` + p.schema + `local_data USING GIN ((payload->'id'));`
}

func (p *pgV0) LocalExists() string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + `local_data
  WHERE payload->'id' ? $1
  LIMIT 1
)`
}

func (p *pgV0) LocalGet() string {
	return `SELECT payload
FROM ` + p.schema + `local_data
WHERE payload->'id' ? $1`
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

func (p *pgV0) LocalStats() string {
	return `SELECT
  COUNT(*) FILTER (WHERE (payload->'inReplyTo') IS NULL),
  COUNT(*) FILTER (WHERE (payload->'inReplyTo') IS NOT NULL)
FROM ` + p.schema + `local_data`
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

func (p *pgV0) CreateIndexIDInboxesTable() string {
	return `CREATE INDEX IF NOT EXISTS inboxes_id_index ON ` + p.schema + `inboxes USING GIN ((inbox->'id'));`
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

func (p *pgV0) CreateIndexIDOutboxesTable() string {
	return `CREATE INDEX IF NOT EXISTS outboxes_id_index ON ` + p.schema + `outboxes USING GIN ((outbox->'id'));`
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
	$3::jsonb)) AS page,
    $3::integer + 1 >= jsonb_path_query(inbox, '$.orderedItems.size()')::numeric AS isEnd
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
      'OrderedCollectionPage') AS page,
  isEnd
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
	$3::jsonb)) AS page,
    $3::integer + 1 >= jsonb_path_query(outbox, '$.orderedItems.size()')::numeric AS isEnd
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
      'OrderedCollectionPage') AS page,
  isEnd
  FROM page`
}

func (p *pgV0) GetPublicInbox() string {
	return `WITH inbox AS (
  SELECT inbox
  FROM ` + p.schema + `inboxes
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
  LEFT JOIN ` + p.schema + `fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN ` + p.schema + `local_data AS ld
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
	$3::jsonb)) AS page,
    $3::integer + 1 >= jsonb_path_query(jsonb_agg(i.page), '$.size()')::numeric AS isEnd
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
      'OrderedCollectionPage') AS page,
  op.isEnd
  FROM inbox AS i, only_public AS op`
}

func (p *pgV0) GetPublicOutbox() string {
	return `WITH outbox AS (
  SELECT outbox
  FROM ` + p.schema + `outboxes
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
  LEFT JOIN ` + p.schema + `fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN ` + p.schema + `local_data AS ld
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
	$3::jsonb)) AS page,
    $3::integer + 1 >= jsonb_path_query(jsonb_agg(i.page), '$.size()')::numeric AS isEnd
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
      'OrderedCollectionPage') AS page,
  op.isEnd
  FROM outbox AS i, only_public AS op`
}

func (p *pgV0) GetInboxLastPage() string {
	return `WITH stats AS (
  SELECT
    inbox,
    GREATEST(0,
      jsonb_path_query(inbox, '$.orderedItems.size()')::numeric - $2) AS startIndex
  FROM ` + p.schema + `inboxes
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
  FROM ` + p.schema + `outboxes
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
  FROM ` + p.schema + `inboxes
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
  LEFT JOIN ` + p.schema + `fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN ` + p.schema + `local_data AS ld
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
  FROM ` + p.schema + `outboxes
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
  LEFT JOIN ` + p.schema + `fed_data AS fd
  ON pd.page = fd.payload->'id'
  WHERE
    fd.payload->'to' ? 'https://www.w3.org/ns/activitystreams#Public'
    OR fd.payload->'cc' ? 'https://www.w3.org/ns/activitystreams#Public'
),
local_public AS (
  SELECT pd.page AS page
  FROM page_elements AS pd
  LEFT JOIN ` + p.schema + `local_data AS ld
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
	return `UPDATE ` + p.schema + `inboxes
SET inbox = inbox || jsonb_build_object(
  'orderedItems',
  jsonb_build_array($2::text) || COALESCE(inbox->'orderedItems', '[]'::jsonb),
  'totalItems',
  (COALESCE(inbox->>'totalItems','0')::int + 1)::text::jsonb)
WHERE inbox->'id' ? $1`
}

func (p *pgV0) PrependOutboxItem() string {
	return `UPDATE ` + p.schema + `outboxes
SET outbox = outbox || jsonb_build_object(
  'orderedItems',
  jsonb_build_array($2::text) || COALESCE(outbox->'orderedItems', '[]'::jsonb),
  'totalItems',
  (COALESCE(outbox->>'totalItems','0')::int + 1)::text::jsonb)
WHERE outbox->'id' ? $1`
}

func (p *pgV0) DeleteInboxItem() string {
	return `UPDATE ` + p.schema + `inboxes
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
	return `UPDATE ` + p.schema + `outboxes
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
  state text NOT NULL,
  n_attempts bigint NOT NULL,
  last_attempt timestamp with time zone DEFAULT current_timestamp
);`
}

func (p *pgV0) InsertAttempt() string {
	return `INSERT INTO ` + p.schema + `delivery_attempts (from_id, deliver_to, payload, state, n_attempts) VALUES ($1, $2, $3, $4, 0) RETURNING id`
}

func (p *pgV0) MarkSuccessfulAttempt() string {
	return `UPDATE ` + p.schema + `delivery_attempts
SET
  state = $2,
  n_attempts = n_attempts + 1,
  last_attempt = current_timestamp
WHERE id = $1`
}

func (p *pgV0) MarkFailedAttempt() string {
	return `UPDATE ` + p.schema + `delivery_attempts
SET
  state = $2,
  n_attempts = n_attempts + 1,
  last_attempt = current_timestamp
WHERE id = $1`
}

func (p *pgV0) MarkAbandonedAttempt() string {
	return `UPDATE ` + p.schema + `delivery_attempts
SET
  state = $2,
  n_attempts = n_attempts + 1,
  last_attempt = current_timestamp
WHERE id = $1`
}

func (p *pgV0) FirstPageRetryableFailures() string {
	return `SELECT id, from_id, deliver_to, payload, n_attempts, last_attempt
FROM ` + p.schema + `delivery_attempts
WHERE state = $1 AND create_time < $2
ORDER BY id DESC
LIMIT $3`
}

func (p *pgV0) NextPageRetryableFailures() string {
	return `SELECT id, from_id, deliver_to, payload, n_attempts, last_attempt
FROM ` + p.schema + `delivery_attempts
WHERE state = $1 AND create_time < $2 AND id < $4
ORDER BY id DESC
LIMIT $3`
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

func (p *pgV0) GetPrivateKeyForInstanceActor() string {
	return `SELECT
  pk.priv_key
FROM ` + p.schema + `private_keys AS pk
LEFT JOIN ` + p.schema + `users AS u
ON u.id = pk.user_id
WHERE u.privileges->>'InstanceActor' = 'true' AND purpose = $1`
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

/* Collection prototype queries */

func (p *pgV0) createCollectionTable(name string) string {
	return `
CREATE TABLE IF NOT EXISTS ` + p.schema + name + `
(
  id text PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id text NOT NULL,
  ` + name + ` jsonb NOT NULL
)`
}

func (p *pgV0) createCollectionIDIndex(name string) string {
	return `CREATE INDEX IF NOT EXISTS ` + name + `_id_index ON ` + p.schema + name + ` USING GIN ((` + name + `->'id'));`
}

func (p *pgV0) insertCollection(name string) string {
	return `INSERT INTO ` + p.schema + name + ` (actor_id, ` + name + `) VALUES ($1, $2)`
}

func (p *pgV0) collectionContainsForActor(name string) string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + name + `
  WHERE actor_id = $1 AND ` + name + `->'items' ? $2
  LIMIT 1
)`
}

func (p *pgV0) collectionContains(name string) string {
	return `SELECT EXISTS (
  SELECT 1
  FROM ` + p.schema + name + `
  WHERE ` + name + `->'id' ? $1 AND ` + name + `->'items' ? $2
  LIMIT 1
)`
}

func (p *pgV0) getCollection(name string) string {
	return `WITH page AS (
  SELECT
    ` + name + `,
    jsonb_path_query_array(
      ` + name + `,
      '$.items[$min to $max]',
      jsonb_build_object(
        'min',
	$2::jsonb,
        'max',
	$3::jsonb)) AS page,
    $3::integer + 1 >= jsonb_path_query(` + name + `, '$.items.size()')::numeric AS isEnd
  FROM ` + p.schema + name + `
  WHERE ` + name + `->'id' ? $1
)
SELECT
  ` + name + `||
    jsonb_build_object(
      'items',
      page,
      'totalItems',
      jsonb_path_query(page, '$.size()'),
      'type',
      'CollectionPage') AS page,
  isEnd
  FROM page`
}

func (p *pgV0) getCollectionLastPage(name string) string {
	return `WITH stats AS (
  SELECT
    ` + name + `,
    GREATEST(0,
      jsonb_path_query(` + name + `, '$.items.size()')::numeric - $2) AS startIndex
  FROM ` + p.schema + name + `
  WHERE ` + name + `->'id' ? $1
),
page AS (
  SELECT
    ` + name + `,
    startIndex,
    jsonb_path_query_array(
      ` + name + `,
      '$.items[$min to last]',
      jsonb_build_object(
        'min',
        startIndex)) AS page
  FROM stats
)
SELECT
  ` + name + `||
    jsonb_build_object(
    'items',
    page,
    'totalItems',
    jsonb_path_query(page, '$.size()'),
    'type',
    'CollectionPage') AS ` + name + `,
  startIndex
FROM page`
}

func (p *pgV0) prependCollectionItem(name string) string {
	return `UPDATE ` + p.schema + name + `
SET ` + name + ` = ` + name + ` || jsonb_build_object(
  'items',
  jsonb_build_array($2::text) || COALESCE(` + name + `->'items', '[]'::jsonb),
  'totalItems',
  (COALESCE(` + name + `->>'totalItems','0')::int + 1)::text::jsonb)
WHERE ` + name + `->'id' ? $1`
}

func (p *pgV0) deleteCollectionItem(name string) string {
	return `UPDATE ` + p.schema + name + `
SET ` + name + `= jsonb_set(
  ` + name + `,
  '{items}',
  (` + name + `->'items') - $2) ||
  jsonb_build_object(
  'totalItems',
  (COALESCE(` + name + `->>'totalItems','0')::int - 1)::text::jsonb)
WHERE ` + name + `->'id' ? $1`
}

func (p *pgV0) getAllCollectionForActor(name string) string {
	return `SELECT ` + name + `
FROM ` + p.schema + name + `
WHERE actor_id = $1`
}

/* Collections */

const (
	v0Followers = "followers"
	v0Following = "following"
	v0Liked     = "liked"
)

func (p *pgV0) CreateFollowersTable() string {
	return p.createCollectionTable(v0Followers)
}

func (p *pgV0) CreateIndexIDFollowersTable() string {
	return p.createCollectionIDIndex(v0Followers)
}

func (p *pgV0) InsertFollowers() string {
	return p.insertCollection(v0Followers)
}

func (p *pgV0) FollowersContainsForActor() string {
	return p.collectionContainsForActor(v0Followers)
}

func (p *pgV0) FollowersContains() string {
	return p.collectionContains(v0Followers)
}

func (p *pgV0) GetFollowers() string {
	return p.getCollection(v0Followers)
}

func (p *pgV0) GetFollowersLastPage() string {
	return p.getCollectionLastPage(v0Followers)
}

func (p *pgV0) PrependFollowersItem() string {
	return p.prependCollectionItem(v0Followers)
}

func (p *pgV0) DeleteFollowersItem() string {
	return p.deleteCollectionItem(v0Followers)
}

func (p *pgV0) GetAllFollowersForActor() string {
	return p.getAllCollectionForActor(v0Followers)
}

func (p *pgV0) CreateFollowingTable() string {
	return p.createCollectionTable(v0Following)
}

func (p *pgV0) CreateIndexIDFollowingTable() string {
	return p.createCollectionIDIndex(v0Following)
}

func (p *pgV0) InsertFollowing() string {
	return p.insertCollection(v0Following)
}

func (p *pgV0) FollowingContainsForActor() string {
	return p.collectionContainsForActor(v0Following)
}

func (p *pgV0) FollowingContains() string {
	return p.collectionContains(v0Following)
}

func (p *pgV0) GetFollowing() string {
	return p.getCollection(v0Following)
}

func (p *pgV0) GetFollowingLastPage() string {
	return p.getCollectionLastPage(v0Following)
}

func (p *pgV0) PrependFollowingItem() string {
	return p.prependCollectionItem(v0Following)
}

func (p *pgV0) DeleteFollowingItem() string {
	return p.deleteCollectionItem(v0Following)
}

func (p *pgV0) GetAllFollowingForActor() string {
	return p.getAllCollectionForActor(v0Following)
}

func (p *pgV0) CreateLikedTable() string {
	return p.createCollectionTable(v0Liked)
}

func (p *pgV0) CreateIndexIDLikedTable() string {
	return p.createCollectionIDIndex(v0Liked)
}

func (p *pgV0) InsertLiked() string {
	return p.insertCollection(v0Liked)
}

func (p *pgV0) LikedContainsForActor() string {
	return p.collectionContainsForActor(v0Liked)
}

func (p *pgV0) LikedContains() string {
	return p.collectionContains(v0Liked)
}

func (p *pgV0) GetLiked() string {
	return p.getCollection(v0Liked)
}

func (p *pgV0) GetLikedLastPage() string {
	return p.getCollectionLastPage(v0Liked)
}

func (p *pgV0) PrependLikedItem() string {
	return p.prependCollectionItem(v0Liked)
}

func (p *pgV0) DeleteLikedItem() string {
	return p.deleteCollectionItem(v0Liked)
}

func (p *pgV0) GetAllLikedForActor() string {
	return p.getAllCollectionForActor(v0Liked)
}

func (p *pgV0) CreatePoliciesTable() string {
	return `CREATE TABLE IF NOT EXISTS ` + p.schema + `policies
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id text NOT NULL,
  purpose text NOT NULL,
  policy jsonb NOT NULL
)`
}

func (p *pgV0) CreatePolicy() string {
	return `INSERT INTO ` + p.schema + `policies (actor_id, purpose, policy) VALUES ($1, $2, $3) RETURNING id`
}

func (p *pgV0) GetPoliciesForActor() string {
	return `SELECT id, purpose, policy FROM ` + p.schema + `policies WHERE actor_id = $1`
}

func (p *pgV0) GetPoliciesForActorAndPurpose() string {
	return `SELECT id, policy FROM ` + p.schema + `policies WHERE actor_id = $1 AND purpose = $2`
}

func (p *pgV0) CreateResolutionsTable() string {
	return `CREATE TABLE IF NOT EXISTS ` + p.schema + `resolutions
(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  policy_id uuid REFERENCES ` + p.schema + `policies(id) ON DELETE CASCADE NOT NULL,
  data_iri text NOT NULL,
  resolution jsonb NOT NULL
)`
}

func (p *pgV0) CreateResolution() string {
	return `INSERT INTO ` + p.schema + `resolutions (policy_id, data_iri, resolution) VALUES ($1, $2, $3)`
}
