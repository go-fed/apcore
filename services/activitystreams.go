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

package services

import (
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/paths"
)

// addNextPrev adds the 'next' and 'prev' properties onto a page, if required.
func addNextPrev(page vocab.ActivityStreamsOrderedCollectionPage, start, n int, isEnd bool) error {
	iri, err := pub.GetId(page)
	if err != nil {
		return err
	}
	// Prev
	if start > 0 {
		pStart := start - n
		if pStart < 0 {
			pStart = 0
		}
		prev := streams.NewActivityStreamsPrevProperty()
		prev.SetIRI(paths.AddPageParams(iri, pStart, n))
		page.SetActivityStreamsPrev(prev)
	}
	// Next
	if !isEnd {
		next := streams.NewActivityStreamsNextProperty()
		next.SetIRI(paths.AddPageParams(iri, start+n, n))
		page.SetActivityStreamsNext(next)
	}
	return nil
}

// addNextPrevCol adds the 'next' and 'prev' properties onto a page, if required.
func addNextPrevCol(page vocab.ActivityStreamsCollectionPage, start, n int, isEnd bool) error {
	iri, err := pub.GetId(page)
	if err != nil {
		return err
	}
	// Prev
	if start > 0 {
		pStart := start - n
		if pStart < 0 {
			pStart = 0
		}
		prev := streams.NewActivityStreamsPrevProperty()
		prev.SetIRI(paths.AddPageParams(iri, pStart, n))
		page.SetActivityStreamsPrev(prev)
	}
	// Next
	if !isEnd {
		next := streams.NewActivityStreamsNextProperty()
		next.SetIRI(paths.AddPageParams(iri, start+n, n))
		page.SetActivityStreamsNext(next)
	}
	return nil
}

func toPersonActor(a app.Application,
	uuid paths.UUID,
	scheme, host, username, preferredUsername, summary string,
	pubKey string) (vocab.ActivityStreamsPerson, *url.URL) {
	p := streams.NewActivityStreamsPerson()
	// id
	idProp := streams.NewJSONLDIdProperty()
	idIRI := paths.UUIDIRIFor(scheme, host, paths.UserPathKey, uuid)
	idProp.SetIRI(idIRI)
	p.SetJSONLDId(idProp)

	// inbox
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxIRI := paths.UUIDIRIFor(scheme, host, paths.InboxPathKey, uuid)
	inboxProp.SetIRI(inboxIRI)
	p.SetActivityStreamsInbox(inboxProp)

	// outbox
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxIRI := paths.UUIDIRIFor(scheme, host, paths.OutboxPathKey, uuid)
	outboxProp.SetIRI(outboxIRI)
	p.SetActivityStreamsOutbox(outboxProp)

	// followers
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersIRI := paths.UUIDIRIFor(scheme, host, paths.FollowersPathKey, uuid)
	followersProp.SetIRI(followersIRI)
	p.SetActivityStreamsFollowers(followersProp)

	// following
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingIRI := paths.UUIDIRIFor(scheme, host, paths.FollowingPathKey, uuid)
	followingProp.SetIRI(followingIRI)
	p.SetActivityStreamsFollowing(followingProp)

	// liked
	likedProp := streams.NewActivityStreamsLikedProperty()
	likedIRI := paths.UUIDIRIFor(scheme, host, paths.LikedPathKey, uuid)
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
	pubKeyIdProp := streams.NewJSONLDIdProperty()
	pubKeyIRI := paths.UUIDIRIFor(scheme, host, paths.HttpSigPubKeyKey, uuid)
	pubKeyIdProp.SetIRI(pubKeyIRI)
	publicKeyType.SetJSONLDId(pubKeyIdProp)

	// publicKey owner
	ownerProp := streams.NewW3IDSecurityV1OwnerProperty()
	ownerProp.SetIRI(idIRI)
	publicKeyType.SetW3IDSecurityV1Owner(ownerProp)

	// publicKey publicKeyPem
	publicKeyPemProp := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	publicKeyPemProp.Set(pubKey)
	publicKeyType.SetW3IDSecurityV1PublicKeyPem(publicKeyPemProp)

	publicKeyProp.AppendW3IDSecurityV1PublicKey(publicKeyType)
	p.SetW3IDSecurityV1PublicKey(publicKeyProp)
	return p, idIRI
}

