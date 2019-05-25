// Package apcore implements a generic, extensible ActivityPub server using the
// go-fed libraries.
//
// Clients must implement the Application interface and call Run.
package apcore

const (
	apcoreName         = "apcore"
	apcoreMajorVersion = 0
	apcoreMinorVersion = 1
	apcorePatchVersion = 0
)

func apCoreSoftware() Software {
	return Software{
		Name:         apcoreName,
		MajorVersion: apcoreMajorVersion,
		MinorVersion: apcoreMinorVersion,
		PatchVersion: apcorePatchVersion,
	}
}
