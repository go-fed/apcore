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
	"time"

	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
)

type retrier struct {
	// Immutable
	da               *services.DeliveryAttempts
	pk               *services.PrivateKeys
	tc               *Controller
	pageSize         int
	abandonLimit     int
	reattemptBackoff func(n int) time.Duration
	retrierFn        *util.SafeStartStop
}

func newRetrier(da *services.DeliveryAttempts, pk *services.PrivateKeys, tc *Controller, c *config.Config) *retrier {
	r := &retrier{
		da:           da,
		pk:           pk,
		tc:           tc,
		pageSize:     c.ActivityPubConfig.RetryPageSize,
		abandonLimit: c.ActivityPubConfig.RetryAbandonLimit,
		reattemptBackoff: func(n int) time.Duration {
			z := time.Duration(c.ActivityPubConfig.RetrySleepPeriod) * time.Second
			// Exponential backoff
			for i := 0; i < n; i++ {
				z += z
			}
			// If larger than a day, cap at one attempt per day
			if z > time.Hour*24 {
				z = time.Hour * 24
			}
			return z
		},
	}
	r.retrierFn = util.NewSafeStartStop(r.retry, time.Duration(c.ActivityPubConfig.RetrySleepPeriod)*time.Second)
	return r
}

func (r *retrier) Start() {
	r.retrierFn.Start()
}

func (r *retrier) Stop() {
	r.retrierFn.Stop()
}

func (r *retrier) retry(ctx context.Context) {
	c := util.Context{ctx}
	now := time.Now()
	failures, err := r.da.FirstPageRetryableFailures(c, r.pageSize)
	if err != nil {
		util.ErrorLogger.Errorf("retrier failed to obtain first page: %s", err)
		return
	}
	for len(failures) > 0 {
		for _, failure := range failures {
			// Skip this if the retry attempt would be too soon;
			// this applies a backoff function.
			if failure.LastAttempt.Sub(now) < r.reattemptBackoff(failure.NAttempts) {
				continue
			}
			privKey, pubKeyID, err := r.pk.GetUserHTTPSignatureKey(c, paths.UUID(failure.UserID))
			if err != nil {
				util.ErrorLogger.Errorf("retrier failed to obtain user's HTTP Signature key: %s", err)
				continue
			}
			tp, err := r.tc.Get(privKey, pubKeyID.String())
			if err != nil {
				util.ErrorLogger.Errorf("retrier failed to obtain a transport for delivery: %s", err)
				continue
			}
			// Attempt delivery and update its associated record.
			err = tp.Deliver(ctx, failure.Payload, failure.DeliverTo)
			if err != nil {
				util.ErrorLogger.Errorf("retrier failed in an attempt to retry delivery: %s", err)
				if failure.NAttempts >= r.abandonLimit {
					err = r.da.MarkAbandonedAttempt(c, failure.ID)
					if err != nil {
						util.ErrorLogger.Errorf("retrier failed to mark attempt as abandoned: %s", err)
					}
				} else {
					err = r.da.MarkRetryFailureAttempt(c, failure.ID)
					if err != nil {
						util.ErrorLogger.Errorf("retrier failed to mark attempt as failed: %s", err)
					}
				}
			} else {
				err = r.da.MarkSuccessfulAttempt(c, failure.ID)
				if err != nil {
					util.ErrorLogger.Errorf("retrier failed to mark attempt as successful: %s", err)
				}
			}
		}
		last := failures[len(failures)-1]
		failures, err = r.da.NextPageRetryableFailures(c, last.ID, last.FetchTime, r.pageSize)
		if err != nil {
			util.ErrorLogger.Errorf("retrier failed to obtain the next page of retriable failures: %s", err)
			return
		}
	}
}
