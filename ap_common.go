package apcore

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

var _ pub.CommonBehavior = &commonBehavior{}

type commonBehavior struct {
	db *database
}

func newCommonBehavior(db *database) *commonBehavior {
	return &commonBehavior{
		db: db,
	}
}

func (a *commonBehavior) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (a *commonBehavior) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (a *commonBehavior) GetOutbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// TODO
	return
}

func (a *commonBehavior) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	// TODO
	return
}
