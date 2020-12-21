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

package app

import (
	"net/http"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/util"
)

type AuthorizeFunc func(c util.Context, w http.ResponseWriter, r *http.Request, db Database) (permit bool, err error)

type CollectionPageHandlerFunc func(http.ResponseWriter, *http.Request, vocab.ActivityStreamsCollectionPage)

type VocabHandlerFunc func(http.ResponseWriter, *http.Request, vocab.Type)
