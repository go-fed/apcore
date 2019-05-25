package apcore

import (
	"gopkg.in/ini.v1"
)

// Overall configuration file structure
type config struct {
	ServerConfig      serverConfig      `ini:"server"`
	DatabaseConfig    databaseConfig    `ini:"database"`
	ActivityPubConfig activityPubConfig `ini:"activitypub"`
}

// Configuration section specifically for the HTTP server.
type serverConfig struct {
	CookieAuthKeyFile           string `ini:"cookie_auth_key_file"`
	CookieEncryptionKeyFile     string `ini:"cookie_encryption_key_file"`
	HttpsReadTimeoutSeconds     int    `ini:"https_read_timeout_seconds"`
	HttpsWriteTimeoutSeconds    int    `ini:"https_write_timeout_seconds"`
	RedirectReadTimeoutSeconds  int    `ini:"redirect_read_timeout_seconds"`
	RedirectWriteTimeoutSeconds int    `ini:"redirect_write_timeout_seconds"`
}

// Configuration section specifically for the database.
type databaseConfig struct{}

// Configuration section specifically for ActivityPub.
type activityPubConfig struct{}

func loadConfigFile(filename string, a Application) (c *config, err error) {
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
