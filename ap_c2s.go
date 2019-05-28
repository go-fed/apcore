package apcore

import (
	"context"
	"net/http"

	"github.com/go-fed/activity/pub"
)

var _ pub.SocialProtocol = &socialBehavior{}

type socialBehavior struct {
	db *database
}

func (s *socialBehavior) AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (authenticated bool, err error) {
	// TODO
	return
}

func (s *socialBehavior) Callbacks(c context.Context) (wrapped pub.SocialWrappedCallbacks, other []interface{}) {
	// TODO
	return
}

func (s *socialBehavior) DefaultCallback(c context.Context, activity pub.Activity) error {
	// TODO
	return nil
}
