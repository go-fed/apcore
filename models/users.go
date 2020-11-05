package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/url"

	"github.com/go-fed/apcore/util"
)

type CreateUser struct {
	Email       string
	Hashpass    []byte
	Salt        []byte
	Actor       ActivityStreamsPerson
	Privileges  Privileges
	Preferences Preferences
}

type User struct {
	ID          string
	Email       string
	Actor       ActivityStreamsPerson
	Privileges  Privileges
	Preferences Preferences
}

type SensitiveUser struct {
	ID       string
	Hashpass []byte
	Salt     []byte
}

var _ driver.Valuer = Privileges{}
var _ sql.Scanner = &Privileges{}

// TODO: Privileges
type Privileges struct{}

func (p Privileges) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Privileges) Scan(src interface{}) error {
	return unmarshal(src, p)
}

var _ driver.Valuer = Preferences{}
var _ sql.Scanner = &Preferences{}

// TODO: Preferences
type Preferences struct {
	OnFollow OnFollowBehavior
}

func (p Preferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Preferences) Scan(src interface{}) error {
	return unmarshal(src, p)
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
func (u *Users) UpdateActor(c util.Context, tx *sql.Tx, id string, actor ActivityStreamsPerson) error {
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
