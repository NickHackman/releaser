package tui

import (
	"context"
	"fmt"

	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/NickHackman/tagger/internal/tui/config"
	"github.com/NickHackman/tagger/internal/tui/pages/organizations"
	"github.com/NickHackman/tagger/internal/tui/pages/repositories"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Execute(gh *service.GitHub, config *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	releasesChan := make(chan []*service.RepositoryReleaseResponse, 1)
	config.Releases = releasesChan

	var page tea.Model = organizations.New(ctx, gh, config)
	if config.Org != "" {
		page = repositories.New(ctx, gh, config)
	}

	if err := tea.NewProgram(page, tea.WithAltScreen(), tea.WithMouseAllMotion()).Start(); err != nil {
		return fmt.Errorf("failed to execute tui: %v", err)
	}

	if releases := <-releasesChan; releases != nil {
		for _, release := range releases {
			printRelease(release)
		}
	}

	return nil
}

var (
	titleStyle   = lipgloss.NewStyle().Foreground(colors.Title).Bold(true)
	urlStyle     = lipgloss.NewStyle().Foreground(colors.URL)
	versionStyle = lipgloss.NewStyle().Foreground(colors.Selected)
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
)

func printRelease(release *service.RepositoryReleaseResponse) {
	fullName := titleStyle.Render("## " + release.Owner + "/" + release.Name)
	version := versionStyle.Render(release.Version)

	fmt.Printf("%s %s\n", fullName, version)
	fmt.Print(release.Body)

	if release.IsError() {
		err := errStyle.Render("Error: " + release.Error.Error())
		fmt.Printf("%s\n\n", err)
		return
	}

	url := urlStyle.Render(release.URL)
	fmt.Printf("%s\n\n", url)
}
