package apcore

import (
	"fmt"
)

type Software struct {
	Name         string
	MajorVersion int
	MinorVersion int
	PatchVersion int
}

func (s Software) String() string {
	return fmt.Sprintf(
		"%s (%d.%d.%d)",
		s.Name,
		s.MajorVersion,
		s.MinorVersion,
		s.PatchVersion)
}
