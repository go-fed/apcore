// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
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

package main

import (
	"fmt"
	"time"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/models"
)

var (
	testActor1          models.ActivityStreamsPerson
	testActor1Inbox     models.ActivityStreamsOrderedCollection
	testActor2Inbox     models.ActivityStreamsOrderedCollection
	testActor3Inbox     models.ActivityStreamsOrderedCollection
	testActor1Outbox    models.ActivityStreamsOrderedCollection
	testActor2Outbox    models.ActivityStreamsOrderedCollection
	testActor3Outbox    models.ActivityStreamsOrderedCollection
	testActivity1       vocab.ActivityStreamsMove     // Federated
	testActivity2       vocab.ActivityStreamsCreate   // Federated
	testActivity3       vocab.ActivityStreamsListen   // Federated
	testActivity4       vocab.ActivityStreamsActivity // Local
	testActivity5       vocab.ActivityStreamsTravel   // Local
	testActivity6       vocab.ActivityStreamsAccept   // Local
	testActivity7       vocab.ActivityStreamsListen   // Federated, Public
	testActivity8       vocab.ActivityStreamsAccept   // Local, Public
	testActor1Followers models.ActivityStreamsCollection
	testActor2Followers models.ActivityStreamsCollection
	testActor3Followers models.ActivityStreamsCollection
	testActor1Following models.ActivityStreamsCollection
	testActor2Following models.ActivityStreamsCollection
	testActor3Following models.ActivityStreamsCollection
	testActor1Liked     models.ActivityStreamsCollection
	testActor2Liked     models.ActivityStreamsCollection
	testActor3Liked     models.ActivityStreamsCollection
)

const (
	testActor1PreferredUsername = "testPreferredUsername"
	testEmail1                  = "test@example.com"
	testActor1IRI               = "https://example.com/actors/test1"
	testActor2IRI               = "https://example.com/actors/test2"
	testActor3IRI               = "https://example.com/actors/test3"
	testPeerActor1InboxIRI      = "https://fed.example.com/actors/test1/inbox"
	testActor1InboxIRI          = "https://example.com/actors/test1/inbox"
	testActor2InboxIRI          = "https://example.com/actors/test2/inbox"
	testActor3InboxIRI          = "https://example.com/actors/test3/inbox"
	testActor1OutboxIRI         = "https://example.com/actors/test1/outbox"
	testActor2OutboxIRI         = "https://example.com/actors/test2/outbox"
	testActor3OutboxIRI         = "https://example.com/actors/test3/outbox"
	testActivity1IRI            = "https://fed.example.com/activities/test1"
	testActivity2IRI            = "https://fed.example.com/activities/test2"
	testActivity3IRI            = "https://fed.example.com/activities/test3"
	testActivity4IRI            = "https://example.com/activities/test4"
	testActivity5IRI            = "https://example.com/activities/test5"
	testActivity6IRI            = "https://example.com/activities/test6"
	testActivity7IRI            = "https://fed.example.com/activities/test7"
	testActivity8IRI            = "https://example.com/activities/test8"
	testActor1FollowersIRI      = "https://example.com/actors/test1/followers"
	testActor2FollowersIRI      = "https://example.com/actors/test2/followers"
	testActor3FollowersIRI      = "https://example.com/actors/test3/followers"
	testActor1FollowingIRI      = "https://example.com/actors/test1/following"
	testActor2FollowingIRI      = "https://example.com/actors/test2/following"
	testActor3FollowingIRI      = "https://example.com/actors/test3/following"
	testActor1LikedIRI          = "https://example.com/actors/test1/liked"
	testActor2LikedIRI          = "https://example.com/actors/test2/liked"
	testActor3LikedIRI          = "https://example.com/actors/test3/liked"
)

func init() {
	initTestActor1()
	initTestActor1Inbox()
	initTestActor2Inbox()
	initTestActor3Inbox()
	initTestActor1Outbox()
	initTestActor2Outbox()
	initTestActor3Outbox()
	initTestActivity1()
	initTestActivity2()
	initTestActivity3()
	initTestActivity4()
	initTestActivity5()
	initTestActivity6()
	initTestActivity7()
	initTestActivity8()
	initTestActor1Followers()
	initTestActor2Followers()
	initTestActor3Followers()
	initTestActor1Following()
	initTestActor2Following()
	initTestActor3Following()
	initTestActor1Liked()
	initTestActor2Liked()
	initTestActor3Liked()
}

