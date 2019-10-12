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
	"github.com/go-fed/activity/pub"
)

type onFollow string

const (
	onFollowAlwaysAccept onFollow = "ALWAYS_ACCEPT"
	onFollowAlwaysReject onFollow = "ALWAYS_REJECT"
	onFollowManual       onFollow = "MANUAL"
)

func toOnFollow(p pub.OnFollowBehavior) onFollow {
	switch p {
	case pub.OnFollowAutomaticallyAccept:
		return onFollowAlwaysAccept
	case pub.OnFollowAutomaticallyReject:
		return onFollowAlwaysReject
	case pub.OnFollowDoNothing:
		fallthrough
	default:
		return onFollowManual
	}
}

type userPreferences struct {
	onFollow onFollow
}

func (u *userPreferences) Load(s scanner) (err error) {
	if err = s.Scan(&u.onFollow); err != nil {
		return
	}
	return
}

func (u userPreferences) OnFollow() pub.OnFollowBehavior {
	switch u.onFollow {
	case onFollowAlwaysAccept:
		return pub.OnFollowAutomaticallyAccept
	case onFollowAlwaysReject:
		return pub.OnFollowAutomaticallyReject
	case onFollowManual:
		fallthrough
	default:
		return pub.OnFollowDoNothing
	}
}
