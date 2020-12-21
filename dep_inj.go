// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
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

package apcore

import (
	"database/sql"
	"math/rand"
	"time"

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

func newServer(configFileName string, appl app.Application, debug bool) (s *framework.Server, err error) {
	// Load the configuration
	c, err := framework.LoadConfigFile(configFileName, appl, debug)
	if err != nil {
		return
	}

	host := c.ServerConfig.Host
	scheme := schemeFromFlags()

	// Create a server clock, a pub.Clock
	clock, err := ap.NewClock(c.ActivityPubConfig.ClockTimezone)
	if err != nil {
		return
	}

	// ** Create the Models & Services **

	// Create the SQL database
	sqldb, dialect, err := db.NewDB(c)
	if err != nil {
		return
	}

	// Create the models & services for higher-level transformations
	cryp, data, dAttempts, followers, following, inboxes, liked, oauthSrv, outboxes, policies, pkeys, users, nodeinfo, any, models := createModelsAndServices(c, sqldb, dialect, appl, host, scheme, clock)

	// Ensure the SQL statements are prepared
	err = prepare(models, sqldb, dialect)
	if err != nil {
		return
	}

	// ** Create Misc Helpers **

	// Create placeholder framework.
	//
	// Creating a placeholder early allows us to inject it into the needed
	// dependencies, even if *Framework is not yet ready for use.
	fw := &framework.Framework{}
	internalErrorHandler := appl.InternalServerErrorHandler(fw)

	// Prepare web sessions behavior
	sess, err := web.NewSessions(c, scheme)
	if err != nil {
		return
	}

	// Prepare OAuth2 server
	oauth, err := oauth2.NewServer(c, scheme, internalErrorHandler, oauthSrv, cryp, sess)
	if err != nil {
		return
	}

	// Create an HTTP client for this server.
	httpClient := framework.NewHTTPClient(c)

	// ** Initialize the ActivityPub behavior **

	// Create a RoutingDatabase
	db := ap.NewDatabase(scheme,
		c,
		inboxes,
		outboxes,
		users,
		data,
		followers,
		following,
		liked,
		any)

	// Create a pub.Database
	apdb := ap.NewAPDB(db, appl)

	// Create a controller for outbound messaging.
	tc, err := conn.NewController(c, appl, clock, httpClient, dAttempts, pkeys)
	if err != nil {
		return
	}

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
		users,
		tc)
	if err != nil {
		return
	}
	// Hook up ActivityPub Actor behavior for non-user actors.
	actorMap := ap.NewActorMap(c,
		clock,
		db,
		apdb,
		pkeys,
		followers,
		tc)

	// ** Initialize the Web Server **

	// Build framework for auxiliary behaviors
	fw = framework.BuildFramework(fw, oauth, sess, actor, appl)

	// Obtain a normal router and fallback web handlers.
	mr := mux.NewRouter()
	mr.NotFoundHandler = appl.NotFoundHandler(fw)
	mr.MethodNotAllowedHandler = appl.MethodNotAllowedHandler(fw)
	badRequestHandler := appl.BadRequestHandler(fw)
	getAuthWebHandler := appl.GetAuthWebHandlerFunc(fw)
	getLoginWebHandler := appl.GetLoginWebHandlerFunc(fw)

	// Build a specialized AP-aware router for managing and routing HTTP requests.
	r := framework.NewRouter(
		mr,
		oauth,
		actor,
		actorMap,
		clock,
		apdb,
		host,
		scheme,
		internalErrorHandler,
		badRequestHandler)

	// Build application routes for default web support
	h, err := framework.BuildHandler(r,
		internalErrorHandler,
		badRequestHandler,
		getAuthWebHandler,
		getLoginWebHandler,
		scheme,
		c,
		appl,
		fw,
		actor,
		apdb,
		users,
		cryp,
		nodeinfo,
		following,
		followers,
		liked,
		sqldb,
		oauth,
		sess,
		fw,
		clock,
		appl.Software(), apCoreSoftware(),
		debug)
	if err != nil {
		return
	}

	// Build list of StartStoppers
	ss := []framework.StartStopper{tc, oauth}

	// Build web server to control server behavior
	if debug {
		s, err = framework.NewInsecureServer(c, h, appl, sqldb, dialect, models, ss)
	} else {
		s, err = framework.NewHTTPSServer(c, h, appl, sqldb, dialect, models, ss)
	}
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

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, m = createModelsAndServices(c, sqldb, dialect, appl, host, scheme, clock)
	return
}

func newUserService(configFileName string, appl app.Application, debug bool, scheme string) (sqldb *sql.DB, users *services.Users, c *config.Config, err error) {
	// Load the configuration
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
	var dialect models.SqlDialect
	sqldb, dialect, err = db.NewDB(c)
	if err != nil {
		return
	}

	var ml []models.Model
	_, _, _, _, _, _, _, _, _, _, _, users, _, _, ml = createModelsAndServices(c, sqldb, dialect, appl, host, scheme, clock)
	err = prepare(ml, sqldb, dialect)
	return
}

func createModelsAndServices(c *config.Config, sqldb *sql.DB, d models.SqlDialect, appl app.Application, host, scheme string, clock pub.Clock) (cryp *services.Crypto,
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
	nodeinfo *services.NodeInfo,
	any *services.Any,
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
	cd := &models.Credentials{}
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
		cd,
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
	data = &services.Data{
		DB:                    sqldb,
		Hostname:              host,
		FedData:               fd,
		LocalData:             ld,
		Users:                 us,
		Following:             following,
		Followers:             followers,
		Liked:                 liked,
		DefaultCollectionSize: c.DatabaseConfig.DefaultCollectionPageSize,
		MaxCollectionPageSize: c.DatabaseConfig.MaxCollectionPageSize,
	}
	oauth = &services.OAuth2{
		DB:     sqldb,
		Client: ci,
		Token:  ti,
		Creds:  cd,
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
	nodeinfo = &services.NodeInfo{
		DB:               sqldb,
		Users:            us,
		LocalData:        ld,
		Rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
		CacheInvalidated: time.Second * time.Duration(c.NodeInfoConfig.AnonymizedStatsCacheInvalidatedSeconds),
	}
	any = &services.Any{
		DB:      sqldb,
		Dialect: d,
	}
	return
}

func prepare(ml []models.Model, db *sql.DB, d models.SqlDialect) error {
	for _, m := range ml {
		if err := m.Prepare(db, d); err != nil {
			return err
		}
	}
	return nil
}
