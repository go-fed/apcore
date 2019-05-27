package apcore

import (
	"gopkg.in/ini.v1"
)

// Overall configuration file structure
type config struct {
	ServerConfig      serverConfig      `ini:"server" comment:"HTTP server configuration"`
	DatabaseConfig    databaseConfig    `ini:"database" comment:"Database configuration"`
	ActivityPubConfig activityPubConfig `ini:"activitypub" comment:"ActivityPub configuration"`
}

// Configuration section specifically for the HTTP server.
type serverConfig struct {
	CookieAuthKeyFile           string `ini:"cookie_auth_key_file" comment:"(required) Path to private key file used for cookie authentication"`
	CookieEncryptionKeyFile     string `ini:"cookie_encryption_key_file" comment:"Path to private key file used for cookie encryption"`
	HttpsReadTimeoutSeconds     int    `ini:"https_read_timeout_seconds" comment:"Timeout in seconds for incoming HTTPS requests"`
	HttpsWriteTimeoutSeconds    int    `ini:"https_write_timeout_seconds" comment:"Timeout in seconds for outgoing HTTPS responses"`
	RedirectReadTimeoutSeconds  int    `ini:"redirect_read_timeout_seconds" comment:"Timeout in seconds for incoming HTTP requests, which will be redirected to HTTPS"`
	RedirectWriteTimeoutSeconds int    `ini:"redirect_write_timeout_seconds" comment:"Timeout in seconds for outgoing HTTP redirect-to-HTTPS responses"`
	StaticRootDirectory         string `ini:"static_root_directory" comment:"(required) Root directory for serving static content, such as ECMAScript, CSS, favicon; !!!Warning: Everything in this directory will be served and accessible!!!"`
}

// Configuration section specifically for the database.
type databaseConfig struct {
	DatabaseKind           string         `ini:"database_kind" comment:"(required) Only \"postgres\" supported"`
	ConnMaxLifetimeSeconds int            `ini:"conn_max_lifetime_seconds" comment:"(default: indefinite) Maximum lifetime of a connection in seconds; a value of zero or unset value means indefinite"`
	MaxOpenConns           int            `ini:"max_open_conns" comment:"(default: infinite) Maximum number of open connections to the database; a value of zero or unset value means indefinite"`
	MaxIdleConns           int            `ini:"max_idle_conns" comment:"(default: 2) Maximum number of idle connections in the connection pool to the database; a value of zero maintains no idle connections; a value greater than max_open_conns is reduced to be equal to max_open_conns"`
	PostgresConfig         postgresConfig `ini:"postgres" comment:"Only needed if database_kind is postgres, and values are based on the github.com/lib/pq driver"`
}

// Configuration section specifically for ActivityPub.
type activityPubConfig struct{}

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

func loadConfigFile(filename string, a Application) (c *config, err error) {
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
