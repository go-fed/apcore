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

package paths

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	queryTrue           = "true"
	queryCollectionPage = "page"
	queryCollectionEnd  = "end"
	queryOffset         = "offset"
	queryNum            = "n"
)

// AddPageParams overwrites the query string of a base URL and returns a copy
// with the pagination parameters set.
func AddPageParams(base *url.URL, offset, n int) *url.URL {
	c := *base
	c.RawQuery = fmt.Sprintf("%s=%s&%s=%d&%s=%d",
		queryCollectionPage,
		queryTrue,
		queryOffset,
		offset,
		queryNum,
		n)
	return &c
}

// IsGetCollectionPage returns true when the IRI requests pagination for an
// OrderedCollection-style of IRI.
func IsGetCollectionPage(u *url.URL) bool {
	return u.Query().Get(queryCollectionPage) == queryTrue
}

// IsGetCollectionEnd returns true when the IRI requests the last page for an
// OrderedCollection-style of IRI.
func IsGetCollectionEnd(u *url.URL) bool {
	return u.Query().Get(queryCollectionEnd) == queryTrue
}

// GetOffsetOrDefault returns the offset requested in the IRI, or default if
// no value or an invalid value is specified.
func GetOffsetOrDefault(u *url.URL, def int) int {
	return queryKeyAsIntOrDefault(u, queryOffset, def)
}

// GetNumOrDefault returns the number requested in the IRI, or default if no
// value or an invalid value is specified. If the requested amount is greater
// than the max, the maximum is returned instead.
func GetNumOrDefault(u *url.URL, def, max int) int {
	n := queryKeyAsIntOrDefault(u, queryNum, def)
	if n > max {
		return max
	}
	return n
}

func queryKeyAsIntOrDefault(u *url.URL, key string, def int) int {
	v := u.Query().Get(key)
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
