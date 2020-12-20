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

package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/web"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
	"github.com/go-fed/oauth2"
	oaerrors "github.com/go-fed/oauth2/errors"
	"github.com/go-fed/oauth2/generates"
	"github.com/go-fed/oauth2/manage"
	oam "github.com/go-fed/oauth2/models"
	oaserver "github.com/go-fed/oauth2/server"
)

const (
	authCodeExp = time.Second * 5
)

type Server struct {
	d *services.OAuth2
	y *services.Crypto
	k *web.Sessions
	m *manage.Manager
	s *oaserver.Server
	// First-party support:
	clientIDBase                string
	host                        string
	scheme                      string
	accessGen                   *generates.AccessGenerate
	accessExpiryDuration        time.Duration
	refreshExpiryDuration       time.Duration
	proxyRefreshAccessDuration  time.Duration
	proxyRefreshRefreshDuration time.Duration
	cleanupFn                   *util.SafeStartStop
}

func NewServer(c *config.Config, scheme string, internalErrorHandler http.Handler, d *services.OAuth2, y *services.Crypto, k *web.Sessions) (s *Server, err error) {
	m := manage.NewDefaultManager()
	// Configure Access token and Refresh token refresh.
	if c.OAuthConfig.AccessTokenExpiry <= 0 {
		err = fmt.Errorf("oauth2 access token expiration duration is <= 0")
		return
	} else if c.OAuthConfig.RefreshTokenExpiry <= 0 {
		err = fmt.Errorf("oauth2 refresh token expiration duration is <= 0")
		return
	}
	m.SetAuthorizeCodeExp(authCodeExp)
	m.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Second * time.Duration(c.OAuthConfig.AccessTokenExpiry),
		RefreshTokenExp:   time.Second * time.Duration(c.OAuthConfig.RefreshTokenExpiry),
		IsGenerateRefresh: true,
	})
	m.SetRefreshTokenCfg(&manage.RefreshingConfig{
		AccessTokenExp:  time.Second * time.Duration(c.OAuthConfig.AccessTokenExpiry),
		RefreshTokenExp: time.Second * time.Duration(c.OAuthConfig.RefreshTokenExpiry),
		// Generate new refresh token
		IsGenerateRefresh: true,
		// Remove previous access token
		IsRemoveAccess: true,
		// Remove previous refreshing token
		IsRemoveRefreshing: true,
	})
	m.MapTokenStorage(d)
	m.MapClientStorage(d)
	// OAuth2 server: PKCE + Authorization Code
	srv := oaserver.NewServer(&oaserver.Config{
		TokenType: "Bearer",
		// Must follow the spec.
		AllowGetAccessRequest: false,
		// Support only the non-implicit flow.
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code},
		// Allow:
		// - Authorization Code (for first & third parties)
		// - Refreshing Tokens
		//
		// Deny:
		// - Resource owner secrets (password grant)
		// - Client secrets
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
		},
	}, m)
	// Determines the user to use when granting an authorization token. If
	// no user is present, then they have not yet logged in and need to do
	// so. Note that an empty string userID plus no error will magically
	// cause the library to stop processing.
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		var s *web.Session
		if s, err = k.Get(r); err != nil {
			util.ErrorLogger.Errorf("error getting session in OAuth2 SetUserAuthorizationHandler: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if userID, err = s.UserID(); err != nil {
			// User is not logged in; redirect to login page with current
			// set of query parameters for OAuth2.
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
			return
		}
		// User is already logged in
		return
	})
	srv.SetInternalErrorHandler(func(err error) (re *oaerrors.Response) {
		re = &oaerrors.Response{
			Error:       oaerrors.ErrServerError,
			ErrorCode:   http.StatusInternalServerError,
			Description: "Internal Error",
			StatusCode:  http.StatusInternalServerError,
		}
		return
	})
	srv.SetResponseErrorHandler(func(re *oaerrors.Response) {
		util.ErrorLogger.Errorf("oauth2 response error: %s", re.Error.Error())
	})
	b64ClientPart := base64.RawStdEncoding.
		WithPadding(base64.NoPadding).
		EncodeToString([]byte((&url.URL{
			Scheme: scheme,
			Host:   c.ServerConfig.Host,
			Path:   "/",
		}).String()))
	s = &Server{
		d:                           d,
		y:                           y,
		k:                           k,
		m:                           m,
		s:                           srv,
		clientIDBase:                fmt.Sprintf("%s.%s", b64ClientPart, c.ServerConfig.Host),
		host:                        c.ServerConfig.Host,
		scheme:                      scheme,
		accessGen:                   generates.NewAccessGenerate(),
		accessExpiryDuration:        time.Second * time.Duration(c.OAuthConfig.AccessTokenExpiry),
		refreshExpiryDuration:       time.Second * time.Duration(c.OAuthConfig.RefreshTokenExpiry),
		proxyRefreshAccessDuration:  time.Second * time.Duration(c.OAuthConfig.AccessTokenExpiry) / 2,
		proxyRefreshRefreshDuration: time.Second * time.Duration(c.OAuthConfig.RefreshTokenExpiry) / 2,
	}
	s.cleanupFn = util.NewSafeStartStop(s.cleanup, time.Hour*1)
	return
}

