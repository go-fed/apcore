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
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/activity/streams/vocab"
	_ "github.com/lib/pq"
)

type Database interface{}

var _ Database = &database{}

type database struct {
	db *sql.DB
	// url.URL.Host name for this server
	hostname string
	// Prepared statements for the database
	inboxContains  *sql.Stmt
	getInbox       *sql.Stmt
	setInbox       *sql.Stmt
	actorForOutbox *sql.Stmt
	actorForInbox  *sql.Stmt
	outboxForInbox *sql.Stmt
	exists         *sql.Stmt
	get            *sql.Stmt
	create         *sql.Stmt
	update         *sql.Stmt
	deleteStmt     *sql.Stmt
	getOutbox      *sql.Stmt
	setOutbox      *sql.Stmt
	followers      *sql.Stmt
	following      *sql.Stmt
	liked          *sql.Stmt
}

func newDatabase(c *config, a Application, debug bool) (db *database, err error) {
	kind := c.DatabaseConfig.DatabaseKind
	var conn string
	var sqlgen sqlGenerator
	switch kind {
	case "postgres":
		sqlgen = newPgV0(c.DatabaseConfig.PostgresConfig.Schema, debug)
		conn, err = postgresConn(c.DatabaseConfig.PostgresConfig)
	default:
		err = fmt.Errorf("unhandled database_kind in config: %s", kind)
	}
	if err != nil {
		return
	}

	InfoLogger.Infof("Opening database")
	var sqldb *sql.DB
	sqldb, err = sql.Open(kind, conn)
	if err != nil {
		return
	}
	InfoLogger.Infof("Open complete (note connection may not yet be attempted until first SQL command issued)")

	// Apply general database configurations
	if c.DatabaseConfig.ConnMaxLifetimeSeconds > 0 {
		sqldb.SetConnMaxLifetime(
			time.Duration(c.DatabaseConfig.ConnMaxLifetimeSeconds) *
				time.Second)
	}
	if c.DatabaseConfig.MaxOpenConns > 0 {
		sqldb.SetMaxOpenConns(c.DatabaseConfig.MaxOpenConns)
	}
	if c.DatabaseConfig.MaxIdleConns >= 0 {
		sqldb.SetMaxIdleConns(c.DatabaseConfig.MaxIdleConns)
	}

	db = &database{
		db: sqldb,
		// TODO: hostname
	}
	db.inboxContains, err = db.db.Prepare(sqlgen.InboxContains())
	if err != nil {
		return
	}
	db.getInbox, err = db.db.Prepare(sqlgen.GetInbox())
	if err != nil {
		return
	}
	db.setInbox, err = db.db.Prepare(sqlgen.SetInbox())
	if err != nil {
		return
	}
	db.actorForOutbox, err = db.db.Prepare(sqlgen.ActorForOutbox())
	if err != nil {
		return
	}
	db.actorForInbox, err = db.db.Prepare(sqlgen.ActorForInbox())
	if err != nil {
		return
	}
	db.outboxForInbox, err = db.db.Prepare(sqlgen.OutboxForInbox())
	if err != nil {
		return
	}
	db.exists, err = db.db.Prepare(sqlgen.Exists())
	if err != nil {
		return
	}
	db.get, err = db.db.Prepare(sqlgen.Get())
	if err != nil {
		return
	}
	db.create, err = db.db.Prepare(sqlgen.Create())
	if err != nil {
		return
	}
	db.update, err = db.db.Prepare(sqlgen.Update())
	if err != nil {
		return
	}
	db.deleteStmt, err = db.db.Prepare(sqlgen.Delete())
	if err != nil {
		return
	}
	db.getOutbox, err = db.db.Prepare(sqlgen.GetOutbox())
	if err != nil {
		return
	}
	db.setOutbox, err = db.db.Prepare(sqlgen.SetOutbox())
	if err != nil {
		return
	}
	db.followers, err = db.db.Prepare(sqlgen.Followers())
	if err != nil {
		return
	}
	db.following, err = db.db.Prepare(sqlgen.Following())
	if err != nil {
		return
	}
	db.liked, err = db.db.Prepare(sqlgen.Liked())
	if err != nil {
		return
	}
	return
}

