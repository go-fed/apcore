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

package app

// Paths is a set of endpoints for apcore handlers provided out of the box. It
// allows applications to override defaults, for example in case localization
// where "/login" can instead be "/{locale}/login".
//
// The Redirect fields are functions. They accept the current path and can
// return a corresponding path. For example, when redirecting to the homepage
// in a locale usecase, it may be passed "/de-DE/login" which means the function
// then returns "/de-DE".
//
// A zero-value struct is valid and uses apcore defaults.
type Paths struct {
	GetLogin            string
	PostLogin           string
	GetLogout           string
	GetOAuth2Authorize  string
	PostOAuth2Authorize string
	RedirectToHomepage  func(string) string
	RedirectToLogin     func(string) string
}

func (p Paths) getOrDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func (p Paths) GetLoginPath() string {
	return p.getOrDefault(p.GetLogin, "/login")
}

func (p Paths) PostLoginPath() string {
	return p.getOrDefault(p.PostLogin, "/login")
}

func (p Paths) GetLogoutPath() string {
	return p.getOrDefault(p.GetLogout, "/logout")
}

func (p Paths) GetOAuth2AuthorizePath() string {
	return p.getOrDefault(p.GetOAuth2Authorize, "/oauth2/authorize")
}

func (p Paths) PostOAuth2AuthorizePath() string {
	return p.getOrDefault(p.PostOAuth2Authorize, "/oauth2/authorize")
}

func (p Paths) RedirectToHomepagePath(currentPath string) string {
	if p.RedirectToHomepage == nil {
		return "/"
	}
	return p.getOrDefault(p.RedirectToHomepage(currentPath), "/")
}

func (p Paths) RedirectToLoginPath(currentPath string) string {
	if p.RedirectToLogin == nil {
		return "/login"
	}
	return p.getOrDefault(p.RedirectToLogin(currentPath), "/")
}
