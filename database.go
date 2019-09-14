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

package apcore

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	_ "github.com/lib/pq"
	"gopkg.in/oauth2.v3"
)

type Database interface{}

var _ Database = &database{}

type database struct {
	db     *sql.DB
	sqlgen sqlGenerator
	// url.URL.Host name for this server
	hostname string
	// default size of fetching pages of inbox, outboxes, etc
	defaultCollectionSize int
	// Prepared statements for apcore
	hashPassForUserID    *sql.Stmt
	userIdForEmail       *sql.Stmt
	userIdForBoxPath     *sql.Stmt
	userPreferences      *sql.Stmt
	insertUserPolicy     *sql.Stmt
	insertInstancePolicy *sql.Stmt
	updateUserPolicy     *sql.Stmt
	updateInstancePolicy *sql.Stmt
	instancePolicies     *sql.Stmt
	userPolicies         *sql.Stmt
	userResolutions      *sql.Stmt
	insertUserPKey       *sql.Stmt
	getUserPKey          *sql.Stmt
	// Prepared statements for persistent delivery
	insertAttempt           *sql.Stmt
	markSuccessfulAttempt   *sql.Stmt
	markRetryFailureAttempt *sql.Stmt
	// Prepared statements for oauth
	createTokenInfo      *sql.Stmt
	removeTokenByCode    *sql.Stmt
	removeTokenByAccess  *sql.Stmt
	removeTokenByRefresh *sql.Stmt
	getTokenByCode       *sql.Stmt
	getTokenByAccess     *sql.Stmt
	getTokenByRefresh    *sql.Stmt
	getClientById        *sql.Stmt
	// Prepared statements for the database required by go-fed
	inboxContains  *sql.Stmt
	getInbox       *sql.Stmt
	actorForOutbox *sql.Stmt
	actorForInbox  *sql.Stmt
	outboxForInbox *sql.Stmt
	exists         *sql.Stmt
	get            *sql.Stmt
	localCreate    *sql.Stmt
	fedCreate      *sql.Stmt
	localUpdate    *sql.Stmt
	fedUpdate      *sql.Stmt
	localDelete    *sql.Stmt
	fedDelete      *sql.Stmt
	getOutbox      *sql.Stmt
	followers      *sql.Stmt
	following      *sql.Stmt
	liked          *sql.Stmt
}

