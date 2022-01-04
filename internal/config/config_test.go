package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	exampleAuth := AuthHosts{
		"example.com": Auth{Username: "example", Token: "token"},
	}

	tests := []struct {
		name       string
		host       string
		inputToken string
		isErr      bool

		expectedToken    string
		expectedUsername string
	}{
		{name: "not empty token", inputToken: "exampleToken", expectedToken: "exampleToken"},
		{name: "empty token - host exists", host: "example.com", expectedToken: "token", expectedUsername: "example"},
		{name: "empty token - host does not exist", host: "invalid.com", isErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := Config{
				AuthHosts: exampleAuth,
				Host:      test.host,
				Token:     test.inputToken,
			}

			err := c.CheckAuth()

			if test.isErr {
				assert.NotNil(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUsername, c.Username)
			assert.Equal(t, test.expectedToken, c.Token)
		})
	}
}

func TestIsRepositoryToRelease(t *testing.T) {
	repositories := []string{"example1", "example2"}

	tests := []struct {
		name           string
		repositoryName string
		shouldRelease  bool
	}{
		{name: "should release", repositoryName: "example1", shouldRelease: true},
		{name: "should not release", repositoryName: "invalid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := Config{
				Repositories: repositories,
			}

			shouldRelease := c.IsRepositoryToRelease(test.repositoryName)
			assert.Equal(t, test.shouldRelease, shouldRelease)
		})
	}
}

func TestConfigDir(t *testing.T) {
	err := os.Setenv("XDG_CONFIG_HOME", "example")
	assert.NoError(t, err)

	dir, err := configDir()
	assert.NoError(t, err)

	assert.Equal(t, "example/releaser", dir)

	err = os.Unsetenv("XDG_CONFIG_HOME")
	assert.NoError(t, err)
}

func TestSetSize(t *testing.T) {
	c := Config{Terminal: &TerminalConfig{}}

	w, h := c.Size()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)

	c.SetSize(5, 6)

	w, h = c.Size()
	assert.Equal(t, 5, w)
	assert.Equal(t, 6, h)
}
