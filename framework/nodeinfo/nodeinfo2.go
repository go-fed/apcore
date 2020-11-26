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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-fed/apcore/app"
	srv "github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

// This file contains the NodeInfo2 v1 implementation.

const (
	nodeInfo2Version       = "1.0"
	nodeInfo2WellKnownPath = "/.well-known/x-nodeinfo2"
)

type nodeInfo2 struct {
	Version           string        `json:"version"`
	Server            server2       `json:"server"`
	Organization      organization2 `json:"organization"`
	Protocols         []string      `json:"protocols"`
	Services          services2     `json:"services"`
	OpenRegistrations bool          `json:"openRegistrations"`
	Usage             usage2        `json:"usage"`
	Relay             string        `json:"relay"`
	OtherFeatures     []feature2    `json:"otherFeatures"`
}

type server2 struct {
	BaseURL  string `json:"baseUrl"`
	Name     string `json:"name"`
	Software string `json:"software"`
	Version  string `json:"version"`
}

type organization2 struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Account string `json:"account"`
}

type services2 struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

type usage2 struct {
	Users         users2 `json:"users"`
	LocalPosts    int    `json:"localPosts"`
	LocalComments int    `json:"localComments"`
}

type users2 struct {
	Total          int `json:"total"`
	ActiveHalfYear int `json:"activeHalfyear"`
	ActiveMonth    int `json:"activeMonth"`
	ActiveWeek     int `json:"activeWeek"`
}

type feature2 struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func toNodeInfo2(s, apcore app.Software, t srv.NodeInfoStats, p srv.ServerProfile) nodeInfo2 {
	n := nodeInfo2{
		Version: nodeInfo2Version,
		Server: server2{
			BaseURL:  p.ServerBaseURL,
			Name:     p.ServerName,
			Software: s.Name,
			Version:  s.Version(),
		},
		Organization: organization2{
			Name:    p.OrgName,
			Contact: p.OrgContact,
			Account: p.OrgAccount,
		},
		Protocols: []string{"activitypub"},
		Services: services2{
			Inbound:  []string{},
			Outbound: []string{},
		},
		OpenRegistrations: p.OpenRegistrations,
		Usage: usage2{
			Users: users2{
				Total:          t.TotalUsers,
				ActiveHalfYear: t.ActiveHalfYear,
				ActiveMonth:    t.ActiveMonth,
				ActiveWeek:     t.ActiveWeek,
			},
			LocalPosts:    t.NLocalPosts,
			LocalComments: t.NLocalComments,
		},
		Relay: "",
		OtherFeatures: []feature2{
			{
				Name:    apcore.Name,
				Version: apcore.Version(),
			},
		},
	}
	return n
}

func nodeInfo2WellKnownHandler(ni *srv.NodeInfo, s, apcore app.Software) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", `application/json`)

		t, err := ni.GetAnonymizedStats()
		if err != nil {
			http.Error(w, fmt.Sprintf("error serving nodeinfo2 response"), http.StatusInternalServerError)
			util.ErrorLogger.Errorf("error in getting anonymized stats for nodeinfo2 response: %s", err)
			return
		}

		p, err := ni.GetServerProfile()
		if err != nil {
			http.Error(w, fmt.Sprintf("error serving nodeinfo2 response"), http.StatusInternalServerError)
			util.ErrorLogger.Errorf("error in getting server profile for nodeinfo2 response: %s", err)
			return
		}

		ni := toNodeInfo2(s, apcore, t, p)
		b, err := json.Marshal(ni)
		if err != nil {
			http.Error(w, fmt.Sprintf("error serving nodeinfo2 response"), http.StatusInternalServerError)
			util.ErrorLogger.Errorf("error marshalling nodeinfo2 response to JSON: %s", err)
			return
		}

		n, err := w.Write(b)
		if err != nil {
			util.ErrorLogger.Errorf("error writing nodeinfo2 response: %s", err)
		} else if n != len(b) {
			util.ErrorLogger.Errorf("error writing nodeinfo2 response: wrote %d of %d bytes", n, len(b))
		}
	}
}
