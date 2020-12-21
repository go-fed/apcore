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

package services

import (
	"database/sql"

	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

type Any struct {
	DB      *sql.DB
	Dialect models.SqlDialect
}

func (a *Any) Begin() app.TxBuilder {
	return &txBuilder{db: a.DB, dialect: a.Dialect}
}

type txBuilder struct {
	db      *sql.DB
	dialect models.SqlDialect
	ops     []*anyOp
}

func (a *txBuilder) QueryOneRow(sql string, cb func(r app.SingleRow) error, args ...interface{}) {
	a.addOp(&anyOp{
		sql:      a.dialect.Apply(sql),
		args:     args,
		isExec:   false,
		isOneRow: true,
		cb:       cb,
	})
}

func (a *txBuilder) Query(sql string, cb func(r app.SingleRow) error, args ...interface{}) {
	a.addOp(&anyOp{
		sql:      a.dialect.Apply(sql),
		args:     args,
		isExec:   false,
		isOneRow: false,
		cb:       cb,
	})
}

func (a *txBuilder) ExecOneRow(sql string, args ...interface{}) {
	a.addOp(&anyOp{
		sql:      a.dialect.Apply(sql),
		args:     args,
		isExec:   true,
		isOneRow: true,
	})
}

func (a *txBuilder) Exec(sql string, args ...interface{}) {
	a.addOp(&anyOp{
		sql:      a.dialect.Apply(sql),
		args:     args,
		isExec:   true,
		isOneRow: false,
	})
}

func (a *txBuilder) addOp(op *anyOp) {
	a.ops = append(a.ops, op)
}

func (a *txBuilder) Do(c util.Context) error {
	return doInTx(c, a.db, func(tx *sql.Tx) error {
		for _, op := range a.ops {
			if err := op.Do(c, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

type anyOp struct {
	sql      string
	args     []interface{}
	isExec   bool
	isOneRow bool
	cb       func(r app.SingleRow) error
}

func (a *anyOp) Do(c util.Context, tx *sql.Tx) (err error) {
	if a.isExec {
		return a.doExec(c, tx)
	} else {
		return a.doQuery(c, tx)
	}
}

func (a *anyOp) doQuery(c util.Context, tx *sql.Tx) error {
	r, err := tx.QueryContext(c, a.sql, a.args...)
	if err != nil {
		return err
	}
	if a.isOneRow {
		return models.MustQueryOneRow(r, func(r models.SingleRow) error {
			return a.cb(r)
		})
	} else {
		return models.QueryRows(r, func(r models.SingleRow) error {
			return a.cb(r)
		})
	}
}

func (a *anyOp) doExec(c util.Context, tx *sql.Tx) error {
	r, err := tx.ExecContext(c, a.sql, a.args...)
	if err != nil {
		return err
	}
	if a.isOneRow {
		return models.MustChangeOneRow(r)
	}
	return nil
}
