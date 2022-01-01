package config

import (
	"time"

	"github.com/NickHackman/releaser/internal/service"
	"github.com/NickHackman/releaser/internal/version"
)

type Config struct {
	Timeout              time.Duration
	TemplateString       string
	TemplateInstructions string
	Org                  string
	Branch               string
	Width, Height        int
	Releases             chan<- []*service.RepositoryReleaseResponse
	VersionChange        version.Change
}
