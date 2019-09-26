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
)

// sqlManager implements the necessary table definitions needed for its
// sqlGenerator.
type sqlManager interface {
	CreateTables(t *sql.Tx) (err error)
	UpgradeTables(t *sql.Tx) (err error)
}

// sqlGenerator is a SQL dialect provider.
type sqlGenerator interface {
	HashPassForUserID() string
	UserIdForEmail() string
	UserIdForBoxPath() string
	UserPreferences() string
	UpdateUserPolicy() string
	UpdateInstancePolicy() string
	InsertUserPolicy() string
	InsertInstancePolicy() string
	InstancePolicies() string
	UserPolicies() string
	InsertResolutions() string
	UserResolutions() string

	InsertUserPKey() string
	GetUserPKey() string
	FollowersByUserUUID() string

	InsertAttempt() string
	MarkSuccessfulAttempt() string
	MarkRetryFailureAttempt() string

	CreateTokenInfo() string
	RemoveTokenByCode() string
	RemoveTokenByAccess() string
	RemoveTokenByRefresh() string
	GetTokenByCode() string
	GetTokenByAccess() string
	GetTokenByRefresh() string
	GetClientById() string

	InboxContains() string
	GetInbox() string
	GetPublicInbox() string
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
	GetPublicOutbox() string
	SetOutboxUpdate() string
	SetOutboxInsert() string
	SetOutboxDelete() string
	Followers() string
	Following() string
	Liked() string
}
