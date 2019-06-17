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
	"net/http"
	"time"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	oaserver "gopkg.in/oauth2.v3/server"
)

type oAuth2Server struct {
	m *manage.Manager
	s *oaserver.Server
}

func newOAuth2Server(c *config) (s *oAuth2Server, err error) {
	m := manage.NewDefaultManager()
	// Configure Access token and Refresh token refresh.
	m.SetAuthorizeCodeTokenCfg(&manage.Config{
		// TODO: Configurable time
		AccessTokenExp:    time.Hour * 2,
		RefreshTokenExp:   time.Hour * 24 * 3,
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
	m.MapTokenStorage( /*TODO*/ nil)
	m.MapClientStorage( /*TODO*/ nil)
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
		// - Resource owner secrets
		// - Client secrets
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
			oauth2.PasswordCredentials,
			oauth2.ClientCredentials,
		},
	}, m)
	// Parse tokens in POST body.
	srv.SetClientInfoHandler(oaserver.ClientFormHandler)
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		re = &errors.Response{
			Error:       errors.ErrServerError,
			ErrorCode:   http.StatusInternalServerError,
			Description: "Internal Error",
			StatusCode:  http.StatusInternalServerError,
		}
		return
	})
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		ErrorLogger.Errorf("OAuth2 response error: %s", re.Error.Error())
	})
	s = &oAuth2Server{
		m: m,
		s: srv,
	}
	return
}

func (o *oAuth2Server) HandleAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	if err := o.s.HandleAuthorizeRequest(w, r); err != nil {
		// oauth2 library would already have written headers by now.
		ErrorLogger.Errorf("OAuth2 HandleAuthorizeRequest error: %s", err)
	}
}

func (o *oAuth2Server) HandleAccessTokenRequest(w http.ResponseWriter, r *http.Request) {
	if err := o.s.HandleTokenRequest(w, r); err != nil {
		// oauth2 library would already have written headers by now.
		ErrorLogger.Errorf("OAuth2 HandleTokenRequest error: %s", err)
	}
}
