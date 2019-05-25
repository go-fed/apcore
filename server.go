package apcore

import (
	"github.com/gorilla/mux"
	_ "github.com/gorilla/schema"
)

type server struct {
	a        Application
	router   *mux.Router
	sessions *sessions
	config   *config
}

func newServer(configFileName string, a Application) (s *server, err error) {
	var c *config
	c, err = loadConfigFile(configFileName, a)
	if err != nil {
		return
	}

	var ses *sessions
	ses, err = newSessions(c)
	if err != nil {
		return
	}

	r := mux.NewRouter()

	s = &server{
		a:        a,
		router:   r,
		sessions: ses,
		config:   c,
	}
	return
}
