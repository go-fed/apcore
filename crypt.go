package apcore

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

func hasPassword(display string) (b bool, err error) {
	p := promptui.Prompt{
		Label:     display,
		IsConfirm: true,
	}
	var s string
	s, err = p.Run()
	if err != nil {
		return
	}
	if s == "Y" {
		b = true
	} else if s == "N" {
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
