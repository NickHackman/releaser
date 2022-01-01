package config

import (
	"time"

	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/version"
	"github.com/spf13/viper"
)

type Config struct {
	Timeout        time.Duration
	TemplateString string
	Org            string
	Branch         string
	Width, Height  int
	Releases       chan<- []*github.RepositoryReleaseResponse
	VersionChange  version.Change
}

func (c *Config) Refresh() {
	// TODO: don't silently fail
	if err := viper.ReadInConfig(); err != nil {
		return
	}

	c.Branch = viper.GetString("branch")
	c.Timeout = viper.GetDuration("timeout")
	c.TemplateString = viper.GetString("template")

	change, err := version.ChangeFromString(viper.GetString("version.change"))
	if err != nil {
		return
	}

	c.VersionChange = change
}