func newDatabase(c *config, a Application, debug bool) (db *database, err error) {
	kind := c.DatabaseConfig.DatabaseKind
	var conn string
	var sqlgen sqlGenerator
	switch kind {
	case "postgres":
		sqlgen = newPgV0(c.DatabaseConfig.PostgresConfig.Schema, debug)
		conn, err = postgresConn(c.DatabaseConfig.PostgresConfig)
	default:
		err = fmt.Errorf("unhandled database_kind in config: %s", kind)
	}
	if err != nil {
		return
	}

	InfoLogger.Infof("Opening database")
	var sqldb *sql.DB
	sqldb, err = sql.Open(kind, conn)
	if err != nil {
		return
	}
	InfoLogger.Infof("DB Open complete")

	// Apply general database configurations
	if c.DatabaseConfig.ConnMaxLifetimeSeconds > 0 {
		sqldb.SetConnMaxLifetime(
			time.Duration(c.DatabaseConfig.ConnMaxLifetimeSeconds) *
				time.Second)
	}
	if c.DatabaseConfig.MaxOpenConns > 0 {
		sqldb.SetMaxOpenConns(c.DatabaseConfig.MaxOpenConns)
	}
	if c.DatabaseConfig.MaxIdleConns >= 0 {
		sqldb.SetMaxIdleConns(c.DatabaseConfig.MaxIdleConns)
	}

	InfoLogger.Infof("Pinging database to force-check an initial connection...")
	start := time.Now()
	err = sqldb.Ping()
	if err != nil {
		InfoLogger.Infof("Unsuccessfully pinged database")
		return
	}
	end := time.Now()
	InfoLogger.Infof("Successfully pinged database with latency: %s", end.Sub(start))

	db = &database{
		db:                    sqldb,
		sqlgen:                sqlgen,
		hostname:              c.ServerConfig.Host,
		defaultCollectionSize: c.DatabaseConfig.DefaultCollectionPageSize,
	}
	// apcore statement preparations
	db.hashPassForUserID, err = db.db.Prepare(sqlgen.HashPassForUserID())
	if err != nil {
		return
	}
	db.userIdForEmail, err = db.db.Prepare(sqlgen.UserIdForEmail())
	if err != nil {
		return
	}
	db.userIdForBoxPath, err = db.db.Prepare(sqlgen.UserIdForBoxPath())
	if err != nil {
		return
	}
	db.userPreferences, err = db.db.Prepare(sqlgen.UserPreferences())
	if err != nil {
		return
	}
	db.updateUserPolicy, err = db.db.Prepare(sqlgen.UpdateUserPolicy())
	if err != nil {
		return
	}
	db.updateInstancePolicy, err = db.db.Prepare(sqlgen.UpdateInstancePolicy())
	if err != nil {
		return
	}
	db.insertUserPolicy, err = db.db.Prepare(sqlgen.InsertUserPolicy())
	if err != nil {
		return
	}
	db.insertInstancePolicy, err = db.db.Prepare(sqlgen.InsertInstancePolicy())
	if err != nil {
		return
	}
	db.instancePolicies, err = db.db.Prepare(sqlgen.InstancePolicies())
	if err != nil {
		return
	}
	db.userPolicies, err = db.db.Prepare(sqlgen.UserPolicies())
	if err != nil {
		return
	}
	db.userResolutions, err = db.db.Prepare(sqlgen.UserResolutions())
	if err != nil {
		return
	}
	db.insertUserPKey, err = db.db.Prepare(sqlgen.InsertUserPKey())
	if err != nil {
		return
	}
	db.getUserPKey, err = db.db.Prepare(sqlgen.GetUserPKey())
	if err != nil {
		return
	}

	// prepared statements for persistent delivery
	db.insertAttempt, err = db.db.Prepare(sqlgen.InsertAttempt())
	if err != nil {
		return
	}
	db.markSuccessfulAttempt, err = db.db.Prepare(sqlgen.MarkSuccessfulAttempt())
	if err != nil {
		return
	}
	db.markRetryFailureAttempt, err = db.db.Prepare(sqlgen.MarkRetryFailureAttempt())
	if err != nil {
		return
	}

	// prepared statements for oauth
	db.createTokenInfo, err = db.db.Prepare(sqlgen.CreateTokenInfo())
	if err != nil {
		return
	}
	db.removeTokenByCode, err = db.db.Prepare(sqlgen.RemoveTokenByCode())
	if err != nil {
		return
	}
	db.removeTokenByAccess, err = db.db.Prepare(sqlgen.RemoveTokenByAccess())
	if err != nil {
		return
	}
	db.removeTokenByRefresh, err = db.db.Prepare(sqlgen.RemoveTokenByRefresh())
	if err != nil {
		return
	}
	db.getTokenByCode, err = db.db.Prepare(sqlgen.GetTokenByCode())
	if err != nil {
		return
	}
	db.getTokenByAccess, err = db.db.Prepare(sqlgen.GetTokenByAccess())
	if err != nil {
		return
	}
	db.getTokenByRefresh, err = db.db.Prepare(sqlgen.GetTokenByRefresh())
	if err != nil {
		return
	}
	db.getClientById, err = db.db.Prepare(sqlgen.GetClientById())
	if err != nil {
		return
	}

	// go-fed statement preparations
	db.inboxContains, err = db.db.Prepare(sqlgen.InboxContains())
	if err != nil {
		return
	}
	db.getInbox, err = db.db.Prepare(sqlgen.GetInbox())
	if err != nil {
		return
	}
	db.actorForOutbox, err = db.db.Prepare(sqlgen.ActorForOutbox())
	if err != nil {
		return
	}
	db.actorForInbox, err = db.db.Prepare(sqlgen.ActorForInbox())
	if err != nil {
		return
	}
	db.outboxForInbox, err = db.db.Prepare(sqlgen.OutboxForInbox())
	if err != nil {
		return
	}
	db.exists, err = db.db.Prepare(sqlgen.Exists())
	if err != nil {
		return
	}
	db.get, err = db.db.Prepare(sqlgen.Get())
	if err != nil {
		return
	}
	db.localCreate, err = db.db.Prepare(sqlgen.LocalCreate())
	if err != nil {
		return
	}
	db.fedCreate, err = db.db.Prepare(sqlgen.FedCreate())
	if err != nil {
		return
	}
	db.localUpdate, err = db.db.Prepare(sqlgen.LocalUpdate())
	if err != nil {
		return
	}
	db.fedUpdate, err = db.db.Prepare(sqlgen.FedUpdate())
	if err != nil {
		return
	}
	db.localDelete, err = db.db.Prepare(sqlgen.LocalDelete())
	if err != nil {
		return
	}
	db.fedDelete, err = db.db.Prepare(sqlgen.FedDelete())
	if err != nil {
		return
	}
	db.getOutbox, err = db.db.Prepare(sqlgen.GetOutbox())
	if err != nil {
		return
	}
	db.followers, err = db.db.Prepare(sqlgen.Followers())
	if err != nil {
		return
	}
	db.following, err = db.db.Prepare(sqlgen.Following())
	if err != nil {
		return
	}
	db.liked, err = db.db.Prepare(sqlgen.Liked())
	if err != nil {
		return
	}
	return
}

