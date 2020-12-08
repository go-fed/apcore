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

package nodeinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fed/apcore/app"
	srv "github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

// This file contains the NodeInfo v2.1 implementation.
//
// NodeInfo is infamous for being uncompromising in its dictatorial content
// requirements of the fields presented below.

const (
	nodeInfoVersion        = "2.1"
	nodeInfoWellKnownPath  = "/.well-known/nodeinfo"
	nodeInfoPath           = "/nodeinfo/" + nodeInfoVersion
	validSoftwareNameChars = "abcdefghijklmnopqrstuvwxyz0123456789-"
)

type nodeInfo struct {
	Version           string                 `json:"version"`
	Software          software               `json:"software"`
	Protocols         []string               `json:"protocols"`
	Services          services               `json:"services"`
	OpenRegistrations bool                   `json:"openRegistrations"`
	Usage             usage                  `json:"usage"`
	Metadata          map[string]interface{} `json:"metadata"`
}

type software struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

type services struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

type usage struct {
	Users         users `json:"users"`
	LocalPosts    int   `json:"localPosts"`
	LocalComments int   `json:"localComments"`
}

type users struct {
	Total          int `json:"total"`
	ActiveHalfYear int `json:"activeHalfyear"`
	ActiveMonth    int `json:"activeMonth"`
}

func sanitizeSoftwareName(name string) string {
	// 1. Lower case everything
	name = strings.ToLower(name)
	// 2. Strip anything not in the valid charset
	var b strings.Builder
	for _, r := range name {
		if !strings.ContainsAny(string(r), validSoftwareNameChars) {
			continue
		}
		b.WriteRune(r)
	}
	// Ship it.
	return b.String()
}

func toNodeInfo(s, apcore app.Software, t *srv.NodeInfoStats, p srv.ServerPreferences) nodeInfo {
	n := nodeInfo{
		Version: nodeInfoVersion,
		Software: software{
			Name:       sanitizeSoftwareName(s.Name),
			Version:    s.Version(),
			Repository: s.Repository,
		},
		Protocols: []string{"activitypub"},
		Services: services{
			Inbound:  []string{},
			Outbound: []string{},
		},
		OpenRegistrations: p.OpenRegistrations,
		Metadata: map[string]interface{}{
			apcore.Name: map[string]interface{}{
				"version":    apcore.Version(),
				"repository": apcore.Repository,
			},
		},
	}
	if t != nil {
		n.Usage = usage{
			Users: users{
				Total:          t.TotalUsers,
				ActiveHalfYear: t.ActiveHalfYear,
				ActiveMonth:    t.ActiveMonth,
			},
			LocalPosts:    t.NLocalPosts,
			LocalComments: t.NLocalComments,
		}
	}
	return n
}

func nodeInfoWellKnownHandler(scheme, host string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/jrd+json")
		var b bytes.Buffer
		b.WriteString(`{"links":[{"rel": "http://nodeinfo.diaspora.software/ns/schema/2.1","href": "`)
		b.WriteString(scheme)
		b.WriteString(`://`)
		b.WriteString(host)
		b.WriteString(nodeInfoPath)
		b.WriteString(`"}]}`)
		bt := b.Bytes()
		n, err := w.Write(bt)
		if err != nil {
			util.ErrorLogger.Errorf("error writing well-known nodeinfo response: %s", err)
		} else if n != len(bt) {
			util.ErrorLogger.Errorf("error writing well-known nodeinfo response: wrote %d of %d bytes", n, len(bt))
		}
	}
}

func nodeInfoHandler(ni *srv.NodeInfo, u *srv.Users, s, apcore app.Software, useStats bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", `application/json; profile="http://nodeinfo.diaspora.software/ns/schema/2.1#"`)

		ctx := util.Context{r.Context()}
		var t *srv.NodeInfoStats
		if useStats {
			st, err := ni.GetAnonymizedStats(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("error serving nodeinfo response"), http.StatusInternalServerError)
				util.ErrorLogger.Errorf("error in getting anonymized stats for nodeinfo response: %s", err)
				return
			}
			t = &st
		}

		p, err := u.GetServerPreferences(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("error serving nodeinfo response"), http.StatusInternalServerError)
			util.ErrorLogger.Errorf("error in getting server profile for nodeinfo response: %s", err)
			return
		}

		ni := toNodeInfo(s, apcore, t, p)
		b, err := json.Marshal(ni)
		if err != nil {
			http.Error(w, fmt.Sprintf("error serving nodeinfo response"), http.StatusInternalServerError)
			util.ErrorLogger.Errorf("error marshalling nodeinfo response to JSON: %s", err)
			return
		}

		n, err := w.Write(b)
		if err != nil {
			util.ErrorLogger.Errorf("error writing nodeinfo response: %s", err)
		} else if n != len(b) {
			util.ErrorLogger.Errorf("error writing nodeinfo response: wrote %d of %d bytes", n, len(b))
		}
	}
}
