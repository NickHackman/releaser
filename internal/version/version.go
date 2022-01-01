package version

import (
	"fmt"

	"github.com/Masterminds/semver"
)

type Change string

const (
	IncMajor Change = "major"
	IncMinor Change = "minor"
	IncPatch Change = "patch"
)

const (
	defaultVersion = "v0.1.0"
	vPrefix        = "v"
)

func ChangeFromString(change string) (Change, error) {
	c := Change(change)

	if c != IncMajor && c != IncMinor && c != IncPatch {
		return "", fmt.Errorf("invalid version '%s' expected one of 'major', 'minor', 'patch'", change)
	}

	return c, nil
}

func New(latest string, change Change) string {
	if latest == "" {
		return defaultVersion
	}

	sem, err := semver.NewVersion(latest)
	if err != nil {
		// Not semantically versioned, return the previous. Handles situations like "Nightly", "Beta", etc
		return latest
	}

	var new semver.Version
	switch change {
	case IncMajor:
		new = sem.IncMajor()
	case IncMinor:
		new = sem.IncMinor()
	case IncPatch:
		new = sem.IncPatch()
	default:
		// Impossible
		panic("provided change doesn't exist - this is a bug")
	}

	return vPrefix + new.String()
}
