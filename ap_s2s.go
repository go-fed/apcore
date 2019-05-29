package apcore

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

var _ pub.FederatingProtocol = &federatingBehavior{}

type federatingBehavior struct {
	db *database
}

func newFederatingBehavior(db *database) *federatingBehavior {
	return &federatingBehavior{
		db: db,
	}
}

func (f *federatingBehavior) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (f *federatingBehavior) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	// TODO
	return
}

func (f *federatingBehavior) Callbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}) {
	// TODO
	return
}

func (f *federatingBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	// TODO
	return nil
}

func (f *federatingBehavior) MaxInboxForwardingRecursionDepth(c context.Context) int {
	// TODO
	return -1
}

func (f *federatingBehavior) MaxDeliveryRecursionDepth(c context.Context) int {
	// TODO
	return -1
}

func (f *federatingBehavior) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	// TODO
	return
}

func (f *federatingBehavior) GetInbox(c context.Context, r *http.Request) (ocp vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// TODO
	return
}
