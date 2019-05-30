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
	"fmt"

	"github.com/go-fed/activity/pub"
)

func newActor(c *config, a Application, db *database) (actor pub.Actor, err error) {
	var clock *clock
	clock, err = newClock(c.ActivityPubConfig.ClockTimezone)
	if err != nil {
		return
	}

	common := newCommonBehavior(db)
	apdb := newApdb(db, a)

	if cs, ss := a.C2SEnabled(), a.S2SEnabled(); !cs && !ss {
		err = fmt.Errorf("neither C2S nor S2S are enabled by the Application")
	} else if cs && ss {
		c2s := newSocialBehavior(db)
		s2s := newFederatingBehavior(db)
		actor = pub.NewActor(
			common,
			c2s,
			s2s,
			apdb,
			clock)
	} else if cs {
		c2s := newSocialBehavior(db)
		actor = pub.NewSocialActor(
			common,
			c2s,
			apdb,
			clock)
	} else if ss {
		s2s := newFederatingBehavior(db)
		actor = pub.NewFederatingActor(
			common,
			s2s,
			apdb,
			clock)
	}
	return
}
