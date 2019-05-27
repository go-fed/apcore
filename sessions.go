package apcore

import (
	"io/ioutil"

	gs "github.com/gorilla/sessions"
)

type sessions struct {
	cookies *gs.CookieStore
}

func newSessions(c *config) (s *sessions, err error) {
	var authKey, encKey []byte
	var keys [][]byte
	authKey, err = ioutil.ReadFile(c.ServerConfig.CookieAuthKeyFile)
	if err != nil {
		return
	}
	if len(c.ServerConfig.CookieEncryptionKeyFile) > 0 {
		InfoLogger.Info("Cookie encryption key file detected")
		encKey, err = ioutil.ReadFile(c.ServerConfig.CookieEncryptionKeyFile)
		if err != nil {
			return
		}
		keys = [][]byte{authKey, encKey}
	} else {
		InfoLogger.Info("No cookie encryption key file detected")
		keys = [][]byte{authKey}
	}
	s = &sessions{
		cookies: gs.NewCookieStore(keys...),
	}
	return
}
