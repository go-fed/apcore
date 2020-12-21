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

package models

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/apcore/util"
)

var _ Model = &Following{}

// Following is a Model that provides additional database methods for Following.
type Following struct {
	insert           *sql.Stmt
	containsForActor *sql.Stmt
	contains         *sql.Stmt
	get              *sql.Stmt
	getLastPage      *sql.Stmt
	prependItem      *sql.Stmt
	deleteItem       *sql.Stmt
	getAllForActor   *sql.Stmt
}

func (i *Following) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(i.insert), s.InsertFollowing()},
			{&(i.containsForActor), s.FollowingContainsForActor()},
			{&(i.contains), s.FollowingContains()},
			{&(i.get), s.GetFollowing()},
			{&(i.getLastPage), s.GetFollowingLastPage()},
			{&(i.prependItem), s.PrependFollowingItem()},
			{&(i.deleteItem), s.DeleteFollowingItem()},
			{&(i.getAllForActor), s.GetAllFollowingForActor()},
		})
}

func (i *Following) CreateTable(t *sql.Tx, s SqlDialect) error {
	if _, err := t.Exec(s.CreateFollowingTable()); err != nil {
		return err
	}
	_, err := t.Exec(s.CreateIndexIDFollowingTable())
	return err
}

func (i *Following) Close() {
	i.insert.Close()
	i.containsForActor.Close()
	i.contains.Close()
	i.get.Close()
	i.getLastPage.Close()
	i.prependItem.Close()
	i.deleteItem.Close()
	i.getAllForActor.Close()
}

// Create a new following entry for the given actor.
func (i *Following) Create(c util.Context, tx *sql.Tx, actor *url.URL, following ActivityStreamsCollection) error {
	r, err := tx.Stmt(i.insert).ExecContext(c,
		actor.String(),
		following)
	return mustChangeOneRow(r, err, "Following.Create")
}

// ContainsForActor returns true if the item is in the actor's following's collection.
func (i *Following) ContainsForActor(c util.Context, tx *sql.Tx, actor, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.containsForActor).QueryContext(c, actor.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Following.ContainsForActor", func(r SingleRow) error {
		return r.Scan(&b)
	})
}

// Contains returns true if the item is in the following's collection.
func (i *Following) Contains(c util.Context, tx *sql.Tx, following, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.contains).QueryContext(c, following.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Following.Contains", func(r SingleRow) error {
		return r.Scan(&b)
	})
}

// GetPage returns a CollectionPage of the Following.
//
// The range of elements retrieved are [min, max).
func (i *Following) GetPage(c util.Context, tx *sql.Tx, following *url.URL, min, max int) (page ActivityStreamsCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.get).QueryContext(c, following.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Following.GetPage", func(r SingleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetLastPage returns the last CollectionPage of the Following collection.
func (i *Following) GetLastPage(c util.Context, tx *sql.Tx, following *url.URL, n int) (page ActivityStreamsCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getLastPage).QueryContext(c, following.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Following.GetLastPage", func(r SingleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// PrependItem prepends the item to the following's ordered items list.
func (i *Following) PrependItem(c util.Context, tx *sql.Tx, following, item *url.URL) error {
	r, err := tx.Stmt(i.prependItem).ExecContext(c, following.String(), item.String())
	return mustChangeOneRow(r, err, "Following.PrependItem")
}

// DeleteItem removes the item from the following's ordered items list.
func (i *Following) DeleteItem(c util.Context, tx *sql.Tx, following, item *url.URL) error {
	r, err := tx.Stmt(i.deleteItem).ExecContext(c, following.String(), item.String())
	return mustChangeOneRow(r, err, "Following.DeleteItem")
}

// GetAllForActor returns the entire Following Collection.
func (i *Following) GetAllForActor(c util.Context, tx *sql.Tx, following *url.URL) (col ActivityStreamsCollection, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getAllForActor).QueryContext(c, following.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return col, enforceOneRow(rows, "Following.GetAllForActor", func(r SingleRow) error {
		return r.Scan(&col)
	})
}