func initTestActor1() {
	testActor1 = models.ActivityStreamsPerson{
		streams.NewActivityStreamsPerson(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor1IRI))
	testActor1.SetJSONLDId(idP)

	ibox := streams.NewActivityStreamsInboxProperty()
	ibox.SetIRI(mustParse(testActor1InboxIRI))
	testActor1.SetActivityStreamsInbox(ibox)

	obox := streams.NewActivityStreamsOutboxProperty()
	obox.SetIRI(mustParse(testActor1OutboxIRI))
	testActor1.SetActivityStreamsOutbox(obox)

	pref := streams.NewActivityStreamsPreferredUsernameProperty()
	pref.SetXMLSchemaString(testActor1PreferredUsername)
	testActor1.SetActivityStreamsPreferredUsername(pref)
}

func initTestActivity1() {
	testActivity1 = streams.NewActivityStreamsMove()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity1IRI))
	testActivity1.SetJSONLDId(idP)
	sally := mustParse("http://sally.example.org")
	obj := streams.NewActivityStreamsDocument()
	nameObj := streams.NewActivityStreamsNameProperty()
	nameObj.AppendXMLSchemaString("sales figures")
	obj.SetActivityStreamsName(nameObj)
	origin := streams.NewActivityStreamsCollection()
	nameOrigin := streams.NewActivityStreamsNameProperty()
	nameOrigin.AppendXMLSchemaString("Folder A")
	origin.SetActivityStreamsName(nameOrigin)
	target := streams.NewActivityStreamsCollection()
	nameTarget := streams.NewActivityStreamsNameProperty()
	nameTarget.AppendXMLSchemaString("Folder B")
	target.SetActivityStreamsName(nameTarget)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally moved the sales figures from Folder A to Folder B")
	testActivity1.SetActivityStreamsSummary(summary)
	objectActor := streams.NewActivityStreamsActorProperty()
	objectActor.AppendIRI(sally)
	testActivity1.SetActivityStreamsActor(objectActor)
	object := streams.NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsDocument(obj)
	testActivity1.SetActivityStreamsObject(object)
	originProp := streams.NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsCollection(origin)
	testActivity1.SetActivityStreamsOrigin(originProp)
	tobj := streams.NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsCollection(target)
	testActivity1.SetActivityStreamsTarget(tobj)
}

func initTestActivity2() {
	testActivity2 = streams.NewActivityStreamsCreate()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity2IRI))
	testActivity2.SetJSONLDId(idP)
	friend := mustParse("http://purl.org/vocab/relationship/friendOf")
	m := mustParse("http://matt.example.org")
	s := mustParse("http://sally.example.org")
	t, err := time.Parse(time.RFC3339, "2015-04-21T12:34:56Z")
	if err != nil {
		panic(err)
	}
	relationship := streams.NewActivityStreamsRelationship()
	subj := streams.NewActivityStreamsSubjectProperty()
	subj.SetIRI(s)
	relationship.SetActivityStreamsSubject(subj)
	friendRel := streams.NewActivityStreamsRelationshipProperty()
	friendRel.AppendIRI(friend)
	relationship.SetActivityStreamsRelationship(friendRel)
	obj := streams.NewActivityStreamsObjectProperty()
	obj.AppendIRI(m)
	relationship.SetActivityStreamsObject(obj)
	startTime := streams.NewActivityStreamsStartTimeProperty()
	startTime.Set(t)
	relationship.SetActivityStreamsStartTime(startTime)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally became a friend of Matt")
	testActivity2.SetActivityStreamsSummary(summary)
	objectActor := streams.NewActivityStreamsActorProperty()
	objectActor.AppendIRI(s)
	testActivity2.SetActivityStreamsActor(objectActor)
	objRoot := streams.NewActivityStreamsObjectProperty()
	objRoot.AppendActivityStreamsRelationship(relationship)
	testActivity2.SetActivityStreamsObject(objRoot)
}

