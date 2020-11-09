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
	"net/url"

	"github.com/go-fed/apcore/util"
)

var _ Model = &Outboxes{}

// Outboxes is a Model that provides additional database methods for Outboxes.
type Outboxes struct {
	insertOutbox           *sql.Stmt
	outboxContainsForActor *sql.Stmt
	outboxContains         *sql.Stmt
	getOutbox              *sql.Stmt
	getPublicOutbox        *sql.Stmt
	getLastPage            *sql.Stmt
	getPublicLastPage      *sql.Stmt
	prependOutboxItem      *sql.Stmt
	deleteOutboxItem       *sql.Stmt
	outboxForInbox         *sql.Stmt
}

func (i *Outboxes) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(i.insertOutbox), s.InsertOutbox()},
			{&(i.outboxContainsForActor), s.OutboxContainsForActor()},
			{&(i.outboxContains), s.OutboxContains()},
			{&(i.getOutbox), s.GetOutbox()},
			{&(i.getPublicOutbox), s.GetPublicOutbox()},
			{&(i.getLastPage), s.GetOutboxLastPage()},
			{&(i.getPublicLastPage), s.GetPublicOutboxLastPage()},
			{&(i.prependOutboxItem), s.PrependOutboxItem()},
			{&(i.deleteOutboxItem), s.DeleteOutboxItem()},
			{&(i.outboxForInbox), s.OutboxForInbox()},
		})
}

func (i *Outboxes) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateOutboxesTable())
	return err
}

func (i *Outboxes) Close() {
	i.insertOutbox.Close()
	i.outboxContainsForActor.Close()
	i.outboxContains.Close()
	i.getOutbox.Close()
	i.getPublicOutbox.Close()
	i.getLastPage.Close()
	i.getPublicLastPage.Close()
	i.prependOutboxItem.Close()
	i.deleteOutboxItem.Close()
	i.outboxForInbox.Close()
}

// Create a new outbox for the given actor.
func (i *Outboxes) Create(c util.Context, tx *sql.Tx, actor *url.URL, outbox ActivityStreamsOrderedCollection) error {
	r, err := tx.Stmt(i.insertOutbox).ExecContext(c,
		actor.String(),
		outbox)
	return mustChangeOneRow(r, err, "Outboxes.Create")
}

// ContainsForActor returns true if the item is in the actor's outbox's collection.
func (i *Outboxes) ContainsForActor(c util.Context, tx *sql.Tx, actor, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.outboxContainsForActor).QueryContext(c, actor.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Outboxes.ContainsForActor", func(r singleRow) error {
		return r.Scan(&b)
	})
}

// Contains returns true if the item is in the outbox's collection.
func (i *Outboxes) Contains(c util.Context, tx *sql.Tx, inbox, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.outboxContains).QueryContext(c, inbox.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Outboxes.Contains", func(r singleRow) error {
		return r.Scan(&b)
	})
}

// GetPage returns an OrderedCollectionPage of the Outbox.
//
// The range of elements retrieved are [min, max).
func (i *Outboxes) GetPage(c util.Context, tx *sql.Tx, outbox *url.URL, min, max int) (page ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getOutbox).QueryContext(c, outbox.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Outboxes.GetPage", func(r singleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetPublicPage returns an OrderedCollectionPage of outbox items that are
// public only.
//
// The range of elements retrieved are [min, max).
func (i *Outboxes) GetPublicPage(c util.Context, tx *sql.Tx, outbox *url.URL, min, max int) (page ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getPublicOutbox).QueryContext(c, outbox.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Outboxes.GetPublicPage", func(r singleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetLastPage returns the last OrderedCollectionPage of the Outbox.
func (i *Outboxes) GetLastPage(c util.Context, tx *sql.Tx, outbox *url.URL, n int) (page ActivityStreamsOrderedCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getLastPage).QueryContext(c, outbox.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Outboxes.GetPage", func(r singleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// GetPublicLastPage returns the last OrderedCollectionPage of outbox items that
// are public only.
func (i *Outboxes) GetPublicLastPage(c util.Context, tx *sql.Tx, outbox *url.URL, n int) (page ActivityStreamsOrderedCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getPublicLastPage).QueryContext(c, outbox.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Outboxes.GetPublicPage", func(r singleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// PrependOutboxItem prepends the item to the outbox's ordered items list.
func (i *Outboxes) PrependOutboxItem(c util.Context, tx *sql.Tx, outbox, item *url.URL) error {
	r, err := tx.Stmt(i.prependOutboxItem).ExecContext(c, outbox.String(), item.String())
	return mustChangeOneRow(r, err, "Outboxes.PrependOutboxItem")
}

// DeleteOutboxItem removes the item from the outbox's ordered items list.
func (i *Outboxes) DeleteOutboxItem(c util.Context, tx *sql.Tx, outbox, item *url.URL) error {
	r, err := tx.Stmt(i.deleteOutboxItem).ExecContext(c, outbox.String(), item.String())
	return mustChangeOneRow(r, err, "Outboxes.DeleteOutboxItem")
}

// OutboxForInbox returns the outbox for the inbox.
func (i *Outboxes) OutboxForInbox(c util.Context, tx *sql.Tx, inbox *url.URL) (outbox URL, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.outboxForInbox).QueryContext(c, inbox.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return outbox, enforceOneRow(rows, "Outboxes.OutboxForInbox", func(r singleRow) error {
		return r.Scan(&outbox)
	})
}
