package apcore

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type database struct {
	db *sql.DB
}

func newDatabase(c *config, a Application) (db *database, err error) {
	kind := c.DatabaseConfig.DatabaseKind
	var conn string
	switch kind {
	case "postgres":
		conn, err = postgresConn(c.DatabaseConfig.PostgresConfig)
	default:
		err = fmt.Errorf("unhandled database_kind in config: %s", kind)
	}
	if err != nil {
		return
	}

	InfoLogger.Infof("Opening database")
	var sqldb *sql.DB
	sqldb, err = sql.Open(kind, conn)
	if err != nil {
		return
	}
	InfoLogger.Infof("Open complete (note connection may not yet be attempted until first SQL command issued)")

	// Apply general database configurations
	if c.DatabaseConfig.ConnMaxLifetimeSeconds > 0 {
		sqldb.SetConnMaxLifetime(
			time.Duration(c.DatabaseConfig.ConnMaxLifetimeSeconds) *
				time.Second)
	}
	if c.DatabaseConfig.MaxOpenConns > 0 {
		sqldb.SetMaxOpenConns(c.DatabaseConfig.MaxOpenConns)
	}
	if c.DatabaseConfig.MaxIdleConns >= 0 {
		sqldb.SetMaxIdleConns(c.DatabaseConfig.MaxIdleConns)
	}

	db = &database{
		db: sqldb,
	}
	return
}

func postgresConn(pg postgresConfig) (s string, err error) {
	InfoLogger.Info("Postgres database configuration")
	if len(pg.DatabaseName) == 0 {
		err = fmt.Errorf("postgres config missing db_name")
		return
	} else if len(pg.UserName) == 0 {
		err = fmt.Errorf("postgres config missing user")
		return
	}
	s = fmt.Sprintf("dbname=%s user=%s", pg.DatabaseName, pg.UserName)
	var hasPw bool
	hasPw, err = promptDoesXHavePassword(
		fmt.Sprintf(
			"user=%q in db_name=%q",
			pg.DatabaseName,
			pg.UserName))
	if err != nil {
		return
	}
	if hasPw {
		var pw string
		pw, err = promptPassword(
			fmt.Sprintf(
				"Please enter the password for db_name=%q and user=%q:",
				pg.DatabaseName,
				pg.UserName))
		if err != nil {
			return
		}
		s = fmt.Sprintf("%s password=%s", s, pw)
	}
	if len(pg.Host) > 0 {
		s = fmt.Sprintf("%s host=%s", s, pg.Host)
	}
	if pg.Port > 0 {
		s = fmt.Sprintf("%s port=%s", s, pg.Port)
	}
	if len(pg.SSLMode) > 0 {
		s = fmt.Sprintf("%s sslmode=%s", s, pg.SSLMode)
	}
	if len(pg.FallbackApplicationName) > 0 {
		s = fmt.Sprintf("%s fallback_application_name=%s", s, pg.FallbackApplicationName)
	}
	if pg.ConnectTimeout > 0 {
		s = fmt.Sprintf("%s connect_timeout=%s", s, pg.ConnectTimeout)
	}
	if len(pg.SSLCert) > 0 {
		s = fmt.Sprintf("%s sslcert=%s", s, pg.SSLCert)
	}
	if len(pg.SSLKey) > 0 {
		s = fmt.Sprintf("%s sslkey=%s", s, pg.SSLKey)
	}
	if len(pg.SSLRootCert) > 0 {
		s = fmt.Sprintf("%s sslrootcert=%s", s, pg.SSLRootCert)
	}
	return
}