func postgresConn(pg postgresConfig) (s string, err error) {
	InfoLogger.Info("Postgres database configuration")
	if len(pg.DatabaseName) == 0 {
		err = fmt.Errorf("postgres config missing db_name")
		return
	} else if len(pg.UserName) == 0 {
		err = fmt.Errorf("postgres config missing user")
		return
	}
	s = fmt.Sprintf("dbname=%s user=%s", pg.DatabaseName, pg.UserName)
	var hasPw bool
	hasPw, err = promptDoesXHavePassword(
		fmt.Sprintf(
			"user=%q in db_name=%q",
			pg.DatabaseName,
			pg.UserName))
	if err != nil {
		return
	}
	if hasPw {
		var pw string
		pw, err = promptPassword(
			fmt.Sprintf(
				"Please enter the password for db_name=%q and user=%q:",
				pg.DatabaseName,
				pg.UserName))
		if err != nil {
			return
		}
		s = fmt.Sprintf("%s password=%s", s, pw)
	}
	if len(pg.Host) > 0 {
		s = fmt.Sprintf("%s host=%s", s, pg.Host)
	}
	if pg.Port > 0 {
		s = fmt.Sprintf("%s port=%d", s, pg.Port)
	}
	if len(pg.SSLMode) > 0 {
		s = fmt.Sprintf("%s sslmode=%s", s, pg.SSLMode)
	}
	if len(pg.FallbackApplicationName) > 0 {
		s = fmt.Sprintf("%s fallback_application_name=%s", s, pg.FallbackApplicationName)
	}
	if pg.ConnectTimeout > 0 {
		s = fmt.Sprintf("%s connect_timeout=%d", s, pg.ConnectTimeout)
	}
	if len(pg.SSLCert) > 0 {
		s = fmt.Sprintf("%s sslcert=%s", s, pg.SSLCert)
	}
	if len(pg.SSLKey) > 0 {
		s = fmt.Sprintf("%s sslkey=%s", s, pg.SSLKey)
	}
	if len(pg.SSLRootCert) > 0 {
		s = fmt.Sprintf("%s sslrootcert=%s", s, pg.SSLRootCert)
	}
	return
}

func (d *database) Close() error {
	// apcore
	d.hashPassForUserID.Close()
	d.userIdForEmail.Close()
	d.userIdForBoxPath.Close()
	d.userPreferences.Close()
	d.insertUserPolicy.Close()
	d.insertInstancePolicy.Close()
	d.updateUserPolicy.Close()
	d.updateInstancePolicy.Close()
	d.instancePolicies.Close()
	d.userPolicies.Close()
	d.userResolutions.Close()
	d.insertUserPKey.Close()
	d.getUserPKey.Close()
	// transport retries
	d.insertAttempt.Close()
	d.markSuccessfulAttempt.Close()
	d.markRetryFailureAttempt.Close()
	// oauth
	d.createTokenInfo.Close()
	d.removeTokenByCode.Close()
	d.removeTokenByAccess.Close()
	d.removeTokenByRefresh.Close()
	d.getTokenByCode.Close()
	d.getTokenByAccess.Close()
	d.getTokenByRefresh.Close()
	d.getClientById.Close()
	// go-fed
	d.inboxContains.Close()
	d.getInbox.Close()
	d.actorForOutbox.Close()
	d.actorForInbox.Close()
	d.outboxForInbox.Close()
	d.exists.Close()
	d.get.Close()
	d.localCreate.Close()
	d.fedCreate.Close()
	d.localUpdate.Close()
	d.fedUpdate.Close()
	d.localDelete.Close()
	d.fedDelete.Close()
	d.getOutbox.Close()
	d.followers.Close()
	d.following.Close()
	d.liked.Close()
	return d.db.Close()
}

