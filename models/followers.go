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

var _ Model = &Followers{}

// Followers is a Model that provides additional database methods for Followers.
type Followers struct {
	insert           *sql.Stmt
	containsForActor *sql.Stmt
	contains         *sql.Stmt
	get              *sql.Stmt
	getLastPage      *sql.Stmt
	prependItem      *sql.Stmt
	deleteItem       *sql.Stmt
	getAllForActor   *sql.Stmt
}

func (i *Followers) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(i.insert), s.InsertFollowers()},
			{&(i.containsForActor), s.FollowersContainsForActor()},
			{&(i.contains), s.FollowersContains()},
			{&(i.get), s.GetFollowers()},
			{&(i.getLastPage), s.GetFollowersLastPage()},
			{&(i.prependItem), s.PrependFollowersItem()},
			{&(i.deleteItem), s.DeleteFollowersItem()},
			{&(i.getAllForActor), s.GetAllFollowersForActor()},
		})
}

func (i *Followers) CreateTable(t *sql.Tx, s SqlDialect) error {
	if _, err := t.Exec(s.CreateFollowersTable()); err != nil {
		return err
	}
	_, err := t.Exec(s.CreateIndexIDFollowersTable())
	return err
}

func (i *Followers) Close() {
	i.insert.Close()
	i.containsForActor.Close()
	i.contains.Close()
	i.get.Close()
	i.getLastPage.Close()
	i.prependItem.Close()
	i.deleteItem.Close()
	i.getAllForActor.Close()
}

// Create a new followers for the given actor.
func (i *Followers) Create(c util.Context, tx *sql.Tx, actor *url.URL, followers ActivityStreamsCollection) error {
	r, err := tx.Stmt(i.insert).ExecContext(c,
		actor.String(),
		followers)
	return mustChangeOneRow(r, err, "Followers.Create")
}

// ContainsForActor returns true if the item is in the actor's followers's collection.
func (i *Followers) ContainsForActor(c util.Context, tx *sql.Tx, actor, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.containsForActor).QueryContext(c, actor.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Followers.ContainsForActor", func(r SingleRow) error {
		return r.Scan(&b)
	})
}

// Contains returns true if the item is in the followers's collection.
func (i *Followers) Contains(c util.Context, tx *sql.Tx, followers, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.contains).QueryContext(c, followers.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Followers.Contains", func(r SingleRow) error {
		return r.Scan(&b)
	})
}

// GetPage returns a CollectionPage of the Followers.
//
// The range of elements retrieved are [min, max).
func (i *Followers) GetPage(c util.Context, tx *sql.Tx, followers *url.URL, min, max int) (page ActivityStreamsCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.get).QueryContext(c, followers.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Followers.GetPage", func(r SingleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetLastPage returns the last CollectionPage of the Followers.
func (i *Followers) GetLastPage(c util.Context, tx *sql.Tx, followers *url.URL, n int) (page ActivityStreamsCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getLastPage).QueryContext(c, followers.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Followers.GetLastPage", func(r SingleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// PrependItem prepends the item to the followers' ordered items list.
func (i *Followers) PrependItem(c util.Context, tx *sql.Tx, followers, item *url.URL) error {
	r, err := tx.Stmt(i.prependItem).ExecContext(c, followers.String(), item.String())
	return mustChangeOneRow(r, err, "Followers.PrependItem")
}

// DeleteItem removes the item from the followers' ordered items list.
func (i *Followers) DeleteItem(c util.Context, tx *sql.Tx, followers, item *url.URL) error {
	r, err := tx.Stmt(i.deleteItem).ExecContext(c, followers.String(), item.String())
	return mustChangeOneRow(r, err, "Followers.DeleteItem")
}

// GetAllForActor returns the entire Collection of the Followers.
func (i *Followers) GetAllForActor(c util.Context, tx *sql.Tx, followers *url.URL) (col ActivityStreamsCollection, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getAllForActor).QueryContext(c, followers.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return col, enforceOneRow(rows, "Followers.GetAllForActor", func(r SingleRow) error {
		return r.Scan(&col)
	})
}
