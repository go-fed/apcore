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
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

var _ pub.Database = &apdb{}

type apdb struct {
	*database
	// Use sync.Map, which is specially optimized:
	//
	// "The Map type is optimized [...] when the entry for a given key is
	// only ever written once but read many times, as in caches that only
	// grow"
	//
	// This means we only ever append to the map during the lifetime of the
	// running application. This may become a scaling bottleneck in the
	// future, but unsure how the performance will look in practice.
	//
	// This map will only store *sync.Mutex, each is 4 bytes. Assuming that
	// conservatively the average key is a string of 124 bytes, this means
	// each entry is 128 bytes of memory.
	//
	// If this map holds 2,000,000 entries then it would take 256 MB of
	// memory. To take up 1 GB, 7,812,500 entries are needed. If one entry
	// is added per second, then in 90 days it will take up 1 GB of memory.
	//
	// TODO: Address this unbounded growth for memory-constrained or very
	// long running applications.
	locks *sync.Map
	app   Application
}

func newApdb(db *database, a Application) *apdb {
	return &apdb{
		database: db,
		locks:    &sync.Map{},
		app:      a,
	}
}

func (a *apdb) Lock(c context.Context, id *url.URL) error {
	mui, _ := a.locks.LoadOrStore(id.String(), &sync.Mutex{})
	if mu, ok := mui.(*sync.Mutex); !ok {
		return fmt.Errorf("lock for Lock is not a *sync.Mutex")
	} else {
		mu.Lock()
		return nil
	}
}

func (a *apdb) Unlock(c context.Context, id *url.URL) error {
	mui, _ := a.locks.Load(id.String())
	if mu, ok := mui.(*sync.Mutex); !ok {
		return fmt.Errorf("lock for Unlock is not a *sync.Mutex")
	} else {
		mu.Unlock()
		return nil
	}
}

func (a *apdb) NewId(c context.Context, t vocab.Type) (id *url.URL, err error) {
	return a.app.NewId(c, t)
}