func (d *database) Ping() error {
	return d.db.Ping()
}

func (d *database) Valid(c context.Context, userId, pass string) (valid bool, err error) {
	var r *sql.Rows
	r, err = d.hashPassForUserID.QueryContext(c, userId)
	if err != nil {
		return
	}
	defer r.Close()
	var n int
	hash := make([]byte, 0)
	salt := make([]byte, 0)
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when obtaining hash pass for user ID")
			return
		}
		if err = r.Scan(&hash, &salt); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	valid = passEquals(pass, salt, hash)
	return
}

func (d *database) UserIDFromEmail(c context.Context, email string) (userId string, err error) {
	var r *sql.Rows
	r, err = d.userIdForEmail.QueryContext(c, email)
	if err != nil {
		return
	}
	defer r.Close()
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when obtaining user id for email")
			return
		}
		if err = r.Scan(&userId); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) UserIdForBoxPath(c context.Context, boxPath string) (userId string, err error) {
	var r *sql.Rows
	r, err = d.userIdForBoxPath.QueryContext(c, boxPath)
	if err != nil {
		return
	}
	defer r.Close()
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when obtaining user id for box path")
			return
		}
		if err = r.Scan(&userId); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) UserPreferences(c context.Context, userId string) (u userPreferences, err error) {
	pu := &u
	var r *sql.Rows
	r, err = d.userPreferences.QueryContext(c, userId)
	if err != nil {
		return
	}
	defer r.Close()
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when obtaining user preferences")
			return
		}
		if err = pu.Load(r); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) InsertPolicy(c context.Context, p policy) (err error) {
	if p.IsInstancePolicy {
		_, err = d.insertInstancePolicy.ExecContext(c,
			p.Order,
			p.Description,
			p.Subject,
			p.Kind)
	} else {
		_, err = d.insertUserPolicy.ExecContext(c,
			p.Order,
			p.UserId,
			p.Description,
			p.Subject,
			p.Kind)
	}
	return
}

func (d *database) UpdatePolicy(c context.Context, p policy) (err error) {
	if p.IsInstancePolicy {
		_, err = d.updateInstancePolicy.ExecContext(c,
			p.Id,
			p.Order,
			p.Description,
			p.Subject,
			p.Kind)
	} else {
		_, err = d.updateUserPolicy.ExecContext(c,
			p.Id,
			p.Order,
			p.UserId,
			p.Description,
			p.Subject,
			p.Kind)
	}
	return
}

