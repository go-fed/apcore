package apcore

import (
	"time"

	"github.com/go-fed/activity/pub"
)

var _ pub.Clock = &clock{}

type clock struct {
	loc *time.Location
}

// Creates new clock with IANA Time Zone database string
func newClock(location string) (c *clock, err error) {
	c = &clock{}
	c.loc, err = time.LoadLocation(location)
	return
}

func (c *clock) Now() time.Time {
	return time.Now().In(c.loc)
}
