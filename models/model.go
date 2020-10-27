package models

import (
	"database/sql"
)

// Model handles managing a single database type.
type Model interface {
	Prepare(*sql.DB, SqlDialect) error
	CreateTable(*sql.Tx, SqlDialect) error
	Close()
}

// stmtPair make a pair of **sql.Stmt and its associated SQL string.
//
// The goal is to populate *stmt based on the associated sqlStr.
type stmtPair struct {
	stmt   **sql.Stmt
	sqlStr string
}

// prepareStmtPair is a mapper that populates the stmtPair.stmt.
func prepareStmtPair(db *sql.DB, s stmtPair) (err error) {
	*s.stmt, err = db.Prepare(s.sqlStr)
	return err
}

// stmtPairs are a list of stmtPair.
type stmtPairs []stmtPair

// prepareStmtPairs turns stmtPairs into a single error, with a side effect of
// populating all stmt.
// TODO: Use this in more models
func prepareStmtPairs(db *sql.DB, s stmtPairs) (err error) {
	doIfNoErr := func(p stmtPair, fn func(*sql.DB, stmtPair) error) error {
		if err == nil {
			return fn(db, p)
		}
		return err
	}
	for _, p := range s {
		err = doIfNoErr(p, prepareStmtPair)
	}
	return
}