func (d *database) InstancePolicies(c context.Context) (p policies, err error) {
	var r *sql.Rows
	r, err = d.instancePolicies.QueryContext(c)
	if err != nil {
		return
	}
	defer r.Close()
	for r.Next() {
		var pol policy
		if err = pol.Load(r, true); err != nil {
			return
		}
		p = append(p, pol)
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) UserPolicies(c context.Context, userId string) (p policies, err error) {
	var r *sql.Rows
	r, err = d.userPolicies.QueryContext(c, userId)
	if err != nil {
		return
	}
	defer r.Close()
	for r.Next() {
		var pol policy
		if err = pol.Load(r, false); err != nil {
			return
		}
		p = append(p, pol)
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) InsertResolutions(c context.Context, r []resolution) (err error) {
	var tx *sql.Tx
	tx, err = d.db.BeginTx(c, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	for _, res := range r {
		_, err = tx.ExecContext(c,
			d.sqlgen.InsertResolutions(),
			res.TargetUserId,
			res.Permit,
			res.ActivityId.String(),
			res.Public,
			res.Reason)
		if err != nil {
			return
		}
	}
	err = tx.Commit()
	return
}

func (d *database) UserResolutions(c context.Context, userId string) (r []resolution, err error) {
	var rw *sql.Rows
	rw, err = d.userResolutions.QueryContext(c, userId)
	if err != nil {
		return
	}
	defer rw.Close()
	for rw.Next() {
		var res resolution
		if err = res.Load(rw); err != nil {
			return
		}
		r = append(r, res)
	}
	if err = rw.Err(); err != nil {
		return
	}
	return
}

func (d *database) InsertUserPKey(c context.Context, userUUID string, k *rsa.PrivateKey) (id int64, err error) {
	var pKeyB []byte
	pKeyB, err = x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return
	}
	var res sql.Result
	res, err = d.insertUserPKey.ExecContext(c, userUUID, pKeyB)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	return
}

func (d *database) GetUserPKey(c context.Context, userUUID string) (kUUID string, k *rsa.PrivateKey, err error) {
	var rw *sql.Rows
	rw, err = d.getUserPKey.QueryContext(c, userUUID)
	if err != nil {
		return
	}
	defer rw.Close()
	var pKeyB []byte
	var n int
	for rw.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when getting user private key: %s", userUUID)
			return
		}
		if err = rw.Scan(&kUUID, &pKeyB); err != nil {
			return
		}
		n++
	}
	if err = rw.Err(); err != nil {
		return
	}
	var pk crypto.PrivateKey
	pk, err = x509.ParsePKCS8PrivateKey(pKeyB)
	var ok bool
	k, ok = pk.(*rsa.PrivateKey)
	if !ok {
		err = fmt.Errorf("private key is not of type *rsa.PrivateKey")
		return
	}
	return
}

// apcore attempt functions

func (d *database) InsertAttempt(c context.Context, payload []byte, to *url.URL, fromUUID string) (id int64, err error) {
	var res sql.Result
	res, err = d.insertAttempt.ExecContext(c, payload, to, fromUUID)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	return
}

func (d *database) MarkSuccessfulAttempt(c context.Context, id int64) (err error) {
	_, err = d.markSuccessfulAttempt.ExecContext(c, id)
	return

}

func (d *database) MarkRetryFailureAttempt(c context.Context, id int64) (err error) {
	_, err = d.markRetryFailureAttempt.ExecContext(c, id)
	return
}

// apcore oauth functions

func (d *database) CreateTokenInfo(c context.Context, info oauth2.TokenInfo) error {
	_, err := d.createTokenInfo.ExecContext(
		c,
		info.GetClientID(),
		info.GetUserID(),
		info.GetRedirectURI(),
		info.GetScope(),
		info.GetCode(),
		info.GetCodeCreateAt(),
		info.GetCodeExpiresIn(),
		info.GetAccess(),
		info.GetAccessCreateAt(),
		info.GetAccessExpiresIn(),
		info.GetRefresh(),
		info.GetRefreshCreateAt(),
		info.GetRefreshExpiresIn())
	return err
}

func (d *database) RemoveTokenByCode(c context.Context, code string) error {
	_, err := d.removeTokenByCode.ExecContext(
		c,
		code)
	return err
}

func (d *database) RemoveTokenByAccess(c context.Context, access string) error {
	_, err := d.removeTokenByAccess.ExecContext(
		c,
		access)
	return err
}

func (d *database) RemoveTokenByRefresh(c context.Context, refresh string) error {
	_, err := d.removeTokenByRefresh.ExecContext(
		c,
		refresh)
	return err
}

func (d *database) mustScanRowsForOneToken(r *sql.Rows, ti *tokenInfo) (err error) {
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when obtaining OAuth2 token")
			return
		}
		if err = r.Scan(
			&ti.clientId,
			&ti.userId,
			&ti.redirectURI,
			&ti.scope,
			&ti.code,
			&ti.codeCreated,
			&ti.codeExpires,
			&ti.access,
			&ti.accessCreated,
			&ti.accessExpires,
			&ti.refresh,
			&ti.refreshCreated,
			&ti.refreshExpires); err != nil {
			return
		}
		n++
	}
	err = r.Err()
	return
}