// TODO: Scopes

func (o *Server) HandleAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	if err := o.s.HandleAuthorizeRequest(w, r); err != nil {
		// oauth2 library would already have written headers by now.
		util.ErrorLogger.Errorf("oauth2 HandleAuthorizeRequest error: %s", err)
	}
}

func (o *Server) HandleAccessTokenRequest(w http.ResponseWriter, r *http.Request) {
	if err := o.s.HandleTokenRequest(w, r); err != nil {
		// oauth2 library would already have written headers by now.
		util.ErrorLogger.Errorf("oauth2 HandleTokenRequest error: %s", err)
	}
}

func (o *Server) ValidateOAuth2AccessToken(w http.ResponseWriter, r *http.Request) (token oauth2.TokenInfo, authenticated bool, err error) {
	token, err = o.s.ValidationBearerToken(r)
	authenticated = err == nil
	if err == oaerrors.ErrInvalidAccessToken {
		authenticated = false
		err = nil
	}
	return
}

func (o *Server) RemoveByAccess(ctx util.Context, t oauth2.TokenInfo) error {
	return o.m.RemoveAccessToken(ctx.Context, t.GetAccess())
}

func (o *Server) CreateProxyCredentials(ctx util.Context, userID string) (id string, err error) {
	now := time.Now()
	var clientID string
	clientID, err = o.generateProxyClientID()
	if err != nil {
		return
	}
	ti := &oam.Token{
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         (&url.URL{Scheme: o.scheme, Host: o.host, Path: "/"}).String(),
		Scope:               "all", // TODO: Hardcoded scope here
		Code:                "",
		CodeCreateAt:        now,
		CodeExpiresIn:       0,
		CodeChallenge:       "",
		CodeChallengeMethod: "",
		AccessCreateAt:      now,
		AccessExpiresIn:     o.accessExpiryDuration,
		RefreshCreateAt:     now,
		RefreshExpiresIn:    o.refreshExpiryDuration,
	}
	data := &oauth2.GenerateBasic{
		Client: &oam.Client{
			ID:     ti.ClientID,
			Domain: o.host,
		},
		UserID:   ti.UserID,
		CreateAt: now,
	}
	ti.Access, ti.Refresh, err = o.accessGen.Token(ctx.Context, data, true)
	return o.d.ProxyCreateCredential(ctx, ti)
}

