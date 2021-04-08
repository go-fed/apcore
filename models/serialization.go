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
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
)

// Marshal takes any ActivityStreams type and serializes it to JSON.
func Marshal(v vocab.Type) (b []byte, err error) {
	var m map[string]interface{}
	m, err = streams.Serialize(v)
	if err != nil {
		return
	}
	b, err = json.Marshal(m)
	if err != nil {
		return
	}
	return
}

// unmarhsal attempts to deserialize JSON bytes into a value.
func unmarshal(maybeByte, v interface{}) error {
	b, ok := maybeByte.([]byte)
	if !ok {
		return errors.New("failed to assert scan to []byte type")
	}
	return json.Unmarshal(b, v)
}

// SingleRow allows *sql.Rows to be treated as *sql.Row
type SingleRow interface {
	app.SingleRow
}

func MustQueryOneRow(r *sql.Rows, fn func(r SingleRow) error) error {
	return enforceOneRow(r, "", fn)
}

// enforceOneRow ensures that there is only one row in the *sql.Rows
//
// Normally, the SQL operations that assume a single row is being returned by
// the database take only the first row and then discard the rest of the rows
// silently. I would rather we know when our expectations are being violated or
// when the database constraints do not match the expected application logic,
// than silently retrieve an arbitrarily row (since the first one grabbed is
// returned arbitrarily, database-and-driver-dependent).
func enforceOneRow(r *sql.Rows, debugname string, fn func(r SingleRow) error) error {
	var n int
	for r.Next() {
		if n > 0 {
			return fmt.Errorf("%s: multiple database rows retrieved when enforcing one row", debugname)
		}
		err := fn(SingleRow(r))
		if err != nil {
			return err
		}
		n++
	}
	if n == 0 {
		return fmt.Errorf("%s: zero database rows retrieved when enforcing one row", debugname)
	}
	return r.Err()
}

func QueryRows(r *sql.Rows, fn func(r SingleRow) error) error {
	return doForRows(r, "", fn)
}

// doForRows iterates over all rows and inspects for any errors.
func doForRows(r *sql.Rows, debugname string, fn func(r SingleRow) error) error {
	for r.Next() {
		err := fn(SingleRow(r))
		if err != nil {
			return err
		}
	}
	return r.Err()
}

func MustChangeOneRow(r sql.Result) error {
	return mustChangeOneRow(r, nil, "")
}

// mustChangeOneRow ensures an Exec SQL statement changes exactly one row, or
// returns an error.
func mustChangeOneRow(r sql.Result, existing error, name string) error {
	if existing != nil {
		return existing
	}
	if n, err := r.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return fmt.Errorf("sql query for %s changed %d rows instead of 1 row", name, n)
	}
	return nil
}

var _ driver.Valuer = Privileges{}
var _ sql.Scanner = &Privileges{}

// Privileges are a user's privileges serializable and deserializable into JSON
// for database storage.
type Privileges struct {
	// Admin indicates whether to treat the user as an administrator by the
	// framework.
	Admin bool
	// InstanceActor indicates whether to treat the user as the instance
	// actor.
	InstanceActor bool
	// Payload is additional privilege information that is app-specific.
	Payload json.RawMessage
}

func (p Privileges) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Privileges) Scan(src interface{}) error {
	return unmarshal(src, p)
}

var _ driver.Valuer = Preferences{}
var _ sql.Scanner = &Preferences{}

// Preferences are a user's preferences serializable and deserializable into
// JSON for database storage.
type Preferences struct {
	// OnFollow indicates default behavior when a Follow request is received
	// by a user.
	OnFollow OnFollowBehavior
	// Payload is additional preference information that is app-specific.
	Payload json.RawMessage
}

func (p Preferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Preferences) Scan(src interface{}) error {
	return unmarshal(src, p)
}

var _ driver.Valuer = InstanceActorPreferences{}
var _ sql.Scanner = &InstanceActorPreferences{}

