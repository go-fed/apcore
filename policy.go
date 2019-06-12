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
	"database/sql"
	"fmt"
	"net/url"
)

type permit int

const (
	deny    permit = 0
	grant   permit = 1
	unknown permit = 2
)

func (p permit) and(o permit) permit {
	if p == deny || o == deny {
		return deny
	} else if p == grant || o == grant {
		return grant
	}
	return unknown
}

// resolution is the decision of one or more policies.
type resolution struct {
	Permit       permit
	ActivityId   *url.URL
	TargetUserId string
	Public       bool
	PolicyId     string
	Reason       string
}

func (r *resolution) Load(row *sql.Row) (err error) {
	// TODO
	return
}

// policy determines what kind of resolution is appropriate.
//
// Used to determine interaction blocks.
type policy struct {
	Id               string
	IsInstancePolicy bool
	UserId           string
	Description      string
	Public           bool
	Resolve          func(from []*url.URL, activityType string) (p permit, reason string)
}

func (p *policy) Load(r *sql.Row) (err error) {
	// TODO
	return
}

func instancePolicies(db *database) (p policies, err error) {
	// TODO
	return
}

func userPolicies(db *database, targetUserId string) (p policies, err error) {
	// TODO
	return
}

type policies []policy

// Apply uses a number of policies to determine and record a resolution.
func (p policies) Apply(db *database, targetUserId string, from []*url.URL, activityId *url.URL, activityType string) (blocked bool, err error) {
	var r []resolution
	// TODO: defer saving the resolutions
	if len(p) == 0 {
		err = fmt.Errorf("no policies to evaluate")
		return
	}
	outcome := unknown
	for _, policy := range p {
		res := resolution{
			ActivityId:   activityId,
			TargetUserId: targetUserId,
			Public:       policy.Public,
			PolicyId:     policy.Id,
		}
		res.Permit, res.Reason = policy.Resolve(from, activityType)
		r = append(r, res)
		outcome = outcome.and(res.Permit)
		blocked = outcome == deny
		if blocked {
			return
		}
	}
	if outcome == unknown {
		err = fmt.Errorf("unknown resolution after evaluating all policies")
	}
	return
}
