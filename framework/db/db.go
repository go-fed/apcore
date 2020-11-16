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

package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

func NewDB(c *config.Config) (sqldb *sql.DB, d models.SqlDialect, err error) {
	kind := c.DatabaseConfig.DatabaseKind
	var conn string
	switch kind {
	case "postgres":
		conn, err = postgresConn(c.DatabaseConfig.PostgresConfig)
		d = NewPgV0(c.DatabaseConfig.PostgresConfig.Schema)
	default:
		err = fmt.Errorf("unhandled database_kind in config: %s", kind)
	}
	if err != nil {
		return
	}

	util.InfoLogger.Infof("Calling sql.Open...")
	sqldb, err = sql.Open(kind, conn)
	if err != nil {
		return
	}
	util.InfoLogger.Infof("Calling sql.Open complete")

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
	util.InfoLogger.Infof("Database connections configured successfully")
	util.InfoLogger.Infof("NOTE: No underlying database connections may have happened yet!")
	return
}

func MustPing(db *sql.DB) (err error) {
	util.InfoLogger.Infof("Opening connection to database by pinging, which will create a connection...")
	start := time.Now()
	err = db.Ping()
	if err != nil {
		util.ErrorLogger.Errorf("Unsuccessfully pinged database: %s", err)
		return
	}
	end := time.Now()
	util.InfoLogger.Infof("Successfully pinged database with latency: %s", end.Sub(start))
	return
}

func postgresConn(pg config.PostgresConfig) (s string, err error) {
	util.InfoLogger.Info("Postgres database configuration")
	if len(pg.DatabaseName) == 0 {
		err = fmt.Errorf("postgres config missing db_name")
		return
	} else if len(pg.UserName) == 0 {
		err = fmt.Errorf("postgres config missing user")
		return
	}
	s = fmt.Sprintf("dbname=%s user=%s", pg.DatabaseName, pg.UserName)
	if len(pg.Password) > 0 {
		s = fmt.Sprintf("%s password=%s", s, pg.Password)
	}
	if len(pg.Host) > 0 {
		s = fmt.Sprintf("%s host=%s", s, pg.Host)
	}
	if pg.Port > 0 {
		s = fmt.Sprintf("%s port=%d", s, pg.Port)
	}
	if len(pg.SSLMode) > 0 {
		s = fmt.Sprintf("%s sslmode=%s", s, pg.SSLMode)
	}
	if len(pg.FallbackApplicationName) > 0 {
		s = fmt.Sprintf("%s fallback_application_name=%s", s, pg.FallbackApplicationName)
	}
	if pg.ConnectTimeout > 0 {
		s = fmt.Sprintf("%s connect_timeout=%d", s, pg.ConnectTimeout)
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
