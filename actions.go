package apcore

import (
	"context"

	"github.com/go-fed/apcore/app"
)

func doCreateTables(configFilePath string, a app.Application, debug bool, scheme string) error {
	db, d, ms, err := newModels(configFilePath, a, debug, scheme)
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, m := range ms {
		if err := m.CreateTable(tx, d); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func doInitAdmin(configFilePath string, a app.Application, debug bool) error {
	// TODO
	return nil
}