func (d *database) GetTokenByCode(c context.Context, code string) (oti oauth2.TokenInfo, err error) {
	ti := &tokenInfo{}
	oti = ti
	var r *sql.Rows
	r, err = d.getTokenByCode.QueryContext(c, code)
	if err != nil {
		return
	}
	defer r.Close()
	err = d.mustScanRowsForOneToken(r, ti)
	return
}

func (d *database) GetTokenByAccess(c context.Context, access string) (oti oauth2.TokenInfo, err error) {
	ti := &tokenInfo{}
	oti = ti
	var r *sql.Rows
	r, err = d.getTokenByAccess.QueryContext(c, access)
	if err != nil {
		return
	}
	defer r.Close()
	err = d.mustScanRowsForOneToken(r, ti)
	return
}

func (d *database) GetTokenByRefresh(c context.Context, refresh string) (oti oauth2.TokenInfo, err error) {
	ti := &tokenInfo{}
	oti = ti
	var r *sql.Rows
	r, err = d.getTokenByRefresh.QueryContext(c, refresh)
	if err != nil {
		return
	}
	defer r.Close()
	err = d.mustScanRowsForOneToken(r, ti)
	return
}

func (d *database) GetClientById(c context.Context, id string) (oci oauth2.ClientInfo, err error) {
	ci := &clientInfo{}
	oci = ci
	var r *sql.Rows
	r, err = d.getClientById.QueryContext(c, id)
	if err != nil {
		return
	}
	defer r.Close()
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when obtaining OAuth2 client")
			return
		}
		if err = r.Scan(
			&ci.id,
			&ci.secret,
			&ci.domain,
			&ci.userId); err != nil {
			return
		}
		n++
	}
	err = r.Err()
	return
}

// go-fed ActivityPub implementation

func (d *database) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	var r *sql.Rows
	r, err = d.inboxContains.QueryContext(c, inbox.String(), id.String())
	if err != nil {
		return
	}
	defer r.Close()
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking inbox containment")
			return
		}
		if err = r.Scan(&contains); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	start := collectionPageStartIndex(inboxIRI)
	length := collectionPageLength(inboxIRI, d.defaultCollectionSize)
	var r *sql.Rows
	r, err = d.getInbox.QueryContext(c, inboxIRI.String(), start, length)
	if err != nil {
		return
	}
	defer r.Close()
	var iris []string
	for r.Next() {
		var iri string
		// unused
		var id int64
		var userId string
		var fedId string
		if err = r.Scan(&id, &userId, &fedId, &iri); err != nil {
			return
		}
		iris = append(iris, iri)
	}
	if err = r.Err(); err != nil {
		return
	}
	var id *url.URL
	id, err = collectionPageId(inboxIRI, start, length, d.defaultCollectionSize)
	if err != nil {
		return
	}
	inbox, err = toOrderedCollectionPage(id, iris, start, length)
	return
}

func (d *database) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// Step 1: Fetch existing data in the database
	iri, err := pub.GetId(inbox)
	if err != nil {
		return err
	}
	start := collectionPageStartIndex(iri)
	length := collectionPageLength(iri, d.defaultCollectionSize)
	tx, err := d.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	r, err := tx.QueryContext(c, d.sqlgen.GetInbox(), iri.String(), start, length)
	if err != nil {
		return err
	}
	defer r.Close()
	var fetched []struct {
		id     int64
		userId string
		fedId  string
		iri    string
	}
	for r.Next() {
		var elem struct {
			id     int64
			userId string
			fedId  string
			iri    string
		}
		if err = r.Scan(&elem.id, &elem.userId, &elem.fedId, &elem.iri); err != nil {
			return err
		}
		fetched = append(fetched, elem)
	}
	if err = r.Err(); err != nil {
		return err
	}
	r.Close() // Go ahead and close it

	// Step 2: Issue diff commands for the set
	oi := inbox.GetActivityStreamsOrderedItems()
	pos := start
	if oi != nil {
		// Update all items
		iter := oi.Begin()
		for iter != oi.End() && pos < start+len(fetched) {
			fedIri, err := pub.GetId(iter.GetType())
			if err != nil {
				return err
			}
			idx := pos - start
			_, err = tx.ExecContext(c, d.sqlgen.SetInboxUpdate(), fetched[idx].id, fetched[idx].userId, fedIri)
			if err != nil {
				return err
			}
			pos++
			iter = iter.Next()
		}
		// Add new items
		for iter != oi.End() {
			fedIri, err := pub.GetId(iter.GetType())
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(c, d.sqlgen.SetInboxInsert(), iri.String(), fedIri)
			if err != nil {
				return err
			}
			pos++
			iter = iter.Next()
		}
	}
	// Remove those in excess
	for pos < start+length && pos < start+len(fetched) {
		idx := pos - start
		_, err := tx.ExecContext(c, d.sqlgen.SetInboxDelete(), fetched[idx].id)
		if err != nil {
			return err
		}
	}

	// Step 3: Commit the transaction
	return tx.Commit()
}

