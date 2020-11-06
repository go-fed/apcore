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
	"bytes"
	"context"
	"crypto"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework/config"
	"github.com/go-fed/apcore/framework/web"
	"github.com/go-fed/apcore/services"
	"github.com/go-fed/apcore/util"
	"github.com/go-fed/httpsig"
	"golang.org/x/time/rate"
)

const (
	activityStreamsContentType = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
)

func containsRequiredHttpHeaders(method string, headers []string) error {
	var hasRequestTarget, hasDate, hasDigest bool
	for _, header := range headers {
		hasRequestTarget = hasRequestTarget || header == httpsig.RequestTarget
		hasDate = hasDate || header == "Date"
		hasDigest = hasDigest || header == "Digest"
	}
	if !hasRequestTarget {
		return fmt.Errorf("missing http header for %s: %s", method, httpsig.RequestTarget)
	} else if !hasDate {
		return fmt.Errorf("missing http header for %s: Date", method)
	} else if !hasDigest {
		return fmt.Errorf("missing http header for %s: Digest", method)
	}
	return nil
}

// TODO: re-launch existing failed deliveries at startup and control rate-limiting.

type Controller struct {
	a           app.Application
	clock       pub.Clock
	client      *http.Client
	algs        []httpsig.Algorithm
	digestAlg   httpsig.DigestAlgorithm
	getHeaders  []string
	postHeaders []string
	l           *rate.Limiter // TODO: Use this
	da          *services.DeliveryAttempts
}

func NewController(
	c *config.Config,
	a app.Application,
	clock pub.Clock,
	client *http.Client,
	da *services.DeliveryAttempts) (tc *Controller, err error) {
	if c.ActivityPubConfig.OutboundRateLimitQPS <= 0 {
		err = fmt.Errorf("outbound rate limit qps is <= 0")
		return
	} else if c.ActivityPubConfig.OutboundRateLimitBurst <= 0 {
		err = fmt.Errorf("outbound rate limit burst is <= 0")
		return
	} else if len(c.ActivityPubConfig.HttpSignaturesConfig.Algorithms) == 0 {
		err = fmt.Errorf("no httpsig algorithms specified")
		return
	} else if err = containsRequiredHttpHeaders(http.MethodGet, c.ActivityPubConfig.HttpSignaturesConfig.GetHeaders); err != nil {
		return
	} else if err = containsRequiredHttpHeaders(http.MethodPost, c.ActivityPubConfig.HttpSignaturesConfig.PostHeaders); err != nil {
		return
	} else if !httpsig.IsSupportedDigestAlgorithm(c.ActivityPubConfig.HttpSignaturesConfig.DigestAlgorithm) {
		err = fmt.Errorf("unsupported digest algorithm: %s", c.ActivityPubConfig.HttpSignaturesConfig.DigestAlgorithm)
		return
	}
	algos := make([]httpsig.Algorithm, len(c.ActivityPubConfig.HttpSignaturesConfig.Algorithms))
	for i, algo := range c.ActivityPubConfig.HttpSignaturesConfig.Algorithms {
		if !httpsig.IsSupportedHttpSigAlgorithm(algo) {
			err = fmt.Errorf("unsupported httpsig algorithm: %s", algo)
			return
		}
		algos[i] = httpsig.Algorithm(algo)
	}

	return &Controller{
		a:           a,
		clock:       clock,
		client:      client,
		algs:        algos,
		digestAlg:   httpsig.DigestAlgorithm(c.ActivityPubConfig.HttpSignaturesConfig.DigestAlgorithm),
		getHeaders:  c.ActivityPubConfig.HttpSignaturesConfig.GetHeaders,
		postHeaders: c.ActivityPubConfig.HttpSignaturesConfig.PostHeaders,
		l:           rate.NewLimiter(rate.Limit(c.ActivityPubConfig.OutboundRateLimitQPS), c.ActivityPubConfig.OutboundRateLimitBurst),
		da:          da,
	}, err
}

func (tc *Controller) Get(
	privKey crypto.PrivateKey,
	pubKeyId string) (t pub.Transport, err error) {
	var getSigner, postSigner httpsig.Signer
	getSigner, _, err = httpsig.NewSigner(tc.algs, tc.digestAlg, tc.getHeaders, httpsig.Signature)
	if err != nil {
		return
	}
	postSigner, _, err = httpsig.NewSigner(tc.algs, tc.digestAlg, tc.postHeaders, httpsig.Signature)
	if err != nil {
		return
	}
	return newTransport(
		tc.a,
		tc.clock,
		tc.client,
		getSigner,
		postSigner,
		privKey,
		pubKeyId,
		tc)
}

func (tc *Controller) GetFirstAlgorithm() httpsig.Algorithm {
	return tc.algs[0]
}

func (tc *Controller) wait(c util.Context) {
	tc.l.Wait(c)
}

func (tc *Controller) insertAttempt(c util.Context, payload []byte, to *url.URL, fromUUID string) (id string, err error) {
	id, err = tc.da.InsertAttempt(c, fromUUID, to, payload)
	return
}

