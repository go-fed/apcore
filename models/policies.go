// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
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

package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-fed/apcore/util"
	"github.com/tidwall/gjson"
)

const (
	FederatedBlockPurpose Purpose = "federated_block"
)

type Purpose string

var _ driver.Valuer = Policy{}
var _ sql.Scanner = &Policy{}

type Policy struct {
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	Matchers    []*KVMatcher `json:"matchers,omitempty"`
}

func (p Policy) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Policy) Scan(src interface{}) error {
	return unmarshal(src, p)
}

func (p Policy) Validate() error {
	if len(p.Name) == 0 {
		return errors.New("missing name")
	}
	for _, m := range p.Matchers {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p Policy) Resolve(json []byte, r *Resolution) error {
	r.Logf("applying policy %q", p.Name)
	var err error
	for idx, m := range p.Matchers {
		r.Logf("resolving matcher %d", idx)
		if err2 := m.Resolve(json, r); err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = fmt.Errorf("%w\n%s", err, err2.Error())
			}
		}
	}
	return err
}

type KVMatcher struct {
	// KeyPathQuery is a GJSON path query
	KeyPathQuery string        `json:"keyPathQuery,omitempty"`
	ValueMatcher *UnaryMatcher `json:"valueMatcher,omitempty"`
}

func (k KVMatcher) Validate() error {
	if len(k.KeyPathQuery) == 0 {
		return errors.New("missing keyPathQuery")
	} else if k.ValueMatcher == nil {
		return errors.New("missing valueMatcher")
	}
	return k.ValueMatcher.Validate()
}

func (k KVMatcher) Resolve(json []byte, r *Resolution) (err error) {
	if r.Matched {
		r.Logf("resolution already found match, skipping examining %q", k.KeyPathQuery)
		return
	}
	r.Logf("examining value of %q", k.KeyPathQuery)
	result := gjson.GetBytes(json, k.KeyPathQuery)
	r.Matched, err = k.ValueMatcher.Match(result, json, r)
	return
}

type UnaryMatcher struct {
	Not   *UnaryMatcher  `json:"not,omitempty"`
	And   *BinaryMatcher `json:"and,omitempty"`
	Or    *BinaryMatcher `json:"or,omitempty"`
	Value *Value         `json:"value,omitempty"`
	Empty bool           `json:"empty,omitempty"`
}

func (u UnaryMatcher) Validate() error {
	n := 0
	if u.Not != nil {
		n++
	}
	if u.And != nil {
		n++
	}
	if u.Or != nil {
		n++
	}
	if u.Value != nil {
		n++
	}
	if u.Empty {
		n++
	}
	if n > 1 {
		return errors.New("unary matcher has >1 field set")
	} else if n == 0 {
		return errors.New("unary matcher has no fields set")
	}
	return nil
}

func (u UnaryMatcher) Match(res gjson.Result, json []byte, r *Resolution) (bool, error) {
	if u.Not != nil {
		in, err := u.Not.Match(res, json, r)
		if err != nil {
			return false, err
		}
		v := !in
		r.Logf("apply NOT(%v)=>%v", in, v)
		return v, nil
	} else if u.And != nil {
		lhs, rhs, err := u.And.Match(res, json, r)
		if err != nil {
			return false, err
		}
		v := lhs && rhs
		r.Logf("apply AND(%v, %v)=>%v", lhs, rhs, v)
		return v, nil
	} else if u.Or != nil {
		lhs, rhs, err := u.Or.Match(res, json, r)
		if err != nil {
			return false, err
		}
		v := lhs || rhs
		r.Logf("apply OR(%v, %v)=>%v", lhs, rhs, v)
		return v, nil
	} else if u.Value != nil {
		v, err := u.Value.Match(res, json, r)
		return v, err
	} else if u.Empty {
		v := !res.Exists()
		r.Logf("apply EMPTY=>%v", v)
		return v, nil
	}
	r.Log("error: Match called with invalid UnaryMatcher")
	return false, errors.New("Match called with invalid UnaryMatcher")
}

type BinaryMatcher struct {
	L *UnaryMatcher `json:"left"`
	R *UnaryMatcher `json:"right"`
}

func (b BinaryMatcher) Validate() error {
	if b.L == nil {
		return errors.New("missing left")
	} else if b.R == nil {
		return errors.New("missing right")
	} else if err := b.L.Validate(); err != nil {
		return err
	}
	return b.R.Validate()
}

func (b BinaryMatcher) Match(res gjson.Result, json []byte, r *Resolution) (lhs, rhs bool, err error) {
	lhs, err = b.L.Match(res, json, r)
	if err != nil {
		return
	}
	rhs, err = b.R.Match(res, json, r)
	if err != nil {
		return
	}
	return
}

type Value struct {
	JSONPath       string `json:"jsonPath,omitempty"`
	EqualsString   string `json:"equalsString,omitempty"`
	ContainsString string `json:"containsString,omitempty"`
	LenEquals      *int   `json:"lenEquals,omitempty"`
	LenGreater     *int   `json:"lenGreater,omitempty"`
	LenLess        *int   `json:"lenLess,omitempty"`
}

