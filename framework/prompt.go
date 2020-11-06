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

package framework

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

func promptYN(display string) (b bool, err error) {
	p := promptui.Prompt{
		Label: display,
		Templates: &promptui.PromptTemplates{
			Prompt:          fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }} {{ "[%s]" | faint }}`, promptui.IconInitial, "y/N"),
			Valid:           fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }} {{ "[%s]" | faint }}`, promptui.IconGood, "y/N"),
			Invalid:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }} {{ "[%s]" | faint }}`, promptui.IconBad, "y/N"),
			ValidationError: fmt.Sprintf(`{{ ">>" | red }} {{ . | red }} {{ "[%s]" | faint }}`, "y/N"),
			Success:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | faint }} {{ "[%s]" | faint }}`, promptui.IconGood, "y/N"),
		},
		Validate: func(input string) error {
			if lower := strings.ToLower(input); lower != "y" && lower != "n" {
				return fmt.Errorf("must be 'y/Y' or 'n/N'")
			}
			return nil
		},
		Default: "n",
	}
	var s string
	s, err = p.Run()
	s = strings.ToLower(s)
	if err != nil {
		return
	}
	if s == "y" {
		b = true
	} else if s == "n" {
		b = false
	} else {
		err = fmt.Errorf("unknown confirm prompt response: %q", s)
	}
	return
}

func promptPassword(display string) (s string, err error) {
	p := promptui.Prompt{
		Label: display,
		Mask:  '*',
	}
	s, err = p.Run()
	return
}

func promptDoesXHavePassword(display string) (b bool, err error) {
	return promptYN(
		fmt.Sprintf(
			"Does %s have a password?",
			display))
}

func PromptFileExistsContinue(path string) (b bool, err error) {
	return promptYN(
		fmt.Sprintf(
			"File exists at: %q. Do you wish to continue?",
			path))
}

func PromptOverwriteExistingFile(path string) (b bool, err error) {
	return promptYN(
		fmt.Sprintf(
			"File exists at: %q. Do you wish to overwrite it?",
			path))
}

func promptString(display string) (s string, err error) {
	s, err = promptStringWithDefault(display, "")
	return
}

func promptStringWithDefault(display, def string) (s string, err error) {
	p := promptui.Prompt{
		Label:     display,
		Default:   def,
		AllowEdit: false,
		Templates: &promptui.PromptTemplates{
			Prompt:          fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconInitial),
			Valid:           fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconGood),
			Invalid:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconBad),
			ValidationError: fmt.Sprintf(`{{ ">>" | red }} {{ . | red }}{{ ":" | bold}}`),
			Success:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | faint }}{{ ":" | bold}}`, promptui.IconGood),
		},
	}
	if len(def) > 0 {
		p.Default = def
	}
	s, err = p.Run()
	return
}

func promptSelection(display string, choices ...string) (s string, err error) {
	p := promptui.Select{
		Label: display,
		Items: choices,
	}
	_, s, err = p.Run()
	if err != nil {
		return
	}
	return
}

func promptIntWithDefault(display string, def int) (v int, err error) {
	p := promptui.Prompt{
		Label:     display,
		Default:   fmt.Sprintf("%d", def),
		AllowEdit: false,
		Validate: func(input string) error {
			_, err := strconv.ParseInt(input, 10, 32)
			if err != nil {
				return fmt.Errorf("Invalid number")
			}
			return nil
		},
		Templates: &promptui.PromptTemplates{
			Prompt:          fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconInitial),
			Valid:           fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconGood),
			Invalid:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconBad),
			ValidationError: fmt.Sprintf(`{{ ">>" | red }} {{ . | red }}{{ ":" | bold}}`),
			Success:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | faint }}{{ ":" | bold}}`, promptui.IconGood),
		},
	}
	var s string
	s, err = p.Run()
	if err != nil {
		return
	}
	var i int64
	i, err = strconv.ParseInt(s, 10, 32)
	v = int(i)
	return
}

func promptFloat64WithDefault(display string, def int) (v float64, err error) {
	p := promptui.Prompt{
		Label:     display,
		Default:   fmt.Sprintf("%d", def),
		AllowEdit: false,
		Validate: func(input string) error {
			_, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return fmt.Errorf("Invalid number")
			}
			return nil
		},
		Templates: &promptui.PromptTemplates{
			Prompt:          fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconInitial),
			Valid:           fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconGood),
			Invalid:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | bold }}{{ ":" | bold}}`, promptui.IconBad),
			ValidationError: fmt.Sprintf(`{{ ">>" | red }} {{ . | red }}{{ ":" | bold}}`),
			Success:         fmt.Sprintf(`{{ "%s" | bold }} {{ . | faint }}{{ ":" | bold}}`, promptui.IconGood),
		},
	}
	var s string
	s, err = p.Run()
	if err != nil {
		return
	}
	v, err = strconv.ParseFloat(s, 64)
	return
}

func promptAdminUser() (username, email, password string, err error) {
	username, err = promptStringWithDefault(
		"Enter the new admin account's username",
		"")
	if err != nil {
		return
	}
	email, err = promptStringWithDefault(
		"Enter the new admin account's email address (will NOT be verified)",
		"")
	if err != nil {
		return
	}
	password, err = promptPassword("Enter the new admin account's password")
	return
}
