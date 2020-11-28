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

package models

import (
	"database/sql"
	"database/sql/driver"
	"net/url"

	"github.com/go-fed/apcore/util"
)

type CreateUser struct {
	Email       string
	Hashpass    []byte
	Salt        []byte
	Actor       driver.Valuer
	Privileges  Privileges
	Preferences Preferences
}

type User struct {
	ID          string
	Email       string
	Actor       ActivityStreams
	Privileges  Privileges
	Preferences Preferences
}

type SensitiveUser struct {
	ID       string
	Hashpass []byte
	Salt     []byte
}

var _ Model = &Users{}

// Users is a Model that provides additional database methods for the
// Users type.
type Users struct {
	insertUser              *sql.Stmt
	updateActor             *sql.Stmt
	sensitiveUserByEmail    *sql.Stmt
	userByID                *sql.Stmt
	userByPreferredUsername *sql.Stmt
	actorIDForOutbox        *sql.Stmt
	actorIDForInbox         *sql.Stmt
	updatePreferences       *sql.Stmt
	updatePrivileges        *sql.Stmt
	instanceUser            *sql.Stmt
}

func (u *Users) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(u.insertUser), s.InsertUser()},
			{&(u.updateActor), s.UpdateUserActor()},
			{&(u.sensitiveUserByEmail), s.SensitiveUserByEmail()},
			{&(u.userByID), s.UserByID()},
			{&(u.userByPreferredUsername), s.UserByPreferredUsername()},
			{&(u.actorIDForOutbox), s.ActorIDForOutbox()},
			{&(u.actorIDForInbox), s.ActorIDForInbox()},
			{&(u.updatePreferences), s.UpdateUserPreferences()},
			{&(u.updatePrivileges), s.UpdateUserPrivileges()},
			{&(u.instanceUser), s.InstanceUser()},
		})
}

func (u *Users) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateUsersTable())
	return err
}

func (u *Users) Close() {
	u.insertUser.Close()
	u.updateActor.Close()
	u.sensitiveUserByEmail.Close()
	u.userByID.Close()
	u.userByPreferredUsername.Close()
	u.actorIDForOutbox.Close()
	u.actorIDForInbox.Close()
	u.updatePreferences.Close()
	u.updatePrivileges.Close()
	u.instanceUser.Close()
}

// Create a User in the database.
func (u *Users) Create(c util.Context, tx *sql.Tx, r *CreateUser) (userID string, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.insertUser).QueryContext(c,
		r.Email,
		r.Hashpass,
		r.Salt,
		r.Actor,
		r.Privileges,
		r.Preferences)
	if err != nil {
		return
	}
	defer rows.Close()
	return userID, enforceOneRow(rows, "Users.Create", func(r singleRow) error {
		return r.Scan(&(userID))
	})
}

// UpdateActor updates the Actor for the userID.
func (u *Users) UpdateActor(c util.Context, tx *sql.Tx, id string, actor ActivityStreams) error {
	r, err := tx.Stmt(u.updateActor).ExecContext(c, id, actor)
	return mustChangeOneRow(r, err, "Users.UpdateActor")
}

// SensitiveUserByEmail returns the credentials for a given user's email.
func (u *Users) SensitiveUserByEmail(c util.Context, tx *sql.Tx, email string) (s *SensitiveUser, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.sensitiveUserByEmail).QueryContext(c, email)
	if err != nil {
		return
	}
	defer rows.Close()
	return s, enforceOneRow(rows, "SensitiveUserByEmail", func(r singleRow) error {
		s = &SensitiveUser{}
		return r.Scan(&(s.ID), &(s.Hashpass), &(s.Salt))
	})
}

// UserByID returns the non-sensitive fields for a User for a given ID.
func (u *Users) UserByID(c util.Context, tx *sql.Tx, id string) (s *User, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.userByID).QueryContext(c, id)
	if err != nil {
		return
	}
	defer rows.Close()
	return s, enforceOneRow(rows, "UserByID", func(r singleRow) error {
		s = &User{}
		return r.Scan(&(s.ID), &(s.Email), &(s.Actor), &(s.Privileges), &(s.Preferences))
	})
}

// UserByPreferredUsername returns the non-sensitive fields for a User for a
// given preferredUsername.
func (u *Users) UserByPreferredUsername(c util.Context, tx *sql.Tx, name string) (s *User, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.userByPreferredUsername).QueryContext(c, name)
	if err != nil {
		return
	}
	defer rows.Close()
	return s, enforceOneRow(rows, "UserByID", func(r singleRow) error {
		s = &User{}
		return r.Scan(&(s.ID), &(s.Email), &(s.Actor), &(s.Privileges), &(s.Preferences))
	})
}

// InstanceActorUser returns the user representing the instance.
func (u *Users) InstanceActorUser(c util.Context, tx *sql.Tx) (s *User, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.instanceUser).QueryContext(c)
	if err != nil {
		return
	}
	defer rows.Close()
	return s, enforceOneRow(rows, "Users.InstanceActorUser", func(r singleRow) error {
		s = &User{}
		return r.Scan(&(s.ID), &(s.Email), &(s.Actor), &(s.Privileges), &(s.Preferences))
	})
}

// ActorIDForOutbox returns the actor associated with the outbox.
func (u *Users) ActorIDForOutbox(c util.Context, tx *sql.Tx, outbox *url.URL) (actor URL, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.actorIDForOutbox).QueryContext(c, outbox.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return actor, enforceOneRow(rows, "Users.ActorIDForOutbox", func(r singleRow) error {
		return r.Scan(&actor)
	})
}

// ActorIDForOutbox returns the actor associated with the inbox.
func (u *Users) ActorIDForInbox(c util.Context, tx *sql.Tx, inbox *url.URL) (actor URL, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(u.actorIDForInbox).QueryContext(c, inbox.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return actor, enforceOneRow(rows, "Users.ActorIDForInbox", func(r singleRow) error {
		return r.Scan(&actor)
	})
}

// UpdatePreferences updates the preferences associated with the user.
func (u *Users) UpdatePreferences(c util.Context, tx *sql.Tx, id string, p Preferences) error {
	r, err := tx.Stmt(u.updatePreferences).ExecContext(c,
		id,
		p)
	return mustChangeOneRow(r, err, "Users.UpdatePreferences")
}

// UpdatePrivileges updates the privileges associated with the user.
func (u *Users) UpdatePrivileges(c util.Context, tx *sql.Tx, id string, p Privileges) error {
	r, err := tx.Stmt(u.updatePrivileges).ExecContext(c,
		id,
		p)
	return mustChangeOneRow(r, err, "Users.UpdatePrivileges")
}