func emptyInbox(actorID *url.URL) (vocab.ActivityStreamsOrderedCollection, error) {
	id, err := paths.IRIForActorID(paths.InboxPathKey, actorID)
	if err != nil {
		return nil, err
	}
	first, err := paths.IRIForActorID(paths.InboxFirstPathKey, actorID)
	if err != nil {
		return nil, err
	}
	last, err := paths.IRIForActorID(paths.InboxLastPathKey, actorID)
	if err != nil {
		return nil, err
	}
	return emptyOrderedCollection(id, first, last), nil
}

func emptyOutbox(actorID *url.URL) (vocab.ActivityStreamsOrderedCollection, error) {
	id, err := paths.IRIForActorID(paths.OutboxPathKey, actorID)
	if err != nil {
		return nil, err
	}
	first, err := paths.IRIForActorID(paths.OutboxFirstPathKey, actorID)
	if err != nil {
		return nil, err
	}
	last, err := paths.IRIForActorID(paths.OutboxLastPathKey, actorID)
	if err != nil {
		return nil, err
	}
	return emptyOrderedCollection(id, first, last), nil
}

func emptyOrderedCollection(id, first, last *url.URL) vocab.ActivityStreamsOrderedCollection {
	oc := streams.NewActivityStreamsOrderedCollection()
	// id
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	oc.SetJSONLDId(idProp)

	// totalItems
	tiProp := streams.NewActivityStreamsTotalItemsProperty()
	tiProp.Set(0)
	oc.SetActivityStreamsTotalItems(tiProp)

	// orderedItems
	oiProp := streams.NewActivityStreamsOrderedItemsProperty()
	oc.SetActivityStreamsOrderedItems(oiProp)

	// first
	firstProp := streams.NewActivityStreamsFirstProperty()
	firstProp.SetIRI(first)
	oc.SetActivityStreamsFirst(firstProp)

	// last
	lastProp := streams.NewActivityStreamsLastProperty()
	lastProp.SetIRI(last)
	oc.SetActivityStreamsLast(lastProp)

	return oc
}

func emptyFollowers(actorID *url.URL) (vocab.ActivityStreamsCollection, error) {
	id, err := paths.IRIForActorID(paths.FollowersPathKey, actorID)
	if err != nil {
		return nil, err
	}
	first, err := paths.IRIForActorID(paths.FollowersFirstPathKey, actorID)
	if err != nil {
		return nil, err
	}
	last, err := paths.IRIForActorID(paths.FollowersLastPathKey, actorID)
	if err != nil {
		return nil, err
	}
	return emptyCollection(id, first, last), nil
}

func emptyFollowing(actorID *url.URL) (vocab.ActivityStreamsCollection, error) {
	id, err := paths.IRIForActorID(paths.FollowingPathKey, actorID)
	if err != nil {
		return nil, err
	}
	first, err := paths.IRIForActorID(paths.FollowingFirstPathKey, actorID)
	if err != nil {
		return nil, err
	}
	last, err := paths.IRIForActorID(paths.FollowingLastPathKey, actorID)
	if err != nil {
		return nil, err
	}
	return emptyCollection(id, first, last), nil
}

func emptyLiked(actorID *url.URL) (vocab.ActivityStreamsCollection, error) {
	id, err := paths.IRIForActorID(paths.LikedPathKey, actorID)
	if err != nil {
		return nil, err
	}
	first, err := paths.IRIForActorID(paths.LikedFirstPathKey, actorID)
	if err != nil {
		return nil, err
	}
	last, err := paths.IRIForActorID(paths.LikedLastPathKey, actorID)
	if err != nil {
		return nil, err
	}
	return emptyCollection(id, first, last), nil
}

func emptyCollection(id, first, last *url.URL) vocab.ActivityStreamsCollection {
	oc := streams.NewActivityStreamsCollection()
	// id
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	oc.SetJSONLDId(idProp)

	// totalItems
	tiProp := streams.NewActivityStreamsTotalItemsProperty()
	tiProp.Set(0)
	oc.SetActivityStreamsTotalItems(tiProp)

	// items
	oiProp := streams.NewActivityStreamsItemsProperty()
	oc.SetActivityStreamsItems(oiProp)

	// first
	firstProp := streams.NewActivityStreamsFirstProperty()
	firstProp.SetIRI(first)
	oc.SetActivityStreamsFirst(firstProp)

	// last
	lastProp := streams.NewActivityStreamsLastProperty()
	lastProp.SetIRI(last)
	oc.SetActivityStreamsLast(lastProp)

	return oc
}
