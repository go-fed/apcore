package apcore

import (
	"fmt"

	"gopkg.in/ini.v1"
)

const (
	postgresDB = "postgres"
)

// Overall configuration file structure
type config struct {
	ServerConfig      serverConfig      `ini:"server" comment:"HTTP server configuration"`
	DatabaseConfig    databaseConfig    `ini:"database" comment:"Database configuration"`
	ActivityPubConfig activityPubConfig `ini:"activitypub" comment:"ActivityPub configuration"`
}

func defaultConfig(dbkind string) (c *config, err error) {
	var dbc databaseConfig
	dbc, err = defaultDatabaseConfig(dbkind)
	if err != nil {
		return
	}
	c = &config{
		ServerConfig:      defaultServerConfig(),
		DatabaseConfig:    dbc,
		ActivityPubConfig: defaultActivityPubConfig(),
	}
	return
}

// Configuration section specifically for the HTTP server.
type serverConfig struct {
	Host                        string `ini:"sr_host" comment:"(required) Host for this instance; ignored in debug mode"`
	CookieAuthKeyFile           string `ini:"sr_cookie_auth_key_file" comment:"(required) Path to private key file used for cookie authentication"`
	CookieEncryptionKeyFile     string `ini:"sr_cookie_encryption_key_file" comment:"Path to private key file used for cookie encryption"`
	HttpsReadTimeoutSeconds     int    `ini:"sr_https_read_timeout_seconds" comment:"Timeout in seconds for incoming HTTPS requests; a zero or unset value does not timeout"`
	HttpsWriteTimeoutSeconds    int    `ini:"sr_https_write_timeout_seconds" comment:"Timeout in seconds for outgoing HTTPS responses; a zero or unset value does not timeout"`
	RedirectReadTimeoutSeconds  int    `ini:"sr_redirect_read_timeout_seconds" comment:"Timeout in seconds for incoming HTTP requests, which will be redirected to HTTPS; a zero or unset value does not timeout"`
	RedirectWriteTimeoutSeconds int    `ini:"sr_redirect_write_timeout_seconds" comment:"Timeout in seconds for outgoing HTTP redirect-to-HTTPS responses; a zero or unset value does not timeout"`
	StaticRootDirectory         string `ini:"sr_static_root_directory" comment:"(required) Root directory for serving static content, such as ECMAScript, CSS, favicon; !!!Warning: Everything in this directory will be served and accessible!!!"`
}

func defaultServerConfig() serverConfig {
	return serverConfig{}
}

// Configuration section specifically for the database.
type databaseConfig struct {
	DatabaseKind           string         `ini:"db_database_kind" comment:"(required) Only \"postgres\" supported"`
	ConnMaxLifetimeSeconds int            `ini:"db_conn_max_lifetime_seconds" comment:"(default: indefinite) Maximum lifetime of a connection in seconds; a value of zero or unset value means indefinite"`
	MaxOpenConns           int            `ini:"db_max_open_conns" comment:"(default: infinite) Maximum number of open connections to the database; a value of zero or unset value means infinite"`
	MaxIdleConns           int            `ini:"db_max_idle_conns" comment:"(default: 2) Maximum number of idle connections in the connection pool to the database; a value of zero maintains no idle connections; a value greater than max_open_conns is reduced to be equal to max_open_conns"`
	PostgresConfig         postgresConfig `ini:"db_postgres,omitempty" comment:"Only needed if database_kind is postgres, and values are based on the github.com/lib/pq driver"`
}

func defaultDatabaseConfig(dbkind string) (d databaseConfig, err error) {
	d = databaseConfig{
		DatabaseKind: dbkind,
		// This default is implicit in Go but could change, so here we
		// make it explicit instead
		MaxIdleConns: 2,
	}
	if dbkind != postgresDB {
		err = fmt.Errorf("unsupported database kind: %s", dbkind)
		return
	}
	d.PostgresConfig = defaultPostgresConfig()
	return
}

// Configuration section specifically for ActivityPub.
type activityPubConfig struct {
	ClockTimezone string `ini:"clock_timezone" comment:"(default: UTC) Timezone for ActivityPub related operations: unset and \"UTC\" are UTC, \"Local\" is local server time, otherwise use IANA Time Zone database values"`
}

func defaultActivityPubConfig() activityPubConfig {
	return activityPubConfig{
		ClockTimezone: "UTC",
	}
}

// Configuration section specifically for Postgres databases.
type postgresConfig struct {
	DatabaseName            string `ini:"pg_db_name" comment:"(required) Database name"`
	UserName                string `ini:"pg_user" comment:"(required) User to connect as (any password will be prompted)"`
	Host                    string `ini:"pg_host" comment:"(default: localhost) The Postgres host to connect to"`
	Port                    int    `ini:"pg_port" comment:"(default: 5432) The port to connect to"`
	SSLMode                 string `ini:"pg_ssl_mode" comment:"(default: require) SSL mode to use when connecting (options are: \"disable\", \"require\", \"verify-ca\", \"verify-full\")"`
	FallbackApplicationName string `ini:"pg_fallback_application_name" comment:"An application_name to fall back to if one is not provided"`
	ConnectTimeout          int    `ini:"pg_connect_timeout" comment:"(default: indefinite) Maximum wait when connecting to a database, zero or unset means indefinite"`
	SSLCert                 string `ini:"pg_ssl_cert" comment:"PEM-encoded certificate file location"`
	SSLKey                  string `ini:"pg_ssl_key" comment:"PEM-encoded private key file location"`
	SSLRootCert             string `ini:"pg_ssl_root_cert" comment:"PEM-encoded root certificate file location"`
}

func defaultPostgresConfig() postgresConfig {
	return postgresConfig{}
}

func loadConfigFile(filename string, a Application, debug bool) (c *config, err error) {
	InfoLogger.Infof("Loading config file: %s", filename)
	var cfg *ini.File
	cfg, err = ini.Load(filename)
	if err != nil {
		return
	}
	err = cfg.MapTo(c)
	if err != nil {
		return
	}
	appCfg := a.NewConfiguration()
	err = cfg.MapTo(appCfg)
	if err != nil {
		return
	}
	err = a.SetConfiguration(appCfg)
	if err != nil {
		return
	}
	if debug {
		c.ServerConfig.Host = "localhost"
	}
	return
}

func saveConfigFile(filename string, c *config, others ...interface{}) error {
	InfoLogger.Infof("Saving config file: %s", filename)
	cfg := ini.Empty()
	err := ini.ReflectFrom(cfg, c)
	if err != nil {
		return err
	}
	for _, o := range others {
		err = ini.ReflectFrom(cfg, o)
		if err != nil {
			return err
		}
	}
	return cfg.SaveTo(filename)
}

func promptNewConfig(file string) (c *config, err error) {
	// TODO

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
		// TODO: Generate the key
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
		// TODO: Generate the key
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

func promptPostgresConfig(c *config) (err error) {
	fmt.Println("Prompting for Postgres database configuration options...")
	c.DatabaseConfig.PostgresConfig.DatabaseName, err = promptStringWithDefault(
		"Enter the postgres database name",
		"pgdb")
	if err != nil {
		return
	}
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
		fmt.Println(clarkeSays(fmt.Sprintf(`
Hey, Clarke the Cow here, I noticed you chose %q! Be sure to check your
configuration file for the %q, %q, and/or %q options to get SSL set up properly!
Toodlemoo~`,
			mode,
			"pg_ssl_cert",
			"pg_ssl_key",
			"pg_ssl_root_cert")))
	}
	return
}
