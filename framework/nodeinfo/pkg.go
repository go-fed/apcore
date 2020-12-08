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

// There are two competing standards: NodeInfo and NodeInfo2. Confusingly,
// NodeInfo is at version 2, so it is NodeInfo 2 vs NodeInfo2.
//
// I have no idea about the background of either. But it is late, and I find
// myself tired and, having thought this would be a quick serialize-to-JSON
// implementation, massively disappointed at what I've found instead.
//
// So, support both, I guess. Thank God, I was worried I would one day find
// myself upon my deathbed and staring into the eyes of my loved ones and
// quietly whisper in the soft exhale of my last breath "if only I could have
// implemented another NodeInfo standard, I would have had a fulfilling life".
// So my survivors would all turn to each other teary-eyed and put on my
// gravestone "NodeInfo 2" with an ambiguously-sized space in-between.
//
// If at least this means one other person doesn't have to deal with this
// dual headache, then I guess it was indeed worth it.
//
// I guess this is what happens when people can't get along and at least one
// person is being uncompromising. A lesson for me to avoid this at all costs.
package nodeinfo

import (
	"net/http"

	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	srv "github.com/go-fed/apcore/services"
)

type PathHandler struct {
	Path    string
	Handler http.HandlerFunc
}

func GetNodeInfoHandlers(c config.NodeInfoConfig, scheme, host string, ni *srv.NodeInfo, u *srv.Users, s, apcore app.Software) []PathHandler {
	var ph []PathHandler
	if c.EnableNodeInfo {
		ph = append(ph, PathHandler{
			Path:    nodeInfoWellKnownPath,
			Handler: nodeInfoWellKnownHandler(scheme, host),
		})
		ph = append(ph, PathHandler{
			Path:    nodeInfoPath,
			Handler: nodeInfoHandler(ni, u, s, apcore, c.EnableAnonymousStatsSharing),
		})
	}
	if c.EnableNodeInfo2 {
		ph = append(ph, PathHandler{
			Path:    nodeInfo2WellKnownPath,
			Handler: nodeInfo2WellKnownHandler(ni, u, s, apcore, c.EnableAnonymousStatsSharing),
		})
	}
	return ph
}
