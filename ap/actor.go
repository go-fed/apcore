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
	"fmt"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/framework/oauth2"
	"github.com/go-fed/apcore/services"
)

func NewActor(c *config.Config,
	a app.Application,
	clock *Clock,
	db *Database,
	apdb *APDB,
	o *oauth2.Server,
	pk *services.PrivateKeys,
	po *services.Policies,
	f *services.Followers,
	u *services.Users,
	tc *conn.Controller) (actor pub.Actor, err error) {

	common := NewCommonBehavior(a, db, tc, o, pk)
	ca, isC2S := a.(app.C2SApplication)
	sa, isS2S := a.(app.S2SApplication)
	if !isC2S && !isS2S {
		err = fmt.Errorf("the Application is neither a C2SApplication nor a S2SApplication")
	} else if isC2S && isS2S {
		c2s := NewSocialBehavior(ca, o)
		s2s := NewFederatingBehavior(c, sa, db, po, pk, f, u, tc)
		actor = pub.NewActor(
			common,
			c2s,
			s2s,
			apdb,
			clock)
	} else if isC2S {
		c2s := NewSocialBehavior(ca, o)
		actor = pub.NewSocialActor(
			common,
			c2s,
			apdb,
			clock)
	} else {
		s2s := NewFederatingBehavior(c, sa, db, po, pk, f, u, tc)
		actor = pub.NewFederatingActor(
			common,
			s2s,
			apdb,
			clock)
	}
	return
}
