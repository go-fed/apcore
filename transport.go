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
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/httpsig"
	"golang.org/x/time/rate"
)

const (
	activityStreamsContentType = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
)

type transportController struct {
	a           Application
	clock       pub.Clock
	client      *http.Client
	algs        []httpsig.Algorithm
	getHeaders  []string
	postHeaders []string
	l           *rate.Limiter
	db          *database
}

func newTransportController(
	a Application,
	clock pub.Clock,
	client *http.Client,
	algs []httpsig.Algorithm,
	getHeaders []string,
	postHeaders []string,
	r rate.Limit, burst int,
	db *database) (tc *transportController, err error) {
	return &transportController{
		a:           a,
		clock:       clock,
		client:      client,
		algs:        algs,
		getHeaders:  getHeaders,
		postHeaders: postHeaders,
		l:           rate.NewLimiter(r, burst),
		db:          db,
	}, err
}

func (tc *transportController) Get(
	privKey crypto.PrivateKey,
	pubKeyId string) (t *transport, err error) {
	var getSigner, postSigner httpsig.Signer
	getSigner, _, err = httpsig.NewSigner(tc.algs, tc.getHeaders, httpsig.Signature)
	if err != nil {
		return
	}
	postSigner, _, err = httpsig.NewSigner(tc.algs, tc.postHeaders, httpsig.Signature)
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

func (tc *transportController) wait(c context.Context) {
	tc.l.Wait(c)
}

func (tc *transportController) insertAttempt() {
	// TODO
}

func (tc *transportController) markSuccess() {
	// TODO
}

func (tc *transportController) markFailure() {
	// TODO
}

func (tc *transportController) markTombstone() {
	// TODO
}

var _ pub.Transport = &transport{}

type transport struct {
	a                         Application
	clock                     pub.Clock
	client                    *http.Client
	getSigner, postSigner     httpsig.Signer
	getSignerMu, postSignerMu *sync.Mutex
	privKey                   crypto.PrivateKey
	pubKeyId                  string
	tc                        *transportController // TODO: Use this
}

func newTransport(a Application,
	clock pub.Clock,
	client *http.Client,
	getSigner, postSigner httpsig.Signer,
	privKey crypto.PrivateKey,
	pubKeyId string,
	tc *transportController) (t *transport, err error) {
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
	err = t.getSigner.SignRequest(t.privKey, t.pubKeyId, req)
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
	// TODO: Better status code handling
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("dereference failed with status (%d): %s", resp.StatusCode, resp.Status)
		return
	}
	b, err = ioutil.ReadAll(resp.Body)
	return
}

func (t *transport) Deliver(c context.Context, b []byte, to *url.URL) (err error) {
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
	req.Header.Add("Digest", t.digest(b))
	t.postSignerMu.Lock()
	err = t.postSigner.SignRequest(t.privKey, t.pubKeyId, req)
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
	// TODO: Better status code handling
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("dereference failed with status (%d): %s", resp.StatusCode, resp.Status)
		return
	}
	return
}

func (t *transport) BatchDeliver(c context.Context, b []byte, recipients []*url.URL) (err error) {
	var wg *sync.WaitGroup
	for _, r := range recipients {
		wg.Add(1)
		go func(r *url.URL) {
			// TODO
		}(r)
	}
	wg.Wait()
	return
}

func (t *transport) userAgent() string {
	return fmt.Sprintf("%s (go-fed/activity go-fed/apcore)", t.a.Software().UserAgent)
}

func (t *transport) date() string {
	return fmt.Sprintf("%s GMT", t.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05"))
}

func (t *transport) digest(b []byte) string {
	sum := sha256.Sum256(b)
	return fmt.Sprintf("SHA-256=%s",
		base64.StdEncoding.EncodeToString(sum[:]))
}