// InstanceActorPreferences are the preferences for an instance actor which are
// serializable and deserializable into JSON for database storage.
type InstanceActorPreferences struct {
	// OnFollow indicates default behavior when a Follow request is received
	// by a user.
	OnFollow OnFollowBehavior
	// OpenRegistrations indicates whether registrations are open for this
	// software.
	OpenRegistrations bool
	// ServerBaseURL indicates the "base URL" of this server.
	ServerBaseURL string
	// ServerName contains the name of this particular server.
	ServerName string
	// OrgName contains the name of the wider organization this server
	// belongs to.
	OrgName string
	// OrgContact contains the contact information for the Organization this
	// server belongs to.
	OrgContact string
	// OrgAccount contains the account information representing the
	// Organization this server belongs to.
	OrgAccount string
	// Payload is additional preference information that is app-specific.
	Payload json.RawMessage
}

func (p InstanceActorPreferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *InstanceActorPreferences) Scan(src interface{}) error {
	return unmarshal(src, p)
}

var _ driver.Valuer = OnFollowBehavior(0)
var _ sql.Scanner = (*OnFollowBehavior)(nil)

// OnFollowBehavior is a wrapper around pub.OnFollowBehavior type that also
// knows how to serialize and deserialize itself for SQL database drivers in a
// more readable manner.
type OnFollowBehavior pub.OnFollowBehavior

const (
	onFollowAlwaysAccept = "ALWAYS_ACCEPT"
	onFollowAlwaysReject = "ALWAYS_REJECT"
	onFollowManual       = "MANUAL"
)

func (o OnFollowBehavior) Value() (driver.Value, error) {
	switch pub.OnFollowBehavior(o) {
	case pub.OnFollowAutomaticallyAccept:
		return onFollowAlwaysAccept, nil
	case pub.OnFollowAutomaticallyReject:
		return onFollowAlwaysReject, nil
	case pub.OnFollowDoNothing:
		fallthrough
	default:
		return onFollowManual, nil
	}
}

func (o *OnFollowBehavior) Scan(src interface{}) error {
	s, ok := src.(string)
	if !ok {
		return errors.New("failed to assert scan to string type")
	}
	switch s {
	case onFollowAlwaysAccept:
		*o = OnFollowBehavior(pub.OnFollowAutomaticallyAccept)
	case onFollowAlwaysReject:
		*o = OnFollowBehavior(pub.OnFollowAutomaticallyReject)
	case onFollowManual:
		fallthrough
	default:
		*o = OnFollowBehavior(pub.OnFollowDoNothing)
	}
	return nil
}

var _ driver.Valuer = ActivityStreams{nil}
var _ sql.Scanner = &ActivityStreams{nil}

// ActivityStreams is a wrapper around any ActivityStreams type that also
// knows how to serialize and deserialize itself for SQL database drivers.
type ActivityStreams struct {
	vocab.Type
}

func (a ActivityStreams) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreams) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	var err error
	a.Type, err = streams.ToType(context.Background(), m)
	return err
}

var _ driver.Valuer = ActivityStreamsPerson{nil}
var _ sql.Scanner = &ActivityStreamsPerson{nil}

// ActivityStreamsPerson is a wrapper around the ActivityStreams type that also
// knows how to serialize and deserialize itself for SQL database drivers.
type ActivityStreamsPerson struct {
	vocab.ActivityStreamsPerson
}

func (a ActivityStreamsPerson) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreamsPerson) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	res, err := streams.NewJSONResolver(func(ctx context.Context, p vocab.ActivityStreamsPerson) error {
		a.ActivityStreamsPerson = p
		return nil
	})
	if err != nil {
		return err
	}
	return res.Resolve(context.Background(), m)
}

var _ driver.Valuer = ActivityStreamsApplication{nil}
var _ sql.Scanner = &ActivityStreamsApplication{nil}

// ActivityStreamsApplication is a wrapper around the ActivityStreams type that also
// knows how to serialize and deserialize itself for SQL database drivers.
type ActivityStreamsApplication struct {
	vocab.ActivityStreamsApplication
}

func (a ActivityStreamsApplication) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreamsApplication) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	res, err := streams.NewJSONResolver(func(ctx context.Context, p vocab.ActivityStreamsApplication) error {
		a.ActivityStreamsApplication = p
		return nil
	})
	if err != nil {
		return err
	}
	return res.Resolve(context.Background(), m)
}

var _ driver.Valuer = ActivityStreamsOrderedCollection{nil}
var _ sql.Scanner = &ActivityStreamsOrderedCollection{nil}