func initTestActivity3() {
	testActivity3 = streams.NewActivityStreamsListen()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity3IRI))
	testActivity3.SetJSONLDId(idP)
	u := mustParse("http://example.org/foo.mp3")
	actor := streams.NewActivityStreamsPerson()
	sally := streams.NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	service := streams.NewActivityStreamsService()
	serviceName := streams.NewActivityStreamsNameProperty()
	serviceName.AppendXMLSchemaString("Acme Music Service")
	service.SetActivityStreamsName(serviceName)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally listened to a piece of music on the Acme Music Service")
	testActivity3.SetActivityStreamsSummary(summary)
	rootActor := streams.NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	testActivity3.SetActivityStreamsActor(rootActor)
	obj := streams.NewActivityStreamsObjectProperty()
	obj.AppendIRI(u)
	testActivity3.SetActivityStreamsObject(obj)
	inst := streams.NewActivityStreamsInstrumentProperty()
	inst.AppendActivityStreamsService(service)
	testActivity3.SetActivityStreamsInstrument(inst)
}

func initTestActivity4() {
	testActivity4 = streams.NewActivityStreamsActivity()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity4IRI))
	testActivity4.SetJSONLDId(idP)
	person := streams.NewActivityStreamsPerson()
	sally := streams.NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	note := streams.NewActivityStreamsNote()
	aNote := streams.NewActivityStreamsNameProperty()
	aNote.AppendXMLSchemaString("A Note")
	note.SetActivityStreamsName(aNote)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally did something to a note")
	testActivity4.SetActivityStreamsSummary(summary)
	actor := streams.NewActivityStreamsActorProperty()
	actor.AppendActivityStreamsPerson(person)
	testActivity4.SetActivityStreamsActor(actor)
	object := streams.NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsNote(note)
	testActivity4.SetActivityStreamsObject(object)
}

func initTestActivity5() {
	testActivity5 = streams.NewActivityStreamsTravel()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity5IRI))
	testActivity5.SetJSONLDId(idP)
	person := streams.NewActivityStreamsPerson()
	sally := streams.NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	place := streams.NewActivityStreamsPlace()
	work := streams.NewActivityStreamsNameProperty()
	work.AppendXMLSchemaString("Work")
	place.SetActivityStreamsName(work)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally went to work")
	testActivity5.SetActivityStreamsSummary(summary)
	actor := streams.NewActivityStreamsActorProperty()
	actor.AppendActivityStreamsPerson(person)
	testActivity5.SetActivityStreamsActor(actor)
	target := streams.NewActivityStreamsTargetProperty()
	target.AppendActivityStreamsPlace(place)
	testActivity5.SetActivityStreamsTarget(target)
}

func initTestActivity6() {
	testActivity6 = streams.NewActivityStreamsAccept()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity6IRI))
	testActivity6.SetJSONLDId(idP)
	person := streams.NewActivityStreamsPerson()
	sally := streams.NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	event := streams.NewActivityStreamsEvent()
	goingAway := streams.NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	event.SetActivityStreamsName(goingAway)
	invite := streams.NewActivityStreamsInvite()
	actor := streams.NewActivityStreamsActorProperty()
	actor.AppendIRI(mustParse("http://john.example.org"))
	invite.SetActivityStreamsActor(actor)
	object := streams.NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsEvent(event)
	invite.SetActivityStreamsObject(object)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally accepted an invitation to a party")
	testActivity6.SetActivityStreamsSummary(summary)
	rootActor := streams.NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	testActivity6.SetActivityStreamsActor(rootActor)
	inviteObject := streams.NewActivityStreamsObjectProperty()
	inviteObject.AppendActivityStreamsInvite(invite)
	testActivity6.SetActivityStreamsObject(inviteObject)
}

func initTestActivity7() {
	testActivity7 = streams.NewActivityStreamsListen()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity7IRI))
	testActivity7.SetJSONLDId(idP)
	u := mustParse("http://fed.example.org/foo.mp3")
	actor := streams.NewActivityStreamsPerson()
	sally := streams.NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	service := streams.NewActivityStreamsService()
	serviceName := streams.NewActivityStreamsNameProperty()
	serviceName.AppendXMLSchemaString("Acme Music Service")
	service.SetActivityStreamsName(serviceName)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally listened to a piece of music on the Acme Music Service")
	testActivity7.SetActivityStreamsSummary(summary)
	rootActor := streams.NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	testActivity7.SetActivityStreamsActor(rootActor)
	obj := streams.NewActivityStreamsObjectProperty()
	obj.AppendIRI(u)
	testActivity7.SetActivityStreamsObject(obj)
	inst := streams.NewActivityStreamsInstrumentProperty()
	inst.AppendActivityStreamsService(service)
	testActivity7.SetActivityStreamsInstrument(inst)
	to := streams.NewActivityStreamsToProperty()
	to.AppendIRI(mustParse("https://www.w3.org/ns/activitystreams#Public"))
	testActivity7.SetActivityStreamsTo(to)
}

