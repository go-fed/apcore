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

package ap

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/framework/conn"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/httpsig"
)

type publicKeyer interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}

func getPublicKeyFromResponse(c context.Context, b []byte, keyId *url.URL) (p crypto.PublicKey, err error) {
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	var t vocab.Type
	t, err = streams.ToType(c, m)
	if err != nil {
		return
	}
	pker, ok := t.(publicKeyer)
	if !ok {
		err = fmt.Errorf("ActivityStreams type cannot be converted to one known to have publicKey property: %T", t)
		return
	}
	pkp := pker.GetW3IDSecurityV1PublicKey()
	if pkp == nil {
		err = fmt.Errorf("publicKey property is not provided")
		return
	}
	var pkpFound vocab.W3IDSecurityV1PublicKey
	for pkpIter := pkp.Begin(); pkpIter != pkp.End(); pkpIter = pkpIter.Next() {
		if !pkpIter.IsW3IDSecurityV1PublicKey() {
			continue
		}
		pkValue := pkpIter.Get()
		var pkId *url.URL
		pkId, err = pub.GetId(pkValue)
		if err != nil {
			return
		}
		if pkId.String() != keyId.String() {
			continue
		}
		pkpFound = pkValue
		break
	}
	if pkpFound == nil {
		err = fmt.Errorf("cannot find publicKey with id: %s", keyId)
		return
	}
	pkPemProp := pkpFound.GetW3IDSecurityV1PublicKeyPem()
	if pkPemProp == nil || !pkPemProp.IsXMLSchemaString() {
		err = fmt.Errorf("publicKeyPem property is not provided or it is not embedded as a value")
		return
	}
	pubKeyPem := pkPemProp.Get()
	var block *pem.Block
	block, _ = pem.Decode([]byte(pubKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		err = fmt.Errorf("could not decode publicKeyPem to PUBLIC KEY pem block type")
		return
	}
	p, err = x509.ParsePKIXPublicKey(block.Bytes)
	return
}

func verifyHttpSignatures(c context.Context,
	r *http.Request,
	p *paths.Paths,
	db *database,
	tc *conn.Controller) (authenticated bool, err error) {
	// 1. Figure out what key we need to verify
	ctx := ctx{c}
	var v httpsig.Verifier
	v, err = httpsig.NewVerifier(r)
	if err != nil {
		return
	}
	kId := v.KeyId()
	var kIdIRI *url.URL
	kIdIRI, err = url.Parse(kId)
	if err != nil {
		return
	}
	// 2. Get our user's credentials
	var userUUID string
	userUUID, err = ctx.UserPathUUID()
	if err != nil {
		return
	}
	var kUUID string
	var privKey *rsa.PrivateKey
	kUUID, privKey, err = db.GetUserPKey(c, userUUID)
	if err != nil {
		return
	}
	var pubKeyURL *url.URL
	pubKeyURL, err = p.PublicKeyPath(userUUID, kUUID)
	if err != nil {
		return
	}
	pubKeyId := pubKeyURL.String()
	// 3. Fetch the public key of the other actor using our credentials
	tp, err := tc.Get(privKey, pubKeyId)
	if err != nil {
		return
	}
	var b []byte
	b, err = tp.Dereference(c, kIdIRI)
	if err != nil {
		return
	}
	pKey, err := getPublicKeyFromResponse(c, b, kIdIRI)
	if err != nil {
		return
	}
	// 4. Verify the other actor's key
	algo := tc.GetFirstAlgorithm()
	authenticated = nil == v.Verify(pKey, algo)
	return
}
