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

var _ Model = &Liked{}

// Liked is a Model that provides additional database methods for Liked.
type Liked struct {
	insert           *sql.Stmt
	containsForActor *sql.Stmt
	contains         *sql.Stmt
	get              *sql.Stmt
	getLastPage      *sql.Stmt
	prependItem      *sql.Stmt
	deleteItem       *sql.Stmt
	getAllForActor   *sql.Stmt
}

func (i *Liked) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(i.insert), s.InsertLiked()},
			{&(i.containsForActor), s.LikedContainsForActor()},
			{&(i.contains), s.LikedContains()},
			{&(i.get), s.GetLiked()},
			{&(i.getLastPage), s.GetLikedLastPage()},
			{&(i.prependItem), s.PrependLikedItem()},
			{&(i.deleteItem), s.DeleteLikedItem()},
			{&(i.getAllForActor), s.GetAllLikedForActor()},
		})
}

func (i *Liked) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateLikedTable())
	return err
}

func (i *Liked) Close() {
	i.insert.Close()
	i.containsForActor.Close()
	i.contains.Close()
	i.get.Close()
	i.getLastPage.Close()
	i.prependItem.Close()
	i.deleteItem.Close()
	i.getAllForActor.Close()
}

// Create a new liked entry for the given actor.
func (i *Liked) Create(c util.Context, tx *sql.Tx, actor *url.URL, liked ActivityStreamsCollection) error {
	r, err := tx.Stmt(i.insert).ExecContext(c,
		actor.String(),
		liked)
	return mustChangeOneRow(r, err, "Liked.Create")
}

// ContainsForActor returns true if the item is in the actor's liked's collection.
func (i *Liked) ContainsForActor(c util.Context, tx *sql.Tx, actor, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.containsForActor).QueryContext(c, actor.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Liked.ContainsForActor", func(r singleRow) error {
		return r.Scan(&b)
	})
}

// Contains returns true if the item is in the liked's collection.
func (i *Liked) Contains(c util.Context, tx *sql.Tx, liked, item *url.URL) (b bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.contains).QueryContext(c, liked.String(), item.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "Liked.Contains", func(r singleRow) error {
		return r.Scan(&b)
	})
}

// GetPage returns a CollectionPage of the Liked.
//
// The range of elements retrieved are [min, max).
func (i *Liked) GetPage(c util.Context, tx *sql.Tx, liked *url.URL, min, max int) (page ActivityStreamsCollectionPage, isEnd bool, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.get).QueryContext(c, liked.String(), min, max-1)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, isEnd, enforceOneRow(rows, "Liked.GetPage", func(r singleRow) error {
		return r.Scan(&page, &isEnd)
	})
}

// GetLastPage returns the last CollectionPage of the Liked collection.
func (i *Liked) GetLastPage(c util.Context, tx *sql.Tx, liked *url.URL, n int) (page ActivityStreamsCollectionPage, startIdx int, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getLastPage).QueryContext(c, liked.String(), n)
	if err != nil {
		return
	}
	defer rows.Close()
	return page, startIdx, enforceOneRow(rows, "Liked.GetLastPage", func(r singleRow) error {
		return r.Scan(&page, &startIdx)
	})
}

// PrependItem prepends the item to the liked's ordered items list.
func (i *Liked) PrependItem(c util.Context, tx *sql.Tx, liked, item *url.URL) error {
	r, err := tx.Stmt(i.prependItem).ExecContext(c, liked.String(), item.String())
	return mustChangeOneRow(r, err, "Liked.PrependItem")
}

// DeleteItem removes the item from the liked's ordered items list.
func (i *Liked) DeleteItem(c util.Context, tx *sql.Tx, liked, item *url.URL) error {
	r, err := tx.Stmt(i.deleteItem).ExecContext(c, liked.String(), item.String())
	return mustChangeOneRow(r, err, "Liked.DeleteItem")
}

// GetAllForActor returns the entire Liked Collection.
func (i *Liked) GetAllForActor(c util.Context, tx *sql.Tx, liked *url.URL) (col ActivityStreamsCollection, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(i.getAllForActor).QueryContext(c, liked.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return col, enforceOneRow(rows, "Liked.GetAllForActor", func(r singleRow) error {
		return r.Scan(&col)
	})
}