func initTestActivity8() {
	testActivity8 = streams.NewActivityStreamsAccept()
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActivity8IRI))
	testActivity8.SetJSONLDId(idP)
	person := streams.NewActivityStreamsPerson()
	sally := streams.NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	event := streams.NewActivityStreamsEvent()
	goingAway := streams.NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	event.SetActivityStreamsName(goingAway)
	invite := streams.NewActivityStreamsInvite()
	actor := streams.NewActivityStreamsActorProperty()
	actor.AppendIRI(mustParse("http://john.example.org"))
	invite.SetActivityStreamsActor(actor)
	object := streams.NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsEvent(event)
	invite.SetActivityStreamsObject(object)
	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally accepted an invitation to a party")
	testActivity8.SetActivityStreamsSummary(summary)
	rootActor := streams.NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	testActivity8.SetActivityStreamsActor(rootActor)
	inviteObject := streams.NewActivityStreamsObjectProperty()
	inviteObject.AppendActivityStreamsInvite(invite)
	testActivity8.SetActivityStreamsObject(inviteObject)
	to := streams.NewActivityStreamsToProperty()
	to.AppendIRI(mustParse("https://www.w3.org/ns/activitystreams#Public"))
	testActivity8.SetActivityStreamsTo(to)
}

func initTestActor1Inbox() {
	testActor1Inbox = models.ActivityStreamsOrderedCollection{
		streams.NewActivityStreamsOrderedCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor1InboxIRI))
	testActor1Inbox.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	testActor1Inbox.SetActivityStreamsTotalItems(totalItems)
	orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendIRI(mustParse(testActivity3IRI))
	orderedItems.AppendIRI(mustParse(testActivity1IRI))
	testActor1Inbox.SetActivityStreamsOrderedItems(orderedItems)
}

func initTestActor2Inbox() {
	testActor2Inbox = models.ActivityStreamsOrderedCollection{
		streams.NewActivityStreamsOrderedCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor2InboxIRI))
	testActor2Inbox.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(100)
	testActor2Inbox.SetActivityStreamsTotalItems(totalItems)
	orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	for i := 0; i < 100; i++ {
		orderedItems.AppendIRI(mustParse(fmt.Sprintf("https://long.example.com/%d", i)))
	}
	testActor2Inbox.SetActivityStreamsOrderedItems(orderedItems)
}

func initTestActor3Inbox() {
	testActor3Inbox = models.ActivityStreamsOrderedCollection{
		streams.NewActivityStreamsOrderedCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor3InboxIRI))
	testActor3Inbox.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(4)
	testActor3Inbox.SetActivityStreamsTotalItems(totalItems)
	orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendIRI(mustParse(testActivity1IRI))
	orderedItems.AppendIRI(mustParse(testActivity8IRI))
	orderedItems.AppendIRI(mustParse(testActivity7IRI))
	orderedItems.AppendIRI(mustParse(testActivity4IRI))
	testActor3Inbox.SetActivityStreamsOrderedItems(orderedItems)
}

func initTestActor1Outbox() {
	testActor1Outbox = models.ActivityStreamsOrderedCollection{
		streams.NewActivityStreamsOrderedCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor1OutboxIRI))
	testActor1Outbox.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	testActor1Outbox.SetActivityStreamsTotalItems(totalItems)
	orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendIRI(mustParse(testActivity3IRI))
	orderedItems.AppendIRI(mustParse(testActivity1IRI))
	testActor1Outbox.SetActivityStreamsOrderedItems(orderedItems)
}

func initTestActor2Outbox() {
	testActor2Outbox = models.ActivityStreamsOrderedCollection{
		streams.NewActivityStreamsOrderedCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor2OutboxIRI))
	testActor2Outbox.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(100)
	testActor2Outbox.SetActivityStreamsTotalItems(totalItems)
	orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	for i := 0; i < 100; i++ {
		orderedItems.AppendIRI(mustParse(fmt.Sprintf("https://long.example.com/%d", i)))
	}
	testActor2Outbox.SetActivityStreamsOrderedItems(orderedItems)
}