// ActivityStreamsOrderedCollection is a wrapper around the ActivityStreams type
// that also knows how to serialize and deserialize itself for SQL database
// drivers.
type ActivityStreamsOrderedCollection struct {
	vocab.ActivityStreamsOrderedCollection
}

func (a ActivityStreamsOrderedCollection) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreamsOrderedCollection) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	res, err := streams.NewJSONResolver(func(ctx context.Context, oc vocab.ActivityStreamsOrderedCollection) error {
		a.ActivityStreamsOrderedCollection = oc
		return nil
	})
	if err != nil {
		return err
	}
	return res.Resolve(context.Background(), m)
}

var _ driver.Valuer = ActivityStreamsOrderedCollectionPage{nil}
var _ sql.Scanner = &ActivityStreamsOrderedCollectionPage{nil}

// ActivityStreamsOrderedCollectionPage is a wrapper around the ActivityStreams
// type that also knows how to serialize and deserialize itself for SQL database
// drivers.
type ActivityStreamsOrderedCollectionPage struct {
	vocab.ActivityStreamsOrderedCollectionPage
}

func (a ActivityStreamsOrderedCollectionPage) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreamsOrderedCollectionPage) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	res, err := streams.NewJSONResolver(func(ctx context.Context, oc vocab.ActivityStreamsOrderedCollectionPage) error {
		a.ActivityStreamsOrderedCollectionPage = oc
		return nil
	})
	if err != nil {
		return err
	}
	return res.Resolve(context.Background(), m)
}

var _ driver.Valuer = ActivityStreamsCollection{nil}
var _ sql.Scanner = &ActivityStreamsCollection{nil}

// ActivityStreamsCollection is a wrapper around the ActivityStreams type
// that also knows how to serialize and deserialize itself for SQL database
// drivers.
type ActivityStreamsCollection struct {
	vocab.ActivityStreamsCollection
}

func (a ActivityStreamsCollection) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreamsCollection) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	res, err := streams.NewJSONResolver(func(ctx context.Context, oc vocab.ActivityStreamsCollection) error {
		a.ActivityStreamsCollection = oc
		return nil
	})
	if err != nil {
		return err
	}
	return res.Resolve(context.Background(), m)
}

var _ driver.Valuer = ActivityStreamsCollectionPage{nil}
var _ sql.Scanner = &ActivityStreamsCollectionPage{nil}

// ActivityStreamsCollectionPage is a wrapper around the ActivityStreams
// type that also knows how to serialize and deserialize itself for SQL database
// drivers.
type ActivityStreamsCollectionPage struct {
	vocab.ActivityStreamsCollectionPage
}

func (a ActivityStreamsCollectionPage) Value() (driver.Value, error) {
	return Marshal(a)
}

func (a *ActivityStreamsCollectionPage) Scan(src interface{}) error {
	var m map[string]interface{}
	if err := unmarshal(src, &m); err != nil {
		return err
	}
	res, err := streams.NewJSONResolver(func(ctx context.Context, oc vocab.ActivityStreamsCollectionPage) error {
		a.ActivityStreamsCollectionPage = oc
		return nil
	})
	if err != nil {
		return err
	}
	return res.Resolve(context.Background(), m)
}

var _ driver.Valuer = NullDuration{}
var _ sql.Scanner = &NullDuration{}

// NullDuration can handle nullable time.Duration values in the database.
type NullDuration struct {
	Duration time.Duration
	Valid    bool
}

func (n NullDuration) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Duration, nil
}

func (n *NullDuration) Scan(src interface{}) error {
	if src == nil {
		n.Duration, n.Valid = 0, false
		return nil
	}
	t, ok := src.(int64)
	if !ok {
		return errors.New("failed to assert scan to int64 type")
	}
	n.Duration, n.Valid = time.Duration(t), true
	return nil
}

var _ driver.Valuer = URL{}
var _ sql.Scanner = &URL{}

// URL handles serializing/deserializing *url.URL types into databases.
type URL struct {
	*url.URL
}

func (u URL) Value() (driver.Value, error) {
	return u.URL.String(), nil
}

func (u *URL) Scan(src interface{}) error {
	s, ok := src.(string)
	if !ok {
		return errors.New("failed to assert scan to string type")
	}
	var err error
	u.URL, err = url.Parse(s)
	return err
}
