package apcore

import (
	"database/sql"
	"net/http"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/apcore/ap"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/framework/db"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/framework/web"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/services"
	"github.com/gorilla/mux"
)

func newServer(configFileName string, appl app.Application, debug bool, scheme string) (s *framework.Server, err error) {
	// Load the configuration
	c, err := framework.LoadConfigFile(configFileName, appl, debug)

	host := c.ServerConfig.Host
	// TODO: scheme = http for debug

	// Create a server clock, a pub.Clock
	clock, err := ap.NewClock(c.ActivityPubConfig.ClockTimezone)

	// ** Create the Models & Services **

	// Create the SQL database
	sqldb, dialect, err := db.NewDB(c)

	// Create the models & services for higher-level transformations
	cryp, data, dAttempts, followers, following, inboxes, liked, oauthSrv, outboxes, policies, pkeys, users, models := createModelsAndServices(sqldb, appl, host, scheme, clock)

	// ** Create Misc Helpers **

	// Prepare web sessions behavior
	sess, err := web.NewSessions(c)

	// Prepare OAuth2 server
	oauth, err := oauth2.NewServer(c, appl, oauthSrv, cryp, sess)

	// Create an HTTP client for this server.
	// TODO: Reexamine this.
	httpClient := &http.Client{}

	// ** Initialize the ActivityPub behavior **

	// Create a RoutingDatabase
	db := ap.NewDatabase(c,
		inboxes,
		outboxes,
		users,
		data,
		followers,
		following,
		liked)

	// Create a pub.Database
	apdb := ap.NewAPDB(db, appl)

	// Create a controller for outbound messaging.
	tc, err := conn.NewController(c, appl, clock, httpClient, dAttempts)

	// Hook up ActivityPub Actor behavior for users.
	actor, err := ap.NewActor(c,
		appl,
		clock,
		db,
		apdb,
		oauth,
		pkeys,
		policies,
		followers,
		tc)

	// ** Initialize the Web Server **

	// Obtain a normal router and fallback web handlers.
	mr := mux.NewRouter()
	mr.NotFoundHandler = appl.NotFoundHandler()
	mr.MethodNotAllowedHandler = appl.MethodNotAllowedHandler()
	internalErrorHandler := appl.InternalServerErrorHandler()
	badRequestHandler := appl.BadRequestHandler()
	getAuthWebHandler := appl.GetAuthWebHandlerFunc()
	getLoginWebHandler := appl.GetLoginWebHandlerFunc()

	// Build a specialized AP-aware router for managing and routing HTTP requests.
	r := framework.NewRouter(
		mr,
		oauth,
		actor,
		clock,
		apdb,
		host,
		scheme,
		internalErrorHandler,
		badRequestHandler)

	// Build framework for auxiliary behaviors
	fw := framework.NewFramework(oauth, actor, appl.S2SEnabled())

	// Build application routes for default web support
	h, err := framework.BuildHandler(r,
		internalErrorHandler,
		badRequestHandler,
		getAuthWebHandler,
		getLoginWebHandler,
		scheme,
		c,
		appl,
		actor,
		apdb,
		users,
		cryp,
		sqldb,
		oauth,
		sess,
		fw,
		clock,
		debug)
	if err != nil {
		return
	}

	// Build web server to control server behavior
	s, err = framework.NewServer(c, h, scheme, appl, sqldb, dialect, models)
	return
}

func newModels(configFileName string, appl app.Application, debug bool, scheme string) (sqldb *sql.DB, dialect models.SqlDialect, m []models.Model, err error) {
	// Load the configuration
	var c *config.Config
	c, err = framework.LoadConfigFile(configFileName, appl, debug)
	if err != nil {
		return
	}
	host := c.ServerConfig.Host

	// Create a server clock, a pub.Clock
	var clock pub.Clock
	clock, err = ap.NewClock(c.ActivityPubConfig.ClockTimezone)
	if err != nil {
		return
	}

	// Create the SQL database
	sqldb, dialect, err = db.NewDB(c)
	if err != nil {
		return
	}

	_, _, _, _, _, _, _, _, _, _, _, _, m = createModelsAndServices(sqldb, appl, host, scheme, clock)
	return
}

func createModelsAndServices(sqldb *sql.DB, appl app.Application, host, scheme string, clock pub.Clock) (cryp *services.Crypto,
	data *services.Data,
	dAttempts *services.DeliveryAttempts,
	followers *services.Followers,
	following *services.Following,
	inboxes *services.Inboxes,
	liked *services.Liked,
	oauth *services.OAuth2,
	outboxes *services.Outboxes,
	policies *services.Policies,
	pkeys *services.PrivateKeys,
	users *services.Users,
	m []models.Model) {
	us := &models.Users{}
	fd := &models.FedData{}
	ld := &models.LocalData{}
	in := &models.Inboxes{}
	ou := &models.Outboxes{}
	da := &models.DeliveryAttempts{}
	pk := &models.PrivateKeys{}
	ci := &models.ClientInfos{}
	ti := &models.TokenInfos{}
	fn := &models.Following{}
	fr := &models.Followers{}
	li := &models.Liked{}
	po := &models.Policies{}
	rs := &models.Resolutions{}
	m = []models.Model{
		us,
		fd,
		ld,
		in,
		ou,
		da,
		pk,
		ci,
		ti,
		fn,
		fr,
		li,
		po,
		rs,
	}
	cryp = &services.Crypto{
		DB:    sqldb,
		Users: us,
	}
	data = &services.Data{
		DB:        sqldb,
		Hostname:  host,
		FedData:   fd,
		LocalData: ld,
	}
	dAttempts = &services.DeliveryAttempts{
		DB:               sqldb,
		DeliveryAttempts: da,
	}
	followers = &services.Followers{
		DB:        sqldb,
		Followers: fr,
	}
	following = &services.Following{
		DB:        sqldb,
		Following: fn,
	}
	inboxes = &services.Inboxes{
		DB:      sqldb,
		Inboxes: in,
	}
	liked = &services.Liked{
		DB:    sqldb,
		Liked: li,
	}
	oauth = &services.OAuth2{
		DB:     sqldb,
		Client: ci,
		Token:  ti,
	}
	outboxes = &services.Outboxes{
		DB:       sqldb,
		Outboxes: ou,
	}
	policies = &services.Policies{
		Clock:       clock,
		DB:          sqldb,
		Policies:    po,
		Resolutions: rs,
	}
	pkeys = &services.PrivateKeys{
		Scheme:      scheme,
		Host:        host,
		DB:          sqldb,
		PrivateKeys: pk,
	}
	users = &services.Users{
		App:         appl,
		DB:          sqldb,
		Users:       us,
		PrivateKeys: pk,
		Inboxes:     in,
		Outboxes:    ou,
		Followers:   fr,
		Following:   fn,
		Liked:       li,
	}
	return
}