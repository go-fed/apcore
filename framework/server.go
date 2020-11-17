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

package framework

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/framework/db"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

type Server struct {
	certFile    string
	keyFile     string
	a           app.Application
	sqldb       *sql.DB
	d           models.SqlDialect
	models      []models.Model
	tc          *conn.Controller
	httpServer  *http.Server
	httpsServer *http.Server
	debug       bool // TODO: http only, no https
}

func NewServer(c *config.Config, h http.Handler, scheme string, a app.Application, sqldb *sql.DB, d models.SqlDialect, models []models.Model, tc *conn.Controller) (s *Server, err error) {
	// TODO: Move to config validator
	const minKeySize = 1024
	// Enforce server level configuration
	if c.ServerConfig.RSAKeySize < minKeySize {
		err = fmt.Errorf("RSA private key size is configured to be < %d, which is forbidden: %d", minKeySize, c.ServerConfig.RSAKeySize)
		return
	}

	// Prepare HTTPS server. No option to run the server as HTTP in prod,
	// because we're living in the future.
	httpsServer := &http.Server{
		Addr:         ":" + scheme,
		Handler:      h,
		ReadTimeout:  time.Duration(c.ServerConfig.HttpsReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(c.ServerConfig.HttpsWriteTimeoutSeconds) * time.Second,
		TLSConfig:    createTlsConfig(),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	// Prepare redirection server. HTTP is only allowed to redirect.
	httpServer := createRedirectServer(c)

	// Create the apcore server
	s = &Server{
		certFile:    c.ServerConfig.CertFile,
		keyFile:     c.ServerConfig.KeyFile,
		a:           a,
		sqldb:       sqldb,
		d:           d,
		tc:          tc,
		httpServer:  httpServer,
		httpsServer: httpsServer,
	}

	// Post-creation hooks
	httpsServer.RegisterOnShutdown(s.onStop)
	return
}

// Do not let clients downgrade connections to use insecure, older
// cryptographic functions or curves.
func createTlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}

func createRedirectServer(c *config.Config) *http.Server {
	return &http.Server{
		Addr:         ":http",
		ReadTimeout:  time.Duration(c.ServerConfig.RedirectReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(c.ServerConfig.RedirectWriteTimeoutSeconds) * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			http.Redirect(w, req, fmt.Sprintf("https://%s%s", c.ServerConfig.Host, req.URL), http.StatusMovedPermanently)
		}),
	}
}

func (s *Server) Start() error {
	err := db.MustPing(s.sqldb)
	if err != nil {
		return err
	}
	util.InfoLogger.Infof("Preparing models")
	for _, m := range s.models {
		if err := m.Prepare(s.sqldb, s.d); err != nil {
			return err
		}
	}
	util.InfoLogger.Infof("Starting conn.Controller")
	s.tc.Start()
	util.InfoLogger.Infof("Starting application")
	err = s.a.Start()
	if err != nil {
		return err
	}
	go func() {
		util.InfoLogger.Infof("Starting http redirection server")
		err := s.httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			util.ErrorLogger.Errorf("Error shutting down http redirect server: %s", err)
		} else {
			util.InfoLogger.Infof("Http redirect server shutdown")
		}
	}()
	util.InfoLogger.Infof("Launching https server")
	err = s.httpsServer.ListenAndServeTLS(
		s.certFile,
		s.keyFile)
	if err != http.ErrServerClosed {
		util.ErrorLogger.Errorf("Error shutting down https server: %s", err)
	} else {
		util.InfoLogger.Infof("HTTPS server shutdown")
	}
	return nil
}

func (s *Server) Stop() {
	util.InfoLogger.Infof("Shutdown HTTPS server")
	s.httpsServer.Shutdown(context.Background())
}

func (s *Server) onStop() {
	util.InfoLogger.Infof("Shutdown HTTP server")
	s.httpServer.Shutdown(context.Background())
	util.InfoLogger.Infof("Stop application")
	if err := s.a.Stop(); err != nil {
		util.ErrorLogger.Errorf("Error shutting down application: %s", err)
	}
	util.InfoLogger.Infof("Stopping conn.Controller")
	s.tc.Stop()
	util.InfoLogger.Infof("Closing models")
	for _, m := range s.models {
		m.Close()
	}
}
