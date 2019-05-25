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
	authKey, err = ioutil.ReadFile(c.ServerConfig.CookieAuthKeyFile)
	if err != nil {
		return
	}
	encKey, err = ioutil.ReadFile(c.ServerConfig.CookieEncryptionKeyFile)
	if err != nil {
		return
	}
	s = &sessions{
		cookies: gs.NewCookieStore(authKey, encKey),
	}
	return
}
