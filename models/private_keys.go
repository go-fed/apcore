package models

import (
	"database/sql"

	"github.com/go-fed/apcore/util"
)

var _ Model = &PrivateKeys{}

// PrivateKeys is a Model that provides additional database methods for the
// PrivateKey type.
type PrivateKeys struct {
	createPrivateKey *sql.Stmt
	getByUserID      *sql.Stmt
}

func (p *PrivateKeys) Prepare(db *sql.DB, s SqlDialect) error {
	var err error
	p.createPrivateKey, err = db.Prepare(s.CreatePrivateKey())
	if err != nil {
		return err
	}
	p.getByUserID, err = db.Prepare(s.GetPrivateKeyByUserID())
	if err != nil {
		return err
	}
	return nil
}

func (p *PrivateKeys) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreatePrivateKeysTable())
	return err
}

func (p *PrivateKeys) Close() {
	p.createPrivateKey.Close()
	p.getByUserID.Close()
}

// Create a new private key entry in the database.
func (p *PrivateKeys) Create(c util.Context, tx *sql.Tx, userID, purpose string, privKey []byte) error {
	_, err := tx.Stmt(p.createPrivateKey).ExecContext(c, userID, purpose, privKey)
	return err
}

// GetByUserID fetches a private key by the userID and purpose of the key.
func (p *PrivateKeys) GetByUserID(c util.Context, tx *sql.Tx, userID, purpose string) (b []byte, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(p.getByUserID).QueryContext(c, userID, purpose)
	if err != nil {
		return
	}
	defer rows.Close()
	return b, enforceOneRow(rows, "GetByUserID", func(r singleRow) error {
		return r.Scan(&(b))
	})
}
