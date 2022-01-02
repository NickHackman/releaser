package repositories

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/bubbles/repository"
	"github.com/NickHackman/releaser/internal/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"
)

type errorCmd error

func (m *Model) publishCmd() tea.Cmd {
	return func() tea.Msg {
		if m.countSelected() == 0 {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
		defer cancel()

		var releases []*github.RepositoryRelease
		for _, item := range m.list.Items() {
			i, ok := item.(repository.Item)
			if !ok || !i.Selected {
				continue
			}

			releases = append(releases, &github.RepositoryRelease{Name: i.Repo.GetName(), Version: i.Version, Body: i.Preview})
		}

		m.config.Terminal.Releases <- m.gh.CreateReleases(ctx, m.config.Org, releases)
		return tea.Quit
	}
}

func (m *Model) openURLCmd() tea.Cmd {
	return func() tea.Msg {
		item, ok := m.list.SelectedItem().(repository.Item)
		if !ok {
			return nil
		}

		var output bytes.Buffer
		browser.Stdout = &output

		if err := browser.OpenURL(item.Repo.GetHTMLURL()); err != nil {
			return errorCmd(err)
		}

		return nil
	}
}

func loadRepositoriesCmd(channel <-chan *github.ReleaseableRepoResponse, config *config.Config) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-channel
		if !ok {
			return nil
		}

		return repository.Item{
			ReleaseableRepoResponse: r,
			Preview:                 previewContent(r, config.Template),
			Branch:                  r.Branch,
			Version:                 version.New(r.LatestTag.GetName(), config.VersionChange),
		}
	}
}

type commitTemplate struct {
	Sha     string
	URL     string
	Summary string
	Message string

	AuthorUsername string
	AuthorName     string
	AuthorEmail    string
	AuthorDate     time.Time
	AuthorURL      string

	CommitterUsername string
	CommitterName     string
	CommitterEmail    string
	CommitterDate     time.Time
	CommitterURL      string
}

func previewContent(r *github.ReleaseableRepoResponse, templatedString string) string {
	var templateCommits []*commitTemplate

	for _, c := range r.Commits {
		templateCommits = append(templateCommits,
			&commitTemplate{
				Sha:     c.GetSHA(),
				URL:     c.GetHTMLURL(),
				Message: c.GetCommit().GetMessage(),
				Summary: strings.Split(c.GetCommit().GetMessage(), "\n")[0],

				AuthorUsername: c.GetAuthor().GetLogin(),
				AuthorURL:      c.GetAuthor().GetURL(),
				AuthorName:     c.GetCommit().GetAuthor().GetName(),
				AuthorEmail:    c.GetCommit().GetAuthor().GetEmail(),
				AuthorDate:     c.GetCommit().GetAuthor().GetDate(),

				CommitterUsername: c.GetCommitter().GetLogin(),
				CommitterURL:      c.GetCommitter().GetURL(),
				CommitterName:     c.GetCommit().GetCommitter().GetName(),
				CommitterEmail:    c.GetCommit().GetCommitter().GetEmail(),
				CommitterDate:     c.GetCommit().GetCommitter().GetDate(),
			})
	}

	params := map[string]interface{}{
		"RepositoryName":          r.Repo.GetName(),
		"RepositoryOwner":         r.Repo.GetOwner().GetLogin(),
		"RepositoryUrl":           r.Repo.GetHTMLURL(),
		"RepositoryDescription":   r.Repo.GetDescription(),
		"RepositoryDefaultBranch": r.Repo.GetDefaultBranch(),
		"Commits":                 templateCommits,
	}

	sf := sprig.TxtFuncMap()

	t, err := template.New("template").Funcs(sf).Parse(templatedString)
	if err != nil {
		return templatedString + fmt.Sprintf("\n\nError: failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, params); err != nil {
		return templatedString + fmt.Sprintf("\n\nError: failed to execute template: %v", err)
	}

	return buf.String()
}
