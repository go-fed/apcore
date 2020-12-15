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
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/util"
)

// AnyOCPageFn fetches any arbitrary OrderedCollectionPage
type AnyOCPageFn func(c util.Context, iri *url.URL, min, n int) (vocab.ActivityStreamsOrderedCollectionPage, error)

// LastOCPageFn fetches the last page of an OrderedCollection.
type LastOCPageFn func(c util.Context, iri *url.URL, n int) (vocab.ActivityStreamsOrderedCollectionPage, error)

// DoPagination examines the query parameters of an IRI, and uses it to either
// fetch the bare ordered collection without values, the very last ordered
// collection page, or an arbitrary ordered collection page using the provided
// fetching functions.
func DoOrderedCollectionPagination(c util.Context, iri *url.URL, defaultSize, maxSize int, any AnyOCPageFn, last LastOCPageFn) (p vocab.ActivityStreamsOrderedCollectionPage, err error) {
	if paths.IsGetCollectionPage(iri) && paths.IsGetCollectionEnd(iri) {
		// The last page was requested
		n := paths.GetNumOrDefault(iri, defaultSize, maxSize)
		p, err = last(c, paths.Normalize(iri), n)
		return
	} else {
		// The first page, or an arbitrary page, was requested
		offset, n := 0, defaultSize
		if paths.IsGetCollectionPage(iri) {
			// An arbitrary page was requested
			offset = paths.GetOffsetOrDefault(iri, 0)
			n = paths.GetNumOrDefault(iri, defaultSize, maxSize)
		}
		p, err = any(c, paths.Normalize(iri), offset, n)
		return
	}
}

// AnyCPageFn fetches any arbitrary CollectionPage
type AnyCPageFn func(c util.Context, iri *url.URL, min, n int) (vocab.ActivityStreamsCollectionPage, error)

// LastCPageFn fetches the last page of an Collection.
type LastCPageFn func(c util.Context, iri *url.URL, n int) (vocab.ActivityStreamsCollectionPage, error)

// DoCollectionPagination examines the query parameters of an IRI, and uses it
// to either fetch the bare ordered collection without values, the very last
// ordered collection page, or an arbitrary ordered collection page using the
// provided fetching functions.
func DoCollectionPagination(c util.Context, iri *url.URL, defaultSize, maxSize int, any AnyCPageFn, last LastCPageFn) (p vocab.ActivityStreamsCollectionPage, err error) {
	if paths.IsGetCollectionPage(iri) && paths.IsGetCollectionEnd(iri) {
		// The last page was requested
		n := paths.GetNumOrDefault(iri, defaultSize, maxSize)
		p, err = last(c, paths.Normalize(iri), n)
		return
	} else {
		// The first page, or an arbitrary page, was requested
		offset, n := 0, defaultSize
		if paths.IsGetCollectionPage(iri) {
			// An arbitrary page was requested
			offset = paths.GetOffsetOrDefault(iri, 0)
			n = paths.GetNumOrDefault(iri, defaultSize, maxSize)
		}
		p, err = any(c, paths.Normalize(iri), offset, n)
		return
	}
}
