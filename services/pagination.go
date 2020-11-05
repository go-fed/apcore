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
