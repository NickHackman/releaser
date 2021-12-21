package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/tui/config"
	"github.com/NickHackman/tagger/internal/tui/pages/organizations"
	"github.com/NickHackman/tagger/internal/tui/pages/repositories"
	tea "github.com/charmbracelet/bubbletea"
)

func Execute(gh *service.GitHub, config *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: ui error - %v", err)
			os.Exit(1)
		}
	}()

	var page tea.Model = organizations.New(ctx, gh, config)
	if config.Org != "" {
		page = repositories.New(ctx, gh, config)
	}

	return tea.NewProgram(page, tea.WithAltScreen(), tea.WithMouseAllMotion()).Start()
}