func postgresConn(pg postgresConfig) (s string, err error) {
	InfoLogger.Info("Postgres database configuration")
	if len(pg.DatabaseName) == 0 {
		err = fmt.Errorf("postgres config missing db_name")
		return
	} else if len(pg.UserName) == 0 {
		err = fmt.Errorf("postgres config missing user")
		return
	}
	s = fmt.Sprintf("dbname=%s user=%s", pg.DatabaseName, pg.UserName)
	var hasPw bool
	hasPw, err = promptDoesXHavePassword(
		fmt.Sprintf(
			"user=%q in db_name=%q",
			pg.DatabaseName,
			pg.UserName))
	if err != nil {
		return
	}
	if hasPw {
		var pw string
		pw, err = promptPassword(
			fmt.Sprintf(
				"Please enter the password for db_name=%q and user=%q:",
				pg.DatabaseName,
				pg.UserName))
		if err != nil {
			return
		}
		s = fmt.Sprintf("%s password=%s", s, pw)
	}
	if len(pg.Host) > 0 {
		s = fmt.Sprintf("%s host=%s", s, pg.Host)
	}
	if pg.Port > 0 {
		s = fmt.Sprintf("%s port=%d", s, pg.Port)
	}
	if len(pg.SSLMode) > 0 {
		s = fmt.Sprintf("%s sslmode=%s", s, pg.SSLMode)
	}
	if len(pg.FallbackApplicationName) > 0 {
		s = fmt.Sprintf("%s fallback_application_name=%s", s, pg.FallbackApplicationName)
	}
	if pg.ConnectTimeout > 0 {
		s = fmt.Sprintf("%s connect_timeout=%d", s, pg.ConnectTimeout)
	}
	if len(pg.SSLCert) > 0 {
		s = fmt.Sprintf("%s sslcert=%s", s, pg.SSLCert)
	}
	if len(pg.SSLKey) > 0 {
		s = fmt.Sprintf("%s sslkey=%s", s, pg.SSLKey)
	}
	if len(pg.SSLRootCert) > 0 {
		s = fmt.Sprintf("%s sslrootcert=%s", s, pg.SSLRootCert)
	}
	return
}

func (d *database) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	var r *sql.Rows
	r, err = d.inboxContains.QueryContext(c, inbox.String(), id.String())
	if err != nil {
		return
	}
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking inbox containment")
			return
		}
		if err = r.Scan(&contains); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// TODO: Default length
	defaultLength := 10
	start := collectionPageStartIndex(inboxIRI)
	length := collectionPageLength(inboxIRI, defaultLength)
	var r *sql.Rows
	r, err = d.getInbox.QueryContext(c, inboxIRI.String(), start, length)
	if err != nil {
		return
	}
	var ids []string
	for r.Next() {
		var id string
		if err = r.Scan(&id); err != nil {
			return
		}
		ids = append(ids, id)
	}
	if err = r.Err(); err != nil {
		return
	}
	var id *url.URL
	id, err = collectionPageId(inboxIRI, start, length, defaultLength)
	if err != nil {
		return
	}
	inbox = toOrderedCollectionPage(id, ids, start, length)
	return
}

func (d *database) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// TODO
	return nil
}

func (d *database) Owns(c context.Context, id *url.URL) (owns bool, err error) {
	owns = id.Host == d.hostname
	return
}

func (d *database) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	var r *sql.Rows
	r, err = d.actorForOutbox.QueryContext(c, outboxIRI.String())
	if err != nil {
		return
	}
	var n int
	var iri string
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking actor for outbox")
			return
		}
		if err = r.Scan(&iri); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	actorIRI, err = url.Parse(iri)
	return
}

func (d *database) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	var r *sql.Rows
	r, err = d.actorForInbox.QueryContext(c, inboxIRI.String())
	if err != nil {
		return
	}
	var n int
	var iri string
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking actor for inbox")
			return
		}
		if err = r.Scan(&iri); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	actorIRI, err = url.Parse(iri)
	return
}

func (d *database) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	// TODO
	return
}

func (d *database) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	var r *sql.Rows
	r, err = d.exists.QueryContext(c, id.String())
	if err != nil {
		return
	}
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking exists")
			return
		}
		if err = r.Scan(&exists); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	// TODO
	return
}

func (d *database) Create(c context.Context, asType vocab.Type) error {
	// TODO
	return nil
}

func (d *database) Update(c context.Context, asType vocab.Type) error {
	// TODO
	return nil
}

func (d *database) Delete(c context.Context, id *url.URL) error {
	// TODO
	return nil
}

func (d *database) GetOutbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// TODO
	return
}

func (d *database) SetOutbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// TODO
	return nil
}

func (d *database) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	// TODO
	return
}

func (d *database) Following(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	// TODO
	return
}

func (d *database) Liked(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	// TODO
	return
}
