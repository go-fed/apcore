package apcore

import ()

func newServer(configFileName string, a app.Application, debug bool, scheme string) (s *server, err error) {
	// Load the configuration
	var c *config.Config
	c, err = loadConfigFile(configFileName, a, debug)
	if err != nil {
		return
	}

	// Enforce server level configuration
	if c.ServerConfig.RSAKeySize < minKeySize {
		err = fmt.Errorf("RSA private key size is configured to be < %d, which is forbidden: %d", minKeySize, c.ServerConfig.RSAKeySize)
		return
	}

	// Connect to database
	var db *database
	db, err = newDatabase(c, a, debug)
	if err != nil {
		return
	}

	// Prepare sessions
	var ses *web.Sessions
	ses, err = newSessions(c)
	if err != nil {
		return
	}

	// Prepare OAuth2 server
	var oa *oauth2.Server
	oa, err = newOAuth2Server(c, a, db, ses)
	if err != nil {
		return
	}

	// TODO: Reexamine this.
	httpClient := &http.Client{}

	// Initialize the ActivityPub portion of the server
	var clock *ap.Clock
	clock, err = newClock(c.ActivityPubConfig.ClockTimezone)
	if err != nil {
		return
	}

	var apdb *apdb
	apdb = newApdb(db, a)

	var tc *transportController
	tc, err = newTransportController(c, a, clock, httpClient, db)
	if err != nil {
		return
	}

	var actor pub.Actor
	actor, err = newActor(c, a, clock, p, db, apdb, oa, tc)
	if err != nil {
		return
	}

	r := NewRouter(
		mr,
		oauth,
		actor,
		clock,
		db,
		c.ServerConfig.Host,
		scheme,
		internalErrorHandler,
		badRequestHandler)

	newFramework(scheme, c.ServerConfig.Host, oauth, sqldb, actor, a.S2SEnabled())

	// Build application routes
	var h *handler
	h, err = BuildHandler(scheme, c, a, actor, apdb, oa, ses, clock, debug)
	if err != nil {
		return
	}
}
