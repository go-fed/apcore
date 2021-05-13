package util

import (
	"context"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
)

func ToActivityStreamsFollow(c Context, t vocab.Type) (f vocab.ActivityStreamsFollow, err error) {
	var res *streams.TypeResolver
	res, err = streams.NewTypeResolver(func(c context.Context, follow vocab.ActivityStreamsFollow) error {
		f = follow
		return nil
	})
	if err != nil {
		return
	}
	err = res.Resolve(c, t)
	return
}
