package config

import (
	"time"

	"github.com/NickHackman/tagger/internal/service"
)

type Config struct {
	Timeout        time.Duration
	TemplateString string
	Org            string
	Width, Height  int
	Releases       chan<- []*service.RepositoryReleaseResponse
}
