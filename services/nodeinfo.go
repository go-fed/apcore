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

package services

import (
	"database/sql"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

type NodeInfoStats struct {
	TotalUsers     int
	ActiveHalfYear int
	ActiveMonth    int
	ActiveWeek     int
	NLocalPosts    int
	NLocalComments int
}

type ServerProfile struct {
	OpenRegistrations bool
	ServerBaseURL     string
	ServerName        string
	OrgName           string
	OrgContact        string
	OrgAccount        string
}

type NodeInfo struct {
	DB               *sql.DB
	Users            *models.Users
	Rand             *rand.Rand
	Mu               *sync.RWMutex
	CacheInvalidated time.Duration
	cache            NodeInfoStats
	cacheSet         bool
	cacheWhen        time.Time
}

func (n *NodeInfo) GetAnonymizedStats(c util.Context) (t NodeInfoStats, err error) {
	// Cache-hit
	n.Mu.RLock()
	if t, ok := n.getCachedAnonymizedStats(); ok {
		n.Mu.RUnlock()
		return t, nil
	}
	n.Mu.RUnlock()
	// Cache-miss...
	n.Mu.Lock()
	defer n.Mu.Unlock()
	// ... but another goroutine may have refreshed ...
	if t, ok := n.getCachedAnonymizedStats(); ok {
		return t, nil
	}
	// ... or we are the one to refresh it.
	var uas models.UserActivityStats
	if err = doInTx(c, n.DB, func(tx *sql.Tx) error {
		uas, err = n.Users.ActivityStats(c, tx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}
	now := time.Now()
	t = NodeInfoStats{
		TotalUsers:     uas.TotalUsers,
		ActiveHalfYear: uas.ActiveHalfYear,
		ActiveMonth:    uas.ActiveMonth,
		ActiveWeek:     uas.ActiveWeek,
		NLocalPosts:    uas.NLocalPosts,
		NLocalComments: uas.NLocalComments,
	}
	n.applyNoise(&t)
	n.setCachedAnonymizedStats(t, now)
	return
}

func (n *NodeInfo) GetServerProfile(c util.Context) (p ServerProfile, err error) {
	var iup models.InstanceUserProfile
	if err = doInTx(c, n.DB, func(tx *sql.Tx) error {
		iup, err = n.Users.InstanceActorProfile(c, tx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}
	p = ServerProfile{
		OpenRegistrations: iup.OpenRegistrations,
		ServerBaseURL:     iup.ServerBaseURL,
		ServerName:        iup.ServerName,
		OrgName:           iup.OrgName,
		OrgContact:        iup.OrgContact,
		OrgAccount:        iup.OrgAccount,
	}
	return
}

func (n *NodeInfo) SetProfile(c util.Context, p ServerProfile) (err error) {
	iup := models.InstanceUserProfile{
		OpenRegistrations: p.OpenRegistrations,
		ServerBaseURL:     p.ServerBaseURL,
		ServerName:        p.ServerName,
		OrgName:           p.OrgName,
		OrgContact:        p.OrgContact,
		OrgAccount:        p.OrgAccount,
	}
	err = doInTx(c, n.DB, func(tx *sql.Tx) error {
		return n.Users.SetInstanceActorProfile(c, tx, iup)
	})
	return
}

// applyNoise ensures that the NodeInfoStats for small instances contains some
// noise around the true value, so that ballpark-correct statistics can be
// obtained from small instances without allowing peers to monitor changes over
// time in number of users or user login activity for the small instances.
//
// The mutex must be locked.
func (n *NodeInfo) applyNoise(t *NodeInfoStats) {
	const (
		uSDev = 2.0
		vSDev = 1.0
	)
	t.TotalUsers = n.maybeGetWithUncertainty(t.TotalUsers, uSDev, vSDev, -1)
	t.ActiveHalfYear = n.maybeGetWithUncertainty(t.ActiveHalfYear, uSDev, vSDev, t.TotalUsers)
	t.ActiveMonth = n.maybeGetWithUncertainty(t.ActiveMonth, uSDev, vSDev, t.TotalUsers)
	t.ActiveWeek = n.maybeGetWithUncertainty(t.ActiveWeek, uSDev, vSDev, t.TotalUsers)
}

// maybeGetWithUncertainty applies noise to counts that do not meet the
// threshold, to ensure privacy.
//
// The mutex must be locked.
func (n *NodeInfo) maybeGetWithUncertainty(v int, s1, s2 float64, max int) int {
	const (
		threshold = 50
	)
	if v >= threshold {
		return v
	}
	return n.getWithUncertainty(v, s1, s2, max)
}

// getWithUncertainty determines a random value using uncertainty in the mean
// and rejection sampling from [0, max]. Max is ignored if <= 0.
//
// The mutex must be locked.
func (n *NodeInfo) getWithUncertainty(v int, s1, s2 float64, max int) int {
	i := -1
	for i < 0 && (max <= 0 || i < max) {
		mu := n.Rand.NormFloat64()*s1 + float64(v)
		val := n.Rand.NormFloat64()*s2 + mu
		i = int(math.Round(val))
	}
	return i
}

// getCachedAnonymizedStats ensures that any stats computed and anonymized with
// noise is not recomputed frequently. Too frequent samples allows guessing the
// true mean, within a uSDev value.
//
// The mutex must be locked.
func (n *NodeInfo) getCachedAnonymizedStats() (t NodeInfoStats, ok bool) {
	now := time.Now()
	ok = n.cacheSet && now.Sub(n.cacheWhen) < n.CacheInvalidated
	if ok {
		t = n.cache
	}
	return
}

// setCachedAnonymizedStats saves anonymized statistics.
//
// The mutex must be locked.
func (n *NodeInfo) setCachedAnonymizedStats(t NodeInfoStats, m time.Time) {
	n.cache = t
	n.cacheSet = true
	n.cacheWhen = m
}
