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

package framework

import (
	"fmt"

	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/services"
)

func PromptNewConfig(file string) (c *config.Config, err error) {
	fmt.Println(ClarkeSays(fmt.Sprintf(`
Welcome to the configuration guided flow!

Here we will visit common configuration choices. While not every option is asked
in the guided flow, you can always open the resulting configuration file to
change options. You can also change your answers to this flow. Note that in
order to take advantage of changed configuration values, the application will
need to be restarted.

Let's go!`)))

	var s string
	s, err = promptSelection(
		"Please choose the database you are using",
		postgresDB)
	if err != nil {
		return
	}
	c, err = defaultConfig(s)
	if err != nil {
		return
	}

	// Prompt for ServerConfig
	c.ServerConfig.Host, err = promptStringWithDefault(
		"Enter the host for this server (ignored in debug mode)",
		"example.com")
	if err != nil {
		return
	}
	c.ServerConfig.Proxy, err = promptYN("Will the application server run behind a proxy?")
	if err != nil {
		return
	}
	if !c.ServerConfig.Proxy {
		c.ServerConfig.CertFile, err = promptString(
			"Enter the path to the file containing the certificate used in HTTPS connections")
		if err != nil {
			return
		}
		c.ServerConfig.KeyFile, err = promptString(
			"Enter the path to the file containing the private key for the certificate used in HTTPS connections")
		if err != nil {
			return
		}
	}
	c.ServerConfig.StaticRootDirectory, err = promptStringWithDefault(
		"Enter the directory for serving static content (WARNING: Everything in it will be served)?",
		"static")
	if err != nil {
		return
	}
	var have bool
	if have, err = promptYN("Do you already have a file containing a cookie authentication private key?"); err != nil {
		return
	} else if have {
		c.ServerConfig.CookieAuthKeyFile, err = promptStringWithDefault(
			"Enter the existing file name for the cookie authentication private key",
			"cookie_authn.key")
		if err != nil {
			return
		}
	} else {
		c.ServerConfig.CookieAuthKeyFile, err = promptStringWithDefault(
			"Enter the new file name for the cookie authentication private key",
			"cookie_authn.key")
		if err != nil {
			return
		}
		err = services.CreateKeyFile(c.ServerConfig.CookieAuthKeyFile)
		if err != nil {
			return
		}
	}
	var want bool
	if have, err = promptYN("Do you already have a file containing a cookie encryption private key?"); err != nil {
		return
	} else if have {
		c.ServerConfig.CookieEncryptionKeyFile, err = promptStringWithDefault(
			"Enter the existing file name for the cookie encryption private key",
			"cookie_enc.key")
		if err != nil {
			return
		}
	} else if want, err = promptYN("Do you want to use a cookie encryption private key?"); err != nil {
		return
	} else if want {
		c.ServerConfig.CookieEncryptionKeyFile, err = promptStringWithDefault(
			"Enter the new file name for the cookie encryption private key",
			"cookie_enc.key")
		if err != nil {
			return
		}
		err = services.CreateKeyFile(c.ServerConfig.CookieEncryptionKeyFile)
		if err != nil {
			return
		}
	}
	c.ServerConfig.CookieSessionName, err = promptStringWithDefault(
		"Session name used to find cookies",
		"my_apcore_session_name")
	if err != nil {
		return
	}
	c.ServerConfig.HttpsReadTimeoutSeconds, err = promptIntWithDefault(
		"Enter the deadline (in seconds) for reading & writing HTTP & HTTPS requests. A value of zero means connections do not timeout",
		60)
	if err != nil {
		return
	}
	c.ServerConfig.HttpsWriteTimeoutSeconds = c.ServerConfig.HttpsReadTimeoutSeconds
	c.ServerConfig.RedirectReadTimeoutSeconds = c.ServerConfig.HttpsReadTimeoutSeconds
	c.ServerConfig.RedirectWriteTimeoutSeconds = c.ServerConfig.HttpsReadTimeoutSeconds

	// Prompt for ActivityPubConfig
	c.ActivityPubConfig.ClockTimezone, err = promptStringWithDefault(
		"Please enter an IANA Time Zone for the server, \"UTC\", or \"Local\"",
		"UTC")
	if err != nil {
		return
	}
	c.ActivityPubConfig.OutboundRateLimitQPS, err = promptFloat64WithDefault(
		"Please enter the steady-state rate limit for outbound ActivityPub QPS",
		2)
	if err != nil {
		return
	}
	c.ActivityPubConfig.OutboundRateLimitBurst, err = promptIntWithDefault(
		"Please enter the burst limit for outbound ActivityPub QPS",
		5)
	if err != nil {
		return
	}

	// Prompt for DatabaseConfig
	c.DatabaseConfig.ConnMaxLifetimeSeconds, err = promptIntWithDefault(
		"Enter the maximum lifetime (in seconds) for database connections. A value of zero means connections do not timeout",
		60)
	if err != nil {
		return
	}
	c.DatabaseConfig.MaxOpenConns, err = promptIntWithDefault(
		"Enter the maximum number of database connections allowed. A value of zero means infinite are permitted.",
		0)

	switch c.DatabaseConfig.DatabaseKind {
	case postgresDB:
		err = promptPostgresConfig(c)
	default:
		err = fmt.Errorf("unknown database kind: %s", c.DatabaseConfig.DatabaseKind)
	}
	return
}

func promptPostgresConfig(c *config.Config) (err error) {
	fmt.Println("Prompting for Postgres database configuration options...")
	c.DatabaseConfig.PostgresConfig.UserName, err = promptStringWithDefault(
		"Enter the postgres user name",
		"pguser")
	if err != nil {
		return
	}
	c.DatabaseConfig.PostgresConfig.Host, err = promptStringWithDefault(
		"Enter the postgres database host name",
		"localhost")
	if err != nil {
		return
	}
	c.DatabaseConfig.PostgresConfig.Port, err = promptIntWithDefault(
		"Enter the postgres database port",
		5432)
	if err != nil {
		return
	}
	c.DatabaseConfig.PostgresConfig.Password, err = promptPassword("Enter the postgres database password")
	if err != nil {
		return
	}
	c.DatabaseConfig.PostgresConfig.SSLMode, err = promptSelection(
		"Please choose a SSL mode (see https://www.postgresql.org/docs/current/libpq-ssl.html)",
		"disable",
		"require",
		"verify-ca",
		"verify-full")
	if err != nil {
		return
	}
	if mode := c.DatabaseConfig.PostgresConfig.SSLMode; mode == "require" || mode == "verify-ca" || mode == "verify-full" {
		fmt.Println(ClarkeSays(fmt.Sprintf(`
Hey, Clarke the Cow here, I noticed you chose %q! Be sure to check your
configuration file for the %q, %q, and/or %q options to get SSL set up properly!
Toodlemoo~`,
			mode,
			"pg_ssl_cert",
			"pg_ssl_key",
			"pg_ssl_root_cert")))
	}
	c.DatabaseConfig.PostgresConfig.DatabaseName, err = promptStringWithDefault(
		"Enter the postgres database name",
		"pgdb")
	if err != nil {
		return
	}
	c.DatabaseConfig.PostgresConfig.Schema, err = promptString("Enter the postgres schema name")
	if err != nil {
		return
	}
	return
}
