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

package ap

import (
	"github.com/go-fed/activity/pub"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/services"
)

func NewActorMap(c *config.Config,
	clock *Clock,
	db *Database,
	apdb *APDB,
	pk *services.PrivateKeys,
	f *services.Followers,
	tc *conn.Controller) (actorMap map[paths.Actor]pub.Actor) {
	actorMap = make(map[paths.Actor]pub.Actor, 1)
	actorMap[paths.InstanceActor] = newInstanceActor(c, clock, db, apdb, pk, f, tc)
	return
}

func newInstanceActor(c *config.Config,
	clock *Clock,
	db *Database,
	apdb *APDB,
	pk *services.PrivateKeys,
	f *services.Followers,
	tc *conn.Controller) (actor pub.Actor) {
	common := newInstanceActorCommonBehavior(db, tc, pk)
	s2s := newInstanceActorFederatingBehavior(c, db, pk, f, tc)
	actor = pub.NewFederatingActor(common, s2s, apdb, clock)
	return
}
