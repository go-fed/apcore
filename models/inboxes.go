package models

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/apcore/util"
)

var _ Model = &Inboxes{}

// Inboxes is a Model that provides additional database methods for Inboxes.
type Inboxes struct {
	insertInbox           *sql.Stmt
	inboxContainsForActor *sql.Stmt
	inboxContains         *sql.Stmt
	getInbox              *sql.Stmt
	getPublicInbox        *sql.Stmt
	getLastPage           *sql.Stmt
	getPublicLastPage     *sql.Stmt
	prependInboxItem      *sql.Stmt
	deleteInboxItem       *sql.Stmt
}

func (i *Inboxes) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(i.insertInbox), s.InsertInbox()},
			{&(i.inboxContainsForActor), s.InboxContainsForActor()},
			{&(i.inboxContains), s.InboxContains()},
			{&(i.getInbox), s.GetInbox()},
			{&(i.getPublicInbox), s.GetPublicInbox()},
			{&(i.getLastPage), s.GetInboxLastPage()},
			{&(i.getPublicLastPage), s.GetPublicInboxLastPage()},
			{&(i.prependInboxItem), s.PrependInboxItem()},
			{&(i.deleteInboxItem), s.DeleteInboxItem()},
		})
}

func (i *Inboxes) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateInboxesTable())
	return err
}

func (i *Inboxes) Close() {
	i.insertInbox.Close()
	i.inboxContainsForActor.Close()
	i.inboxContains.Close()
	i.getInbox.Close()
	i.getPublicInbox.Close()
	i.getLastPage.Close()
	i.getPublicLastPage.Close()
	i.prependInboxItem.Close()
	i.deleteInboxItem.Close()
}

// Create a new inbox for the given actor.
func (i *Inboxes) Create(c util.Context, tx *sql.Tx, actor *url.URL, inbox ActivityStreamsOrderedCollection) error {
	_, err := tx.Stmt(i.insertInbox).ExecContext(c,
		actor.String(),
		inbox)
	return err
}

// ContainsForActor returns true if the item is in the actor's inbox's collection.
func (i *Inboxes) ContainsForActor(c util.Context, tx *sql.Tx, actor, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.inboxContainsForActor).QueryContext(c, actor.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Inboxes.ContainsForActor", func(r singleRow) error {
		return r.Scan(&b)
	})
}

// Contains returns true if the item is in the inbox's collection.
func (i *Inboxes) Contains(c util.Context, tx *sql.Tx, inbox, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.inboxContains).QueryContext(c, inbox.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Inboxes.Contains", func(r singleRow) error {
		return r.Scan(&b)
	})
}

// GetPage returns an OrderedCollectionPage of the Inbox.
//
// The range of elements retrieved are [min, max).
func (i *Inboxes) GetPage(c util.Context, tx *sql.Tx, inbox *url.URL, min, max int) (page ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getInbox).QueryContext(c, inbox.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Inboxes.GetPage", func(r singleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetPublicPage returns an OrderedCollectionPage of inbox items that are
// public only.
//
// The range of elements retrieved are [min, max).
func (i *Inboxes) GetPublicPage(c util.Context, tx *sql.Tx, inbox *url.URL, min, max int) (page ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getPublicInbox).QueryContext(c, inbox.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Inboxes.GetPublicPage", func(r singleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetLastPage returns the last OrderedCollectionPage of the Inbox.
func (i *Inboxes) GetLastPage(c util.Context, tx *sql.Tx, inbox *url.URL, n int) (page ActivityStreamsOrderedCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getLastPage).QueryContext(c, inbox.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Inboxes.GetLastPage", func(r singleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// GetPublicLastPage returns the last OrderedCollectionPage of inbox items that
// are public only.
func (i *Inboxes) GetPublicLastPage(c util.Context, tx *sql.Tx, inbox *url.URL, n int) (page ActivityStreamsOrderedCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getPublicLastPage).QueryContext(c, inbox.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Inboxes.GetPublicLastPage", func(r singleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// PrependInboxItem prepends the item to the inbox's ordered items list.
func (i *Inboxes) PrependInboxItem(c util.Context, tx *sql.Tx, inbox, item *url.URL) error {
	_, err := tx.Stmt(i.prependInboxItem).ExecContext(c, inbox.String(), item.String())
	return err
}

// DeleteInboxItem removes the item from the inbox's ordered items list.
func (i *Inboxes) DeleteInboxItem(c util.Context, tx *sql.Tx, inbox, item *url.URL) error {
	_, err := tx.Stmt(i.deleteInboxItem).ExecContext(c, inbox.String(), item.String())
	return err
}
