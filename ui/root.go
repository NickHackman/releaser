package ui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/NickHackman/tagger/service"
	"github.com/NickHackman/tagger/ui/pages/organizations"
	tea "github.com/charmbracelet/bubbletea"
)

func Execute(gh *service.GitHub, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error: ui error - %v", err)
			os.Exit(1)
		}
	}()

	orgPage := organizations.New(ctx, gh)
	return tea.NewProgram(orgPage, tea.WithAltScreen(), tea.WithMouseAllMotion()).Start()
}
