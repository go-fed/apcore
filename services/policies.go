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

package services

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
)

type Policies struct {
	Clock       pub.Clock
	DB          *sql.DB
	Policies    *models.Policies
	Resolutions *models.Resolutions
}

func (p *Policies) IsBlocked(c util.Context, actorID *url.URL, a pub.Activity) (blocked bool, err error) {
	var iri *url.URL
	iri, err = pub.GetId(a)
	if err != nil {
		return
	}
	var jsonb []byte
	jsonb, err = models.Marshal(a)
	if err != nil {
		return
	}
	err = doInTx(c, p.DB, func(tx *sql.Tx) error {
		pd, err := p.Policies.GetForActorAndPurpose(c, tx, actorID, models.FederatedBlockPurpose)
		if err != nil {
			return err
		}
		for _, policy := range pd {
			var res models.Resolution
			res.Time = p.Clock.Now()
			err = policy.Policy.Resolve(jsonb, &res)
			if err != nil {
				return err
			}
			err = p.Resolutions.Create(c, tx, models.CreateResolution{
				PolicyID: policy.ID,
				IRI:      iri,
				R:        res,
			})
			if err != nil {
				return err
			}
			blocked = blocked || res.Matched
		}
		return nil
	})
	return
}