func (tc *Controller) markSuccess(c util.Context, id string) (err error) {
	err = tc.da.MarkSuccessfulAttempt(c, id)
	return
}

func (tc *Controller) markFailure(c util.Context, id string) (err error) {
	err = tc.da.MarkRetryFailureAttempt(c, id)
	return
}

var _ pub.Transport = &transport{}

type transport struct {
	a                         app.Application
	clock                     pub.Clock
	client                    *http.Client
	getSigner, postSigner     httpsig.Signer
	getSignerMu, postSignerMu *sync.Mutex
	privKey                   crypto.PrivateKey
	pubKeyId                  string
	tc                        *Controller
}

func newTransport(a app.Application,
	clock pub.Clock,
	client *http.Client,
	getSigner, postSigner httpsig.Signer,
	privKey crypto.PrivateKey,
	pubKeyId string,
	tc *Controller) (t *transport, err error) {
	return &transport{
		a:            a,
		clock:        clock,
		client:       client,
		getSigner:    getSigner,
		postSigner:   postSigner,
		getSignerMu:  &sync.Mutex{},
		postSignerMu: &sync.Mutex{},
		privKey:      privKey,
		pubKeyId:     pubKeyId,
		tc:           tc,
	}, nil
}

func (t *transport) Dereference(c context.Context, iri *url.URL) (b []byte, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, iri.String(), nil)
	if err != nil {
		return
	}
	req.WithContext(c)
	req.Header.Add("Accept", activityStreamsContentType)
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("Date", t.date())
	req.Header.Add("User-Agent", t.userAgent())
	t.getSignerMu.Lock()
	err = t.getSigner.SignRequest(t.privKey, t.pubKeyId, req, nil)
	t.getSignerMu.Unlock()
	if err != nil {
		return
	}
	var resp *http.Response
	resp, err = t.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if err = t.handleDereferenceResponse(resp); err != nil {
		return
	}
	b, err = ioutil.ReadAll(resp.Body)
	return
}

func (t *transport) Deliver(c context.Context, b []byte, to *url.URL) (err error) {
	uc := util.Context{c}
	var fromUUID string
	fromUUID, err = uc.UserPathUUID()
	if err != nil {
		err = fmt.Errorf("failed to determine user to deliver on behalf of: %s", err)
		return
	}
	var attemptId string
	if attemptId, err = t.tc.insertAttempt(uc, b, to, fromUUID); err != nil {
		err = fmt.Errorf("failed to create delivery attempt: %s", err)
		return
	}

	byteCopy := make([]byte, len(b))
	copy(byteCopy, b)
	buf := bytes.NewBuffer(byteCopy)
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, to.String(), buf)
	if err != nil {
		return
	}
	req.WithContext(c)
	req.Header.Add("Content-Type", activityStreamsContentType)
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("Date", t.date())
	req.Header.Add("User-Agent", t.userAgent())
	t.postSignerMu.Lock()
	err = t.postSigner.SignRequest(t.privKey, t.pubKeyId, req, b)
	t.postSignerMu.Unlock()
	if err != nil {
		return
	}
	var resp *http.Response
	resp, err = t.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if err = t.handleDeliverResponse(resp); err != nil {
		err2 := t.tc.markFailure(uc, attemptId)
		if err2 != nil {
			err = fmt.Errorf("failed delivery and failed to mark as failure (%d): [%s, %s]", attemptId, err, err2)
		}
		return
	}
	if err = t.tc.markSuccess(uc, attemptId); err != nil {
		err = fmt.Errorf("failed to mark delivery as successful (%d): %s", attemptId, err)
		return
	}
	return
}

func (t *transport) BatchDeliver(c context.Context, b []byte, recipients []*url.URL) (err error) {
	var wg *sync.WaitGroup
	for i, r := range recipients {
		wg.Add(1)
		go func(i int, r *url.URL) {
			err := t.Deliver(c, b, r)
			if err != nil {
				util.ErrorLogger.Errorf("BatchDeliver (%d of %d): %s", i, len(recipients), err)
			}
		}(i, r)
	}
	wg.Wait()
	return
}

func (t *transport) handleDereferenceResponse(r *http.Response) (err error) {
	ok := r.StatusCode == http.StatusOK
	if !ok {
		err = fmt.Errorf("url IRI dereference failed with status (%d): %s", r.StatusCode, r.Status)
	}
	return
}

func (t *transport) handleDeliverResponse(r *http.Response) (err error) {
	ok := r.StatusCode == http.StatusOK ||
		r.StatusCode == http.StatusCreated ||
		r.StatusCode == http.StatusAccepted
	if !ok {
		err = fmt.Errorf("delivery failed with status (%d): %s", r.StatusCode, r.Status)
	}
	return
}

func (t *transport) userAgent() string {
	return web.UserAgent(t.a.Software())
}

func (t *transport) date() string {
	return fmt.Sprintf("%s GMT", t.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05"))
}