func (o *Server) RefreshProxyCredentialsIfNeeded(ctx util.Context, id, userID string) error {
	now := time.Now()

	// Ensure the refresh request is valid
	ti, err := o.d.ProxyGetCredential(ctx, id)
	if err != nil {
		return err
	} else if !strings.Contains(ti.GetClientID(), o.clientIDBase) {
		return oaerrors.ErrInvalidRefreshToken
	} else if ti.GetUserID() != userID {
		return oaerrors.ErrInvalidRefreshToken
	}

	// Do not refresh if there's plenty of time left, or if the access token
	// never expires.
	aei := ti.GetAccessExpiresIn()
	aca := ti.GetAccessCreateAt()
	rei := ti.GetRefreshExpiresIn()
	rca := ti.GetRefreshCreateAt()
	if aei == 0 || now.After(aca.Add(aei)) || aca.Add(o.proxyRefreshAccessDuration).After(now) {
		return nil
	} else if rei == 0 || now.After(rca.Add(rei)) || rca.Add(o.proxyRefreshRefreshDuration).After(now) {
		return nil
	}

	ti.SetAccessCreateAt(now)
	ti.SetAccessExpiresIn(o.accessExpiryDuration)
	ti.SetRefreshCreateAt(now)
	ti.SetRefreshExpiresIn(o.refreshExpiryDuration)
	data := &oauth2.GenerateBasic{
		Client: &oam.Client{
			ID:     ti.GetClientID(),
			Domain: o.host,
		},
		UserID:   ti.GetUserID(),
		CreateAt: now,
	}
	acc, ref, err := o.accessGen.Token(ctx.Context, data, true)
	if err != nil {
		return err
	}
	ti.SetAccess(acc)
	ti.SetRefresh(ref)
	return o.d.ProxyUpdateCredential(ctx, id, ti)
}

func (o *Server) ValidateFirstPartyProxyAccessToken(ctx util.Context, sn *web.Session) (id string, authenticated bool, err error) {
	if sn.HasFirstPartyCredentialID() {
		id, err = sn.FirstPartyCredentialID()
		if err != nil {
			return
		}
		now := time.Now()
		var ti oauth2.TokenInfo
		ti, err = o.d.ProxyGetCredential(ctx, id)
		if err != nil {
			return
		} else if ti == nil {
			err = fmt.Errorf("invalid first party credential token")
			return
		} else if ti.GetRefresh() != "" && ti.GetRefreshExpiresIn() != 0 &&
			ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(now) {
			err = fmt.Errorf("refresh token is expired")
			return
		} else if ti.GetAccessExpiresIn() != 0 &&
			ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(now) {
			err = fmt.Errorf("access token is expired")
			return
		}
		authenticated = true
	}
	return
}

func (o *Server) RemoveFirstPartyProxyAccessToken(w http.ResponseWriter, r *http.Request, ctx util.Context, sn *web.Session) error {
	id, auth, err := o.ValidateFirstPartyProxyAccessToken(ctx, sn)
	if err != nil {
		return err
	}
	if auth {
		if err = o.d.ProxyRemoveCredential(r.Context(), id); err != nil {
			return err
		}
		sn.DeleteFirstPartyCredentialID()
		sn.DeleteUserID()
		if err = sn.Save(r, w); err != nil {
			return err
		}
	}
	return nil
}

func (o *Server) Start() {
	o.cleanupFn.Start()
}

func (o *Server) Stop() {
	o.cleanupFn.Stop()
}

func (o *Server) cleanup(ctx context.Context) {
	err := o.d.DeleteExpiredFirstPartyCredentials(ctx)
	if err != nil {
		util.ErrorLogger.Errorf("first party expired creds cleanup failed: %s", err)
		return
	}
}

func (o *Server) generateProxyClientID() (string, error) {
	b := make([]byte, 256)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	b64SecretPart := base64.RawStdEncoding.
		WithPadding(base64.NoPadding).
		EncodeToString(b)
	return fmt.Sprintf("%s.%s", b64SecretPart, o.clientIDBase), nil
}
