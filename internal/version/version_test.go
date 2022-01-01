package version_test

import (
	"testing"

	"github.com/NickHackman/tagger/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		change   version.Change
		latest   string
		expected string
	}{
		{name: "empty", change: version.IncMinor, latest: "", expected: "v0.1.0"},
		{name: "patch", change: version.IncPatch, latest: "v1.0.0", expected: "v1.0.1"},
		{name: "minor", change: version.IncMinor, latest: "v1.0.0", expected: "v1.1.0"},
		{name: "major", change: version.IncMajor, latest: "v1.0.0", expected: "v2.0.0"},
		{name: "beta", change: version.IncMajor, latest: "beta", expected: "beta"},
		{name: "nightly", change: version.IncMajor, latest: "nightly", expected: "nightly"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, version.New(test.latest, test.change))
		})
	}
}

func TestChangeFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected version.Change
		isErr    bool
	}{
		{name: "invalid", input: "invalid", isErr: true},
		{name: "patch", input: "patch", expected: version.IncPatch},
		{name: "minor", input: "minor", expected: version.IncMinor},
		{name: "major", input: "major", expected: version.IncMajor},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			change, err := version.ChangeFromString(test.input)
			if test.isErr {
				assert.NotNil(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expected, change)
		})
	}
}
