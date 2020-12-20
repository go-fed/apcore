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

package util

import (
	"context"
	"sync"
	"time"
)

// SafeStartStop guarantees at most one asynchronous function is being
// periodically run, no matter how many asynchronous calls to Start or Stop
// are being invoked.
//
// There is no order to how the
type SafeStartStop struct {
	// Immutable
	goFunc func(context.Context)
	period time.Duration
	wg     sync.WaitGroup // To coordinate when goFunc is done stopping
	mu     sync.Mutex     // Must be locked to modify any of the below
	// Mutable
	fnTimer  *time.Timer        // For periodic invocation of goFunc
	fnCtx    context.Context    // For managing stopping
	fnCancel context.CancelFunc // For beginning the stopping process
}

func NewSafeStartStop(fn func(context.Context), period time.Duration) *SafeStartStop {
	return &SafeStartStop{
		goFunc: fn,
		period: period,
	}
}

func (s *SafeStartStop) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fnCtx != nil {
		return
	}
	s.fnCtx, s.fnCancel = context.WithCancel(context.Background())
	s.fnTimer = time.NewTimer(s.period)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.fnTimer.C:
				s.goFunc(s.fnCtx)
				// Timers are tricky to get correct, especially
				// when calling Reset. From the documentation:
				//
				// Reset should be invoked only on stopped or
				// expired timers with drained channels. If a
				// program has already received a value from
				// t.C, the timer is known to have expired and
				// the channel drained, so t.Reset can be used
				// directly.
				s.fnTimer.Reset(s.period)
			case <-s.fnCtx.Done():
				return
			}
		}
	}()
}

func (s *SafeStartStop) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fnCancel == nil {
		return
	}
	s.fnCancel()
	s.wg.Wait()
	if !s.fnTimer.Stop() {
		<-s.fnTimer.C
	}
	s.fnTimer = nil
	s.fnCtx = nil
	s.fnCancel = nil
}
