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
	"fmt"
	"net/http"
	"time"

	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/web"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
	"github.com/go-oauth2/oauth2/v4"
	oaerrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	oaserver "github.com/go-oauth2/oauth2/v4/server"
)

type Server struct {
	d *services.OAuth2
	y *services.Crypto
	k *web.Sessions
	m *manage.Manager
	s *oaserver.Server
}

func NewServer(c *config.Config, a app.Application, d *services.OAuth2, y *services.Crypto, k *web.Sessions) (s *Server, err error) {
	m := manage.NewDefaultManager()
	// Configure Access token and Refresh token refresh.
	if c.OAuthConfig.AccessTokenExpiry <= 0 {
		err = fmt.Errorf("oauth2 access token expiration duration is <= 0")
		return
	} else if c.OAuthConfig.RefreshTokenExpiry <= 0 {
		err = fmt.Errorf("oauth2 refresh token expiration duration is <= 0")
		return
	}
	m.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Second * time.Duration(c.OAuthConfig.AccessTokenExpiry),
		RefreshTokenExp:   time.Second * time.Duration(c.OAuthConfig.RefreshTokenExpiry),
		IsGenerateRefresh: true,
	})
	m.SetRefreshTokenCfg(&manage.RefreshingConfig{
		// Generate new refresh token
		IsGenerateRefresh: true,
		// Remove previous access token
		IsRemoveAccess: true,
		// Remove previous refreshing token
		IsRemoveRefreshing: true,
	})
	m.MapTokenStorage(d)
	m.MapClientStorage(d)
	// OAuth2 server
	srv := oaserver.NewServer(&oaserver.Config{
		TokenType: "Bearer",
		// Must follow the spec.
		AllowGetAccessRequest: false,
		// Support only the non-implicit flow.
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code},
		// Allow:
		// - Authorization Code (for third parties)
		// - Refreshing Tokens
		// - Client secrets
		//
		// Deny:
		// - Resource owner secrets (password grant)
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
			oauth2.ClientCredentials,
		},
	}, m)
	// Parse tokens in POST body.
	srv.SetClientInfoHandler(oaserver.ClientFormHandler)
	// Determines the user to use when granting an authorization token. If
	// no user is present, then they have not yet logged in and need to do
	// so. Note that an empty string userID plus no error will magically
	// cause the library to stop processing.
	badRequestHandler := a.BadRequestHandler()
	internalErrorHandler := a.InternalServerErrorHandler()
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		var s *web.Session
		if s, err = k.Get(r); err != nil {
			util.ErrorLogger.Errorf("error getting session in OAuth2 SetUserAuthorizationHandler: %s", err)
			internalErrorHandler.ServeHTTP(w, r)
			return
		}
		if userID, err = s.UserID(); err != nil {
			if r.Form == nil {
				err = r.ParseForm()
				if err != nil {
					badRequestHandler.ServeHTTP(w, r)
					return
				}
			}
			s.SetOAuthRedirectFormValues(r.Form)
			err = s.Save(r, w)
			if err != nil {
				util.ErrorLogger.Errorf("error saving session in OAuth2 SetUserAuthorizationHandler: %s", err)
				internalErrorHandler.ServeHTTP(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// User is already logged in
		return
	})
	// Called when requesting a token through the password credential grant
	// flow.
	//
	// NOTE: This grant type is currently not supported, but the handler is
	// here if it were to be enabled.
	srv.SetPasswordAuthorizationHandler(func(email, password string) (userID string, err error) {
		// TODO: Fix oauth2 to support request contexts.
		var valid bool
		userID, valid, err = y.Valid(util.Context{context.Background()}, email, password)
		if err != nil {
			return
		} else if !valid {
			err = fmt.Errorf("username and/or password is invalid")
			return
		}
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
	s = &Server{
		d: d,
		y: y,
		k: k,
		m: m,
		s: srv,
	}
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
