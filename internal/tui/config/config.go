package config

import (
	"time"

	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/version"
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