func (u Value) Validate() error {
	n := 0
	if len(u.JSONPath) > 0 {
		n++
	}
	if len(u.EqualsString) > 0 {
		n++
	}
	if len(u.ContainsString) > 0 {
		n++
	}
	if u.LenEquals != nil {
		n++
	}
	if u.LenGreater != nil {
		n++
	}
	if u.LenLess != nil {
		n++
	}
	if n > 1 {
		return errors.New("value has >1 field set")
	} else if n == 0 {
		return errors.New("value has no fields set")
	}
	return nil
}

func (u Value) Match(res gjson.Result, json []byte, r *Resolution) (bool, error) {
	if len(u.JSONPath) > 0 {
		other := gjson.GetBytes(json, u.JSONPath)
		v := resultsEqual(res, other)
		r.Logf("apply EQUALS(JSONPath(%s))=>%v", u.JSONPath, v)
		return v, nil
	} else if len(u.EqualsString) > 0 {
		v := res.String() == u.EqualsString
		r.Logf("apply EQUALS(%s)=>%v", u.EqualsString, v)
		return v, nil
	} else if len(u.ContainsString) > 0 {
		v := strings.Contains(res.String(), u.ContainsString)
		r.Logf("apply CONTAINS(%s)=>%v", u.ContainsString, v)
		return v, nil
	} else if u.LenEquals != nil {
		l := resultsLen(res)
		v := l == *u.LenEquals
		r.Logf("apply EQUALS(LEN(), %d)=>%v", *u.LenEquals, v)
		return v, nil
	} else if u.LenGreater != nil {
		l := resultsLen(res)
		v := l > *u.LenGreater
		r.Logf("apply GREATER(LEN(), %d)=>%v", *u.LenGreater, v)
		return v, nil
	} else if u.LenLess != nil {
		l := resultsLen(res)
		v := l < *u.LenLess
		r.Logf("apply LESS(LEN(), %d)=>%v", *u.LenLess, v)
		return v, nil
	}
	r.Log("error: Match called with invalid Value")
	return false, errors.New("Match called with invalid Value")
}

func resultsEqual(lhs, rhs gjson.Result) bool {
	return reflect.DeepEqual(lhs.Value(), rhs.Value())
}

func resultsLen(r gjson.Result) int {
	l := 0
	if r.Exists() {
		l = 1
		if r.IsArray() {
			l = len(r.Array())
		}
	}
	return l
}

type CreatePolicy struct {
	ActorID *url.URL
	Purpose Purpose
	Policy  Policy
}

type PolicyAndPurpose struct {
	ID      string
	Purpose Purpose
	Policy  Policy
}

type PolicyAndID struct {
	ID     string
	Policy Policy
}

var _ Model = &Policies{}

// Policies is a Model that provides additional database methods for the
// Policy type.
type Policies struct {
	create                *sql.Stmt
	getForActor           *sql.Stmt
	getForActorAndPurpose *sql.Stmt
}

func (p *Policies) Prepare(db *sql.DB, s SqlDialect) error {
	return prepareStmtPairs(db,
		stmtPairs{
			{&(p.create), s.CreatePolicy()},
			{&(p.getForActor), s.GetPoliciesForActor()},
			{&(p.getForActorAndPurpose), s.GetPoliciesForActorAndPurpose()},
		})
}

func (p *Policies) CreateTable(t *sql.Tx, s SqlDialect) error {
	_, err := t.Exec(s.CreatePoliciesTable())
	return err
}

func (p *Policies) Close() {
	p.create.Close()
	p.getForActor.Close()
	p.getForActorAndPurpose.Close()
}

// Create a new Policy
func (p *Policies) Create(c util.Context, tx *sql.Tx, cp CreatePolicy) (policyID string, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(p.create).QueryContext(c,
		cp.ActorID.String(),
		cp.Purpose,
		cp.Policy)
	if err != nil {
		return
	}
	defer rows.Close()
	return policyID, enforceOneRow(rows, "Policies.Create", func(r SingleRow) error {
		return r.Scan(&(policyID))
	})
}

// GetForActor obtains all policies for an Actor.
func (p *Policies) GetForActor(c util.Context, tx *sql.Tx, actorID *url.URL) (po []PolicyAndPurpose, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(p.getForActor).QueryContext(c, actorID.String())
	if err != nil {
		return
	}
	defer rows.Close()
	return po, doForRows(rows, "Policies.GetForActor", func(r SingleRow) error {
		var pp PolicyAndPurpose
		if err := r.Scan(&(pp.ID), &(pp.Purpose), &(pp.Policy)); err != nil {
			return err
		}
		po = append(po, pp)
		return nil
	})
}

// GetForActorAndPurpose obtains all policies for an Actor and Purpose.
func (p *Policies) GetForActorAndPurpose(c util.Context, tx *sql.Tx, actorID *url.URL, u Purpose) (po []PolicyAndID, err error) {
	var rows *sql.Rows
	rows, err = tx.Stmt(p.getForActorAndPurpose).QueryContext(c, actorID.String(), u)
	if err != nil {
		return
	}
	defer rows.Close()
	return po, doForRows(rows, "Policies.GetForActorAndPurpose", func(r SingleRow) error {
		var pp PolicyAndID
		if err := r.Scan(&(pp.ID), &(pp.Policy)); err != nil {
			return err
		}
		po = append(po, pp)
		return nil
	})
}
