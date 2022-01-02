package organizations

import (
	"bytes"
	"context"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/bubbles/organization"
	"github.com/NickHackman/releaser/internal/tui/colors"
	"github.com/NickHackman/releaser/internal/tui/pages/repositories"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
)

const (
	listTitle = "GitHub Organizations"
)

type Model struct {
	list     list.Model
	progress progress.Model
	keys     *keyMap

	gh      *github.Client
	channel <-chan *github.OrgResponse
	orgs    int

	config *config.Config
}

func New(gh *github.Client, config *config.Config) *Model {
	listKeys := newOrganizationsListKeyMap()

	list := list.NewModel([]list.Item{}, organization.Delegate{}, 0, 0)
	list.Title = listTitle
	list.Styles.Title = orgListTitleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{listKeys.Selection, listKeys.Refresh, listKeys.Open}
	}
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{listKeys.Selection, listKeys.Refresh, listKeys.Open}
	}

	progress := progress.NewModel(progress.WithoutPercentage(), progress.WithGradient(colors.ProgressStart, colors.ProgressEnd))

	return &Model{
		gh:       gh,
		list:     list,
		progress: progress,
		keys:     listKeys,
		channel:  fetch(config, gh),
		config:   config,
	}
}

func fetch(config *config.Config, gh *github.Client) <-chan *github.OrgResponse {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	channel, callback := gh.Orgs(ctx)

	go func() {
		defer cancel()
		if err := callback(); err != nil {
			// TODO: better handle error
			panic(err)
		}
	}()

	return channel
}

func (m Model) Init() tea.Cmd {
	return awaitCmd(m.channel)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.config.SetSize(msg.Width, msg.Height)
		m.progress.Width = msg.Width

		_, h := lipgloss.Size(m.progress.View())
		m.list.SetSize(msg.Width, msg.Height-h)
	case organization.Item:
		m.orgs++

		index := len(m.list.Items()) - 1
		percent := float64(m.orgs) / float64(msg.Total)
		cmds = append(cmds, awaitCmd(m.channel), m.progress.SetPercent(percent), m.list.InsertItem(index, msg))
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Open):
			organization, ok := m.list.SelectedItem().(organization.Item)
			if !ok {
				break
			}

			var output bytes.Buffer
			browser.Stdout = &output

			var statusMsg string
			if err := browser.OpenURL(organization.Org.GetHTMLURL()); err != nil {
				statusMsg = "Error: " + err.Error()
			} else {
				statusMsg = output.String()
			}

			cmds = append(cmds, m.list.NewStatusMessage(statusMsg))
		case key.Matches(msg, m.keys.Selection):
			organization, ok := m.list.SelectedItem().(organization.Item)
			if !ok {
				return m, nil
			}

			m.config.Org = organization.Org.GetLogin()

			repositories := repositories.New(m.gh, m.config)
			return repositories, repositories.Init()
		case key.Matches(msg, m.keys.Refresh):
			m.orgs = 0
			m.channel = fetch(m.config, m.gh)
			cmds = append(cmds, m.progress.SetPercent(0), m.list.SetItems([]list.Item{}), m.Init())
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), m.progress.View())
}

func awaitCmd(channel <-chan *github.OrgResponse) tea.Cmd {
	return func() tea.Msg {
		org, ok := <-channel
		if !ok {
			return nil
		}

		return organization.Item{OrgResponse: org}
	}
}
