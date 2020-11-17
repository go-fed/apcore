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

package conn

import (
	"context"
	"sync"
	"time"

	"github.com/go-fed/apcore/framework/config"
	"golang.org/x/time/rate"
)

type entry struct {
	L        *rate.Limiter
	LastUsed time.Time
}

type hostLimiter struct {
	// Immutable
	limit       rate.Limit
	burst       int
	prunePeriod time.Duration
	pruneAge    time.Duration
	wg          sync.WaitGroup
	// Mutable
	pruneTicker *time.Ticker
	pruneCtx    context.Context
	pruneCancel context.CancelFunc
	pMu         sync.Mutex
	m           map[string]entry
	mu          sync.Mutex
}

func newHostLimiter(c *config.Config) *hostLimiter {
	return &hostLimiter{
		limit:       rate.Limit(c.ActivityPubConfig.OutboundRateLimitQPS),
		burst:       c.ActivityPubConfig.OutboundRateLimitBurst,
		prunePeriod: time.Duration(c.ActivityPubConfig.OutboundRateLimitPrunePeriodSeconds) * time.Second,
		pruneAge:    time.Duration(c.ActivityPubConfig.OutboundRateLimitPruneAgeSeconds) * time.Second,
		m:           make(map[string]entry),
	}
}

func (h *hostLimiter) Start() {
	h.resetMap()
	h.goPrune()
}

func (h *hostLimiter) Stop() {
	h.stopPrune()
}

func (h *hostLimiter) Get(host string) *rate.Limiter {
	h.mu.Lock()
	defer h.mu.Unlock()
	e, ok := h.m[host]
	if ok {
		e.LastUsed = time.Now()
		h.m[host] = e
		return e.L
	} else {
		e = entry{
			L:        rate.NewLimiter(h.limit, h.burst),
			LastUsed: time.Now(),
		}
		h.m[host] = e
		return e.L
	}
}

func (h *hostLimiter) resetMap() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.m = make(map[string]entry)
}

func (h *hostLimiter) stopPrune() {
	h.pMu.Lock()
	defer h.pMu.Unlock()
	if h.pruneCancel == nil {
		return
	}
	h.pruneCancel()
	h.wg.Wait()
}

func (h *hostLimiter) goPrune() {
	h.pMu.Lock()
	defer h.pMu.Unlock()
	if h.pruneTicker != nil {
		return
	}
	h.pruneTicker = time.NewTicker(h.prunePeriod)
	h.pruneCtx, h.pruneCancel = context.WithCancel(context.Background())
	h.wg.Add(1)
	go func() {
		defer func() {
			h.pMu.Lock()
			defer h.pMu.Unlock()
			h.pruneTicker.Stop()
			h.pruneTicker = nil
			h.pruneCtx = nil
			h.pruneCancel = nil
			h.wg.Done()
		}()
		for {
			select {
			case <-h.pruneTicker.C:
				h.prune()
			case <-h.pruneCtx.Done():
				return
			}
		}
	}()
}

func (h *hostLimiter) prune() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	for k, v := range h.m {
		if v.LastUsed.Sub(now) > h.pruneAge {
			delete(h.m, k)
		}
	}
}