func (d *database) Owns(c context.Context, id *url.URL) (owns bool, err error) {
	owns = id.Host == d.hostname
	return
}

func (d *database) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	var r *sql.Rows
	r, err = d.actorForOutbox.QueryContext(c, outboxIRI.String())
	if err != nil {
		return
	}
	var n int
	var iri string
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking actor for outbox")
			return
		}
		if err = r.Scan(&iri); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	actorIRI, err = url.Parse(iri)
	return
}

func (d *database) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	var r *sql.Rows
	r, err = d.actorForInbox.QueryContext(c, inboxIRI.String())
	if err != nil {
		return
	}
	var n int
	var iri string
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking actor for inbox")
			return
		}
		if err = r.Scan(&iri); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	actorIRI, err = url.Parse(iri)
	return
}

func (d *database) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	var r *sql.Rows
	r, err = d.outboxForInbox.QueryContext(c, inboxIRI.String())
	if err != nil {
		return
	}
	var n int
	var iri string
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking outbox for inbox")
			return
		}
		if err = r.Scan(&iri); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	outboxIRI, err = url.Parse(iri)
	return
}

func (d *database) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	var r *sql.Rows
	r, err = d.exists.QueryContext(c, id.String())
	if err != nil {
		return
	}
	var n int
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when checking exists")
			return
		}
		if err = r.Scan(&exists); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	return
}

func (d *database) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	var r *sql.Rows
	r, err = d.get.QueryContext(c, id.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when getting from db for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	value, err = streams.ToType(c, m)
	return
}

func (d *database) Create(c context.Context, asType vocab.Type) (err error) {
	var m map[string]interface{}
	m, err = streams.Serialize(asType)
	if err != nil {
		return
	}
	var b []byte
	b, err = json.Marshal(m)
	if err != nil {
		return
	}
	var id *url.URL
	id, err = pub.GetId(asType)
	if err != nil {
		return
	}
	var owns bool
	if owns, err = d.Owns(c, id); err != nil {
		return
	} else if owns {
		_, err = d.localCreate.ExecContext(c, string(b))
		return
	} else {
		_, err = d.fedCreate.ExecContext(c, string(b))
		return
	}
}

func (d *database) Update(c context.Context, asType vocab.Type) (err error) {
	var m map[string]interface{}
	m, err = streams.Serialize(asType)
	if err != nil {
		return
	}
	var b []byte
	b, err = json.Marshal(m)
	if err != nil {
		return
	}
	var id *url.URL
	id, err = pub.GetId(asType)
	if err != nil {
		return
	}
	var owns bool
	if owns, err = d.Owns(c, id); err != nil {
		return
	} else if owns {
		_, err = d.localUpdate.ExecContext(c, string(b))
		return
	} else {
		_, err = d.fedUpdate.ExecContext(c, string(b))
		return
	}
}

func (d *database) Delete(c context.Context, id *url.URL) (err error) {
	var owns bool
	if owns, err = d.Owns(c, id); err != nil {
		return
	} else if owns {
		_, err = d.localDelete.ExecContext(c, id.String())
		return
	} else {
		_, err = d.fedDelete.ExecContext(c, id.String())
		return
	}
}

