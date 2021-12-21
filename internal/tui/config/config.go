package config

import "time"

type Config struct {
	Timeout        time.Duration
	TemplateString string
	Org            string
	Width, Height  int
}