func initTestActor3Outbox() {
	testActor3Outbox = models.ActivityStreamsOrderedCollection{
		streams.NewActivityStreamsOrderedCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor3OutboxIRI))
	testActor3Outbox.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(4)
	testActor3Outbox.SetActivityStreamsTotalItems(totalItems)
	orderedItems := streams.NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendIRI(mustParse(testActivity1IRI))
	orderedItems.AppendIRI(mustParse(testActivity8IRI))
	orderedItems.AppendIRI(mustParse(testActivity7IRI))
	orderedItems.AppendIRI(mustParse(testActivity4IRI))
	testActor3Outbox.SetActivityStreamsOrderedItems(orderedItems)
}

func initTestActor1Followers() {
	testActor1Followers = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor1FollowersIRI))
	testActor1Followers.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	testActor1Followers.SetActivityStreamsTotalItems(totalItems)
	items := streams.NewActivityStreamsItemsProperty()
	items.AppendIRI(mustParse(testActor2IRI))
	items.AppendIRI(mustParse(testActor3IRI))
	testActor1Followers.SetActivityStreamsItems(items)
}

func initTestActor2Followers() {
	testActor2Followers = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor2FollowersIRI))
	testActor2Followers.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(100)
	testActor2Followers.SetActivityStreamsTotalItems(totalItems)
	items := streams.NewActivityStreamsItemsProperty()
	for i := 0; i < 100; i++ {
		items.AppendIRI(mustParse(fmt.Sprintf("https://long.example.com/actor%d", i)))
	}
	testActor2Followers.SetActivityStreamsItems(items)
}

func initTestActor3Followers() {
	testActor3Followers = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor3FollowersIRI))
	testActor3Followers.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(0)
	testActor3Followers.SetActivityStreamsTotalItems(totalItems)
}

func initTestActor1Following() {
	testActor1Following = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor1FollowingIRI))
	testActor1Following.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	testActor1Following.SetActivityStreamsTotalItems(totalItems)
	items := streams.NewActivityStreamsItemsProperty()
	items.AppendIRI(mustParse(testActor2IRI))
	items.AppendIRI(mustParse(testActor3IRI))
	testActor1Following.SetActivityStreamsItems(items)
}

func initTestActor2Following() {
	testActor2Following = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor2FollowingIRI))
	testActor2Following.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(100)
	testActor2Following.SetActivityStreamsTotalItems(totalItems)
	items := streams.NewActivityStreamsItemsProperty()
	for i := 0; i < 100; i++ {
		items.AppendIRI(mustParse(fmt.Sprintf("https://long.example.com/actor%d", i)))
	}
	testActor2Following.SetActivityStreamsItems(items)
}

func initTestActor3Following() {
	testActor3Following = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor3FollowingIRI))
	testActor3Following.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(0)
	testActor3Following.SetActivityStreamsTotalItems(totalItems)
}

func initTestActor1Liked() {
	testActor1Liked = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor1LikedIRI))
	testActor1Liked.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	testActor1Liked.SetActivityStreamsTotalItems(totalItems)
	items := streams.NewActivityStreamsItemsProperty()
	items.AppendIRI(mustParse(testActivity2IRI))
	items.AppendIRI(mustParse(testActivity3IRI))
	testActor1Liked.SetActivityStreamsItems(items)
}

func initTestActor2Liked() {
	testActor2Liked = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor2LikedIRI))
	testActor2Liked.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(100)
	testActor2Liked.SetActivityStreamsTotalItems(totalItems)
	items := streams.NewActivityStreamsItemsProperty()
	for i := 0; i < 100; i++ {
		items.AppendIRI(mustParse(fmt.Sprintf("https://long.example.com/note%d", i)))
	}
	testActor2Liked.SetActivityStreamsItems(items)
}

func initTestActor3Liked() {
	testActor3Liked = models.ActivityStreamsCollection{
		streams.NewActivityStreamsCollection(),
	}
	idP := streams.NewJSONLDIdProperty()
	idP.SetIRI(mustParse(testActor3LikedIRI))
	testActor3Liked.SetJSONLDId(idP)
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(0)
	testActor3Liked.SetActivityStreamsTotalItems(totalItems)
}
