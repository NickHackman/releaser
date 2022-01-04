package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "github.com", input: "github.com", expected: "https://api.github.com/"},
		{name: "GitHub enterprise", input: "git.enterprise.com", expected: "https://git.enterprise.com/api/v3/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := New().Host(test.input)

			assert.Equal(t, test.expected, b.url)
		})
	}
}

func TestBuilderToken(t *testing.T) {
	_, err := New().Build()
	assert.NotNil(t, err)

	assert.Contains(t, err.Error(), "token")
}

func TestBuilderValid(t *testing.T) {
	gh, err := New().Token("token").Host("github.com").Build()
	assert.NoError(t, err)

	assert.NotNil(t, gh)
}
