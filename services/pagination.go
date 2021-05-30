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

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/util"
)

func getOffsetN(iri *url.URL, defaultSize, maxSize int) (offset, n int) {
	offset, n = 0, defaultSize
	if paths.IsGetCollectionPage(iri) {
		offset = paths.GetOffsetOrDefault(iri, 0)
		n = paths.GetNumOrDefault(iri, defaultSize, maxSize)
	}
	return
}

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
		offset, n := getOffsetN(iri, defaultSize, maxSize)
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
		offset, n := getOffsetN(iri, defaultSize, maxSize)
		p, err = any(c, paths.Normalize(iri), offset, n)
		return
	}
}

// PrependFn are functions that prepend items to a collection.
type PrependFn func(c util.Context, collectionID, item *url.URL) error

// UpdateCollectionToPrependCalls takes new beginning elements of a collection
// in order to generate calls to PrependFn in order.
//
// This function only prepends to the very beginning of the collection, and
// expects the page to be the first one, though it is written as if for the
// general case.
//
// TODO: Could generalize this to apply a diff to a portion of the collection
func UpdateCollectionToPrependCalls(c util.Context, updated vocab.ActivityStreamsCollection, defaultSize, maxSize int, firstPageFn AnyCPageFn, prependFn PrependFn) error {
	iri, err := pub.GetId(updated)
	if err != nil {
		return err
	}
	// Get the updated items -- early out if none.
	newItems := updated.GetActivityStreamsItems()
	if newItems == nil || newItems.Len() == 0 {
		return nil
	}
	// Obtain the same number as the pre-updated ID
	offset, n := getOffsetN(iri, defaultSize, maxSize)
	original, err := firstPageFn(c, paths.Normalize(iri), offset, n)
	if err != nil {
		return err
	}
	// Call Prepend for items that come before the first element.
	var firstIRI *url.URL
	if items := original.GetActivityStreamsItems(); items != nil && items.Len() > 0 {
		firstIRI, err = pub.ToId(items.At(0))
		if err != nil {
			return err
		}
	}
	found := firstIRI == nil // If firstIRI is nil, add everything
	for i := newItems.Len() - 1; i >= 0; i-- {
		elemID, err := pub.ToId(newItems.At(i))
		if err != nil {
			return err
		}
		if found {
			// We already found the matching formerly-first
			// element, so prepend the rest.
			if err = prependFn(c, iri, elemID); err != nil {
				return err
			}
		} else if elemID.String() == firstIRI.String() {
			found = true
		}
	}
	return nil
}
