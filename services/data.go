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
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/util"
)

type Data struct {
	DB                    *sql.DB
	Hostname              string
	FedData               *models.FedData
	LocalData             *models.LocalData
	Users                 *models.Users
	Following             *Following
	Followers             *Followers
	Liked                 *Liked
	DefaultCollectionSize int
	MaxCollectionPageSize int
}

// Owns determines if this IRI is a local or federated piece of data.
func (d *Data) Owns(id *url.URL) bool {
	return id.Host == d.Hostname
}

// Exists determines if this ActivityStreams ID already exists locally or
// federated.
func (d *Data) Exists(c util.Context, id *url.URL) (exists bool, err error) {
	if d.Owns(id) {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			exists, err = d.LocalData.Exists(c, tx, id)
			return err
		})
	} else {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			exists, err = d.FedData.Exists(c, tx, id)
			return err
		})
	}
	return
}

// Get obtains the federated or local ActivityStreams data.
func (d *Data) Get(c util.Context, id *url.URL) (v vocab.Type, err error) {
	if d.Owns(id) {
		// Determine whether this is a user, any of a user's sub-path data, or local data
		if paths.IsFollowersPath(id) {
			any := d.Followers.GetPage
			last := d.Followers.GetLastPage
			v, err = DoCollectionPagination(c,
				id,
				d.DefaultCollectionSize,
				d.MaxCollectionPageSize,
				any,
				last)
		} else if paths.IsFollowingPath(id) {
			any := d.Following.GetPage
			last := d.Following.GetLastPage
			v, err = DoCollectionPagination(c,
				id,
				d.DefaultCollectionSize,
				d.MaxCollectionPageSize,
				any,
				last)
		} else if paths.IsLikedPath(id) {
			any := d.Liked.GetPage
			last := d.Liked.GetLastPage
			v, err = DoCollectionPagination(c,
				id,
				d.DefaultCollectionSize,
				d.MaxCollectionPageSize,
				any,
				last)
		} else if paths.IsInstanceActorPath(id) {
			err = doInTx(c, d.DB, func(tx *sql.Tx) error {
				var as *models.User
				as, err = d.Users.InstanceActorUser(c, tx)
				if err != nil {
					return err
				}
				v = as.Actor.Type
				return nil
			})
		} else if paths.IsUserPath(id) {
			var uid paths.UUID
			uid, err = paths.UUIDFromUserPath(id.Path)
			if err != nil {
				return
			}
			err = doInTx(c, d.DB, func(tx *sql.Tx) error {
				var as *models.User
				as, err = d.Users.UserByID(c, tx, string(uid))
				if err != nil {
					return err
				}
				v = as.Actor.Type
				return nil
			})
		} else {
			err = doInTx(c, d.DB, func(tx *sql.Tx) error {
				var as models.ActivityStreams
				as, err = d.LocalData.Get(c, tx, id)
				if err != nil {
					return err
				}
				v = as.Type
				return nil
			})
		}
	} else {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			var as models.ActivityStreams
			as, err = d.FedData.Get(c, tx, id)
			if err != nil {
				return err
			}
			v = as.Type
			return nil
		})
	}
	return
}

// Create stores the ActivityStreams payload locally or federated.
func (d *Data) Create(c util.Context, v vocab.Type) (err error) {
	var iri *url.URL
	iri, err = pub.GetId(v)
	if err != nil {
		return
	}
	// Prevent multiple delivery of the same data from creating
	// multiple copies of the same data.
	exists, err := d.Exists(c, iri)
	if err != nil || exists {
		return
	}
	if d.Owns(iri) {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			return d.LocalData.Create(c, tx, models.ActivityStreams{v})
		})
	} else {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			return d.FedData.Create(c, tx, models.ActivityStreams{v})
		})
	}
	return
}

// Update updates the ActivityStreams payload locally or federated.
func (d *Data) Update(c util.Context, v vocab.Type) (err error) {
	var iri *url.URL
	iri, err = pub.GetId(v)
	if err != nil {
		return
	}
	if d.Owns(iri) {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			return d.LocalData.Update(c, tx, iri, models.ActivityStreams{v})
		})
	} else {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			return d.FedData.Update(c, tx, iri, models.ActivityStreams{v})
		})
	}
	return
}

// Delete removes the ActivityStreams payload locally or federated.
func (d *Data) Delete(c util.Context, iri *url.URL) (err error) {
	if d.Owns(iri) {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			return d.LocalData.Delete(c, tx, iri)
		})
	} else {
		err = doInTx(c, d.DB, func(tx *sql.Tx) error {
			return d.FedData.Delete(c, tx, iri)
		})
	}
	return
}
