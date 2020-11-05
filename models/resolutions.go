package models

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/apcore/util"
)

type Resolution struct {
	Time time.Time `json:"time",omitempty`

	// The following are used by Policies
	Matched  bool     `json:"matched",omitempty`
	MatchLog []string `json:"matchLog",omitempty`
}

func (r *Resolution) Logf(s string, i ...interface{}) {
	r.Log(fmt.Sprintf(s, i...))
}

func (r *Resolution) Log(s string) {
	r.MatchLog = append(r.MatchLog, s)
}

type CreateResolution struct {
	PolicyID string
	IRI      *url.URL
	R        Resolution
}

var _ Model = &Resolutions{}

// Resolutions is a Model that provides additional database methods for the
// Resolution type.
type Resolutions struct {
	create *sql.Stmt
}

func (r *Resolutions) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(r.create), s.CreateResolution()},
		})
}

func (r *Resolutions) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreateResolutionsTable())
	return err
}

func (r *Resolutions) Close() {
	r.create.Close()
}

// Create a new Resolution
func (r *Resolutions) Create(c util.Context, tx *sql.Tx, cr CreateResolution) error {
	rows, err := tx.Stmt(r.create).ExecContext(c,
		cr.PolicyID,
		cr.IRI.String(),
		cr.R)
	return mustChangeOneRow(rows, err, "Resolutions.Create")
}
