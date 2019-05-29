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
	apdb := newApdb(db)

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
