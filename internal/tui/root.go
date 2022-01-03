package tui

import (
	"context"
	"fmt"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/colors"
	"github.com/NickHackman/releaser/internal/tui/pages/organizations"
	"github.com/NickHackman/releaser/internal/tui/pages/repositories"
	"github.com/NickHackman/releaser/internal/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Execute(gh *github.Client, config *config.Config) error {
	// Release without interactive UI
	if len(config.Repositories) != 0 {
		return noninteractiveReleases(gh, config)
	}

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

func noninteractiveReleases(gh *github.Client, config *config.Config) error {
	var channel <-chan *github.ReleaseableRepoResponse
	var callback func() error

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

	var owner string
	if config.Org == "" {
		owner = config.Username
		channel, callback = gh.ReleaseableReposByUser(ctx, config.Username, config.Branch)
	} else {
		owner = config.Org
		channel, callback = gh.ReleasableReposByOrg(ctx, config.Org, config.Branch)
	}

	go func() {
		defer cancel()
		err := callback()
		if err != nil {
			// TODO: better handle error
			panic(err)
		}
	}()

	var releases []*github.RepositoryRelease
	for repo := range channel {
		name := repo.Repo.GetName()
		if !config.IsRepositoryToRelease(name) {
			continue
		}

		newVersion := version.New(repo.LatestTag.GetName(), config.VersionChange)
		description := repositories.PreviewContent(repo, config.Template)

		releases = append(releases, &github.RepositoryRelease{
			Name:    name,
			Version: newVersion,
			Body:    description,
		})
	}

	ctx, cancel = context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	response := gh.CreateReleases(ctx, owner, releases)

	printReleases(response)
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
