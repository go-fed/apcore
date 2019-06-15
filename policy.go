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
	"context"
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

type scanner interface {
	Scan(...interface{}) error
}

// resolution is the decision of one or more policies.
type resolution struct {
	Id           string
	Order        int
	Permit       permit
	ActivityId   *url.URL
	TargetUserId string
	Public       bool
	PolicyId     string
	Reason       string
}

func (r *resolution) Load(row scanner) (err error) {
	var activityIRI string
	if err = row.Scan(
		&r.Id,
		&r.TargetUserId,
		&r.Permit,
		&activityIRI,
		&r.Order,
		&r.Public,
		&r.Reason,
		&r.PolicyId); err != nil {
		return
	}
	if r.ActivityId, err = url.Parse(activityIRI); err != nil {
		return
	}
	return
}

const (
	alwaysGrant   = "always_grant"
	alwaysDeny    = "always_deny"
	instanceGrant = "instance_grant"
	instanceDeny  = "instance_deny"
	actorGrant    = "actor_grant"
	actorDeny     = "actor_deny"
)

// policy determines what kind of resolution is appropriate.
//
// Used to determine interaction blocks.
type policy struct {
	Id               string
	Order            int
	IsInstancePolicy bool
	UserId           string
	Description      string
	Public           bool
	Subject          string
	Kind             string
	Resolve          func(from []*url.URL, activityType string) (p permit, reason string)
}

func (p *policy) Load(r scanner, isInstance bool) (err error) {
	p.IsInstancePolicy = isInstance
	if p.IsInstancePolicy {
		p.Public = true
		if err = r.Scan(
			&p.Id,
			&p.Order,
			&p.Description,
			&p.Subject,
			&p.Kind); err != nil {
			return
		}
	} else {
		p.Public = false
		if err = r.Scan(
			&p.Id,
			&p.UserId,
			&p.Description,
			&p.Subject,
			&p.Kind); err != nil {
			return
		}
	}
	switch p.Kind {
	case alwaysGrant:
		p.Resolve = func(from []*url.URL, activityType string) (perm permit, reason string) {
			perm = grant
			reason = "always permit"
			return
		}
	case alwaysDeny:
		p.Resolve = func(from []*url.URL, activityType string) (perm permit, reason string) {
			perm = deny
			reason = "always deny"
			return
		}
	case instanceGrant:
		p.Resolve = func(from []*url.URL, activityType string) (perm permit, reason string) {
			perm = unknown
			reason = fmt.Sprintf("could not match host %q for instance grant", p.Subject)
			for _, f := range from {
				if f.Host == p.Subject {
					perm = grant
					reason = fmt.Sprintf("%q matched host %q for instance grant", f, p.Subject)
					return
				}
			}
			return
		}
	case instanceDeny:
		p.Resolve = func(from []*url.URL, activityType string) (perm permit, reason string) {
			perm = unknown
			reason = fmt.Sprintf("could not match host %q for instance deny", p.Subject)
			for _, f := range from {
				if f.Host == p.Subject {
					perm = deny
					reason = fmt.Sprintf("%q matched host %q for instance deny", f, p.Subject)
					return
				}
			}
			return
		}
	case actorGrant:
		p.Resolve = func(from []*url.URL, activityType string) (perm permit, reason string) {
			perm = unknown
			reason = fmt.Sprintf("could not match actor %q for actor grant", p.Subject)
			for _, f := range from {
				if f.String() == p.Subject {
					perm = grant
					reason = fmt.Sprintf("%q matched actor for grant", f)
					return
				}
			}
			return
		}
	case actorDeny:
		p.Resolve = func(from []*url.URL, activityType string) (perm permit, reason string) {
			perm = unknown
			reason = fmt.Sprintf("could not match actor %q for actor deny", p.Subject)
			for _, f := range from {
				if f.String() == p.Subject {
					perm = deny
					reason = fmt.Sprintf("%q matched actor for deny", f)
					return
				}
			}
			return
		}
	default:
		err = fmt.Errorf("unknown kind of policy: %s", p.Kind)
	}
	return
}

type policies []policy

// IsBlocked uses a number of policies to determine and record resolutions.
func (p policies) IsBlocked(c context.Context, db *database, targetUserId string, from []*url.URL, activityIRI *url.URL, activityType string) (blocked bool, err error) {
	var r []resolution
	defer func() {
		if err == nil {
			err = db.InsertResolutions(c, r)
		}
	}()
	if len(p) == 0 {
		err = fmt.Errorf("no policies to evaluate")
		return
	}
	outcome := unknown
	for i, policy := range p {
		res := resolution{
			ActivityId:   activityIRI,
			TargetUserId: targetUserId,
			Public:       policy.Public,
			PolicyId:     policy.Id,
			Order:        i,
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
