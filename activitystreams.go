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
	"crypto"
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

func toOrderedCollectionPage(id *url.URL, ids []string, current, length int) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	ocp = streams.NewActivityStreamsOrderedCollectionPage()
	// id
	idProp := streams.NewActivityStreamsIdProperty()
	idProp.Set(id)
	ocp.SetActivityStreamsId(idProp)
	// items
	oiProp := streams.NewActivityStreamsOrderedItemsProperty()
	for _, i := range ids {
		var iri *url.URL
		iri, err = url.Parse(i)
		if err != nil {
			return
		}
		oiProp.AppendIRI(iri)
	}
	ocp.SetActivityStreamsOrderedItems(oiProp)
	// total len
	tlProp := streams.NewActivityStreamsTotalItemsProperty()
	tlProp.Set(oiProp.Len())
	ocp.SetActivityStreamsTotalItems(tlProp)
	return
}

func toPersonActor(a Application,
	scheme, host, username, preferredUsername, summary string,
	pubKey crypto.PublicKey) (p vocab.ActivityStreamsPerson, err error) {
	p = streams.NewActivityStreamsPerson()
	// id
	idProp := streams.NewActivityStreamsIdProperty()
	idIRI := knownUserIRIFor(scheme, host, userPathKey, username)
	idProp.SetIRI(idIRI)
	p.SetActivityStreamsId(idProp)

	// inbox
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxIRI := knownUserIRIFor(scheme, host, inboxPathKey, username)
	inboxProp.SetIRI(inboxIRI)
	p.SetActivityStreamsInbox(inboxProp)

	// outbox
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxIRI := knownUserIRIFor(scheme, host, outboxPathKey, username)
	outboxProp.SetIRI(outboxIRI)
	p.SetActivityStreamsOutbox(outboxProp)

	// followers
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersIRI := knownUserIRIFor(scheme, host, followersPathKey, username)
	followersProp.SetIRI(followersIRI)
	p.SetActivityStreamsFollowers(followersProp)

	// following
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingIRI := knownUserIRIFor(scheme, host, followingPathKey, username)
	followingProp.SetIRI(followingIRI)
	p.SetActivityStreamsFollowing(followingProp)

	// liked
	likedProp := streams.NewActivityStreamsLikedProperty()
	likedIRI := knownUserIRIFor(scheme, host, likedPathKey, username)
	likedProp.SetIRI(likedIRI)
	p.SetActivityStreamsLiked(likedProp)

	// name
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(username)
	p.SetActivityStreamsName(nameProp)

	// preferredUsername
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(preferredUsername)
	p.SetActivityStreamsPreferredUsername(preferredUsernameProp)

	// url
	urlProp := streams.NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(idIRI)
	p.SetActivityStreamsUrl(urlProp)

	// summary
	summaryProp := streams.NewActivityStreamsSummaryProperty()
	summaryProp.AppendXMLSchemaString(summary)
	p.SetActivityStreamsSummary(summaryProp)

	// publicKey property
	publicKeyProp := streams.NewW3IDSecurityV1PublicKeyProperty()

	// publicKey type
	publicKeyType := streams.NewW3IDSecurityV1PublicKey()

	// publicKey id
	pubKeyIdProp := streams.NewActivityStreamsIdProperty()
	pubKeyIRI := knownUserIRIFor(scheme, host, pubKeyKey, username)
	pubKeyIdProp.SetIRI(pubKeyIRI)
	publicKeyType.SetActivityStreamsId(pubKeyIdProp)

	// publicKey owner
	ownerProp := streams.NewW3IDSecurityV1OwnerProperty()
	ownerProp.SetIRI(idIRI)
	publicKeyType.SetW3IDSecurityV1Owner(ownerProp)

	// publicKey publicKeyPem
	publicKeyPemProp := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	var pubStr string
	pubStr, err = marshalPublicKey(pubKey)
	if err != nil {
		return
	}
	publicKeyPemProp.Set(pubStr)
	publicKeyType.SetW3IDSecurityV1PublicKeyPem(publicKeyPemProp)

	publicKeyProp.AppendW3IDSecurityV1PublicKey(publicKeyType)
	p.SetW3IDSecurityV1PublicKey(publicKeyProp)
	return
}
