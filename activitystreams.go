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
	"net/url"
	"strconv"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
)

const (
	collectionStartQuery = "start"
	collectionLenQuery   = "len"
)

func collectionPageStartIndex(id *url.URL) int {
	startStr := id.Query().Get(collectionStartQuery)
	if start, err := strconv.Atoi(startStr); err != nil {
		return 0
	} else {
		return start
	}
}

func collectionPageLength(id *url.URL, def int) int {
	lenStr := id.Query().Get(collectionLenQuery)
	if length, err := strconv.Atoi(lenStr); err != nil {
		return def
	} else {
		return length
	}
}

func collectionPageId(base *url.URL, start, length, def int) (u *url.URL, err error) {
	u, err = url.Parse(base.String())
	qv := url.Values{}
	if start > 0 {
		qv.Set(collectionStartQuery, strconv.Itoa(start))
	}
	if length != def {
		qv.Set(collectionLenQuery, strconv.Itoa(length))
	}
	u.RawQuery = qv.Encode()
	return
}

func toOrderedCollectionPage(id *url.URL, ids []string, current, length int) (ocp vocab.ActivityStreamsOrderedCollectionPage) {
	ocp = streams.NewActivityStreamsOrderedCollectionPage()
	// TODO
	// id
	// items
	// total len
	return
}