func (d *database) GetOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	start := collectionPageStartIndex(outboxIRI)
	length := collectionPageLength(outboxIRI, d.defaultCollectionSize)
	var r *sql.Rows
	r, err = d.getOutbox.QueryContext(c, outboxIRI.String(), start, length)
	if err != nil {
		return
	}
	defer r.Close()
	var iris []string
	for r.Next() {
		var iri string
		// unused
		var id int64
		var userId string
		var fedId string
		if err = r.Scan(&id, &userId, &fedId, &iri); err != nil {
			return
		}
		iris = append(iris, iri)
	}
	if err = r.Err(); err != nil {
		return
	}
	var id *url.URL
	id, err = collectionPageId(outboxIRI, start, length, d.defaultCollectionSize)
	if err != nil {
		return
	}
	outbox, err = toOrderedCollectionPage(id, iris, start, length)
	return
}

func (d *database) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// Step 1: Fetch existing data in the database
	iri, err := pub.GetId(outbox)
	if err != nil {
		return err
	}
	start := collectionPageStartIndex(iri)
	length := collectionPageLength(iri, d.defaultCollectionSize)
	tx, err := d.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	r, err := tx.QueryContext(c, d.sqlgen.GetOutbox(), iri.String(), start, length)
	if err != nil {
		return err
	}
	defer r.Close()
	var fetched []struct {
		id     int64
		userId string
		fedId  string
		iri    string
	}
	for r.Next() {
		var elem struct {
			id     int64
			userId string
			fedId  string
			iri    string
		}
		if err = r.Scan(&elem.id, &elem.userId, &elem.fedId, &elem.iri); err != nil {
			return err
		}
		fetched = append(fetched, elem)
	}
	if err = r.Err(); err != nil {
		return err
	}
	r.Close() // Go ahead and close it

	// Step 2: Issue diff commands for the set
	oi := outbox.GetActivityStreamsOrderedItems()
	pos := start
	if oi != nil {
		// Update all items
		iter := oi.Begin()
		for iter != oi.End() && pos < start+len(fetched) {
			localIri, err := pub.GetId(iter.GetType())
			if err != nil {
				return err
			}
			idx := pos - start
			_, err = tx.ExecContext(c, d.sqlgen.SetOutboxUpdate(), fetched[idx].id, fetched[idx].userId, localIri)
			if err != nil {
				return err
			}
			pos++
			iter = iter.Next()
		}
		// Add new items
		for iter != oi.End() {
			localIri, err := pub.GetId(iter.GetType())
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(c, d.sqlgen.SetOutboxInsert(), iri.String(), localIri)
			if err != nil {
				return err
			}
			pos++
			iter = iter.Next()
		}
	}
	// Remove those in excess
	for pos < start+length && pos < start+len(fetched) {
		idx := pos - start
		_, err := tx.ExecContext(c, d.sqlgen.SetOutboxDelete(), fetched[idx].id)
		if err != nil {
			return err
		}
	}

	// Step 3: Commit the transaction
	return tx.Commit()
}

func (d *database) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	var r *sql.Rows
	r, err = d.followers.QueryContext(c, actorIRI.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when fetching followers for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	var res *streams.JSONResolver
	res, err = streams.NewJSONResolver(func(ctx context.Context, i vocab.ActivityStreamsCollection) error {
		followers = i
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, m)
	return
}

func (d *database) Following(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	var r *sql.Rows
	r, err = d.following.QueryContext(c, actorIRI.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when fetching following for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	var res *streams.JSONResolver
	res, err = streams.NewJSONResolver(func(ctx context.Context, i vocab.ActivityStreamsCollection) error {
		followers = i
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, m)
	return
}

func (d *database) Liked(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	var r *sql.Rows
	r, err = d.liked.QueryContext(c, actorIRI.String())
	if err != nil {
		return
	}
	var n int
	var jsonb []byte
	for r.Next() {
		if n > 0 {
			err = fmt.Errorf("multiple rows when fetching liked for IRI")
			return
		}
		if err = r.Scan(&jsonb); err != nil {
			return
		}
		n++
	}
	if err = r.Err(); err != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(jsonb, &m)
	if err != nil {
		return
	}
	var res *streams.JSONResolver
	res, err = streams.NewJSONResolver(func(ctx context.Context, i vocab.ActivityStreamsCollection) error {
		followers = i
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, m)
	return
}
