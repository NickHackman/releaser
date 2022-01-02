package tui

import (
	"fmt"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/colors"
	"github.com/NickHackman/releaser/internal/tui/pages/organizations"
	"github.com/NickHackman/releaser/internal/tui/pages/repositories"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Execute(gh *github.Client, config *config.Config) error {
	releasesChan := make(chan []*github.RepositoryReleaseResponse, 1)
	config.Terminal.Releases = releasesChan

	var page tea.Model = organizations.New(gh, config)
	if config.Org != "" {
		page = repositories.New(gh, config)
	}

	if err := tea.NewProgram(page, tea.WithAltScreen(), tea.WithMouseAllMotion()).Start(); err != nil {
		return fmt.Errorf("failed to execute tui: %v", err)
	}

	select {
	case releases := <-releasesChan:
		if len(releases) > 0 {
			printReleases(releases)
		}
	default:
		fmt.Println("No releases were created.")
	}

	return nil
}

var (
	titleStyle   = lipgloss.NewStyle().Foreground(colors.Title).Bold(true)
	urlStyle     = lipgloss.NewStyle().Foreground(colors.URL)
	versionStyle = lipgloss.NewStyle().Foreground(colors.Selected)
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
)

func printReleases(releases []*github.RepositoryReleaseResponse) {
	for _, release := range releases {
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
}
