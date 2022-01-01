package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/bubbles/preview"
	"github.com/NickHackman/releaser/internal/tui/bubbles/repository"
	"github.com/NickHackman/releaser/internal/tui/colors"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minTerminalWidth         = 150
	listWidth                = 75
	increaseTerminalWidthMsg = "Increase width of terminal to display content."
)

type Model struct {
	list     list.Model
	progress progress.Model
	preview  preview.Model
	keys     *keyMap

	gh      *github.Client
	channel <-chan *github.ReleaseableRepoResponse
	repos   int
	config  *config.Config
}

func New(gh *github.Client, config *config.Config) *Model {
	keys := newKeyMap()
	delegate := repository.NewDelegate(gh, config)

	list := list.NewModel([]list.Item{}, delegate, 0, 0)
	list.Title = fmt.Sprintf("%s Repositories", strings.Title(config.Org))
	list.SetShowHelp(false)
	list.Styles.Title = listTitleStyle

	helpKeys := func() []key.Binding {
		return []key.Binding{
			delegate.Keys.Selection,
			keys.Open,
			keys.Publish,
			keys.RefreshConfig,
			keys.RefreshRepos,
		}
	}

	list.AdditionalFullHelpKeys = helpKeys
	list.AdditionalShortHelpKeys = helpKeys

	m := &Model{
		list:     list,
		progress: progress.NewModel(progress.WithoutPercentage(), progress.WithGradient(colors.ProgressStart, colors.ProgressEnd)),
		preview:  preview.New(),
		keys:     keys,
		gh:       gh,
		channel:  fetch(config, gh, config.Org),
		config:   config,
	}

	m.SetSize(config.Width, config.Height)
	return m
}

func (m Model) Init() tea.Cmd {
	return loadRepositoriesCmd(m.channel, m.config)
}

func (m *Model) SetSize(width, height int) {
	m.config.Width, m.config.Height = width, height

	statusHeight := lipgloss.Height(m.statusView())

	m.list.SetSize(width, height-statusHeight-1)

	// Status bars take up full width of screen
	m.progress.Width = width
	m.list.Help.Width = width

	m.preview.SetSize(width-listWidth, height-statusHeight-1)
}

func (m Model) countSelected() int {
	var selected int

	items := m.list.Items()
	for _, item := range items {
		current, ok := item.(repository.Item)
		if !ok {
			continue
		}

		if !current.Selected {
			continue
		}

		selected += 1
	}

	return selected
}

func (m *Model) refreshPreview() {
	currentItem := m.list.SelectedItem()
	current, ok := currentItem.(repository.Item)
	if !ok {
		return
	}

	m.preview.SetContent(current.Preview, current.Branch, current.Version)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := m.updateSubmodels(msg)

	switch msg := msg.(type) {
	case errorCmd:
		cmds = append(cmds, m.list.NewStatusMessage(msg.Error()))
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case repository.Item:
		m.repos++

		index := len(m.list.Items()) - 1
		percent := float64(m.repos) / float64(msg.Total)
		cmds = append(cmds,
			loadRepositoriesCmd(m.channel, m.config),
			m.progress.SetPercent(percent),
			m.list.InsertItem(index, msg),
		)

		// Refresh preview every time, since the current item may change
		m.refreshPreview()
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.More):
			// Force reset size
			m.SetSize(m.config.Width, m.config.Height)
		case key.Matches(msg, m.keys.Open):
			cmds = append(cmds, m.openURLCmd())
		case key.Matches(msg, m.keys.Publish):
			cmds = append(cmds, m.publishCmd())
		case key.Matches(msg, m.keys.RefreshConfig):
			m.config.Refresh()
			cmds = append(cmds, m.list.NewStatusMessage("Refreshing config..."))
			fallthrough // refresh all
		case key.Matches(msg, m.keys.RefreshRepos):
			m.repos = 0
			m.channel = fetch(m.config, m.gh, m.config.Org)
			m.preview.SetLoading()
			cmds = append(cmds, m.progress.SetPercent(0), m.list.SetItems([]list.Item{}), m.Init())
		}
	}

	switch m.countSelected() {
	case 0:
		m.keys.Publish.SetEnabled(false)
	default:
		m.keys.Publish.SetEnabled(true)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateSubmodels(msg tea.Msg) []tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	currentIndex := m.list.Index()

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	newIndex := m.list.Index()

	// Only update preview if the current index has changed
	if currentIndex != newIndex {
		m.refreshPreview()
	}

	var previewModel tea.Model
	previewModel, cmd = m.preview.Update(msg)
	m.preview = previewModel.(preview.Model)

	cmds = append(cmds, cmd)

	return cmds
}

func (m Model) statusView() string {
	helpView := m.list.Styles.HelpStyle.Render(m.list.Help.View(m.list))
	return lipgloss.JoinVertical(lipgloss.Left, helpView, m.progress.View())
}

func (m Model) View() string {
	if m.config.Width < minTerminalWidth {
		return increaseTerminalWidthMsg
	}

	top := lipgloss.JoinHorizontal(lipgloss.Left, listStyle.Render(m.list.View()), m.preview.View())
	return lipgloss.JoinVertical(lipgloss.Left, top, m.statusView())
}

func fetch(config *config.Config, gh *github.Client, org string) <-chan *github.ReleaseableRepoResponse {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	channel, callback := gh.ReleasableReposByOrg(ctx, org, config.Branch)

	go func() {
		defer cancel()
		if err := callback(); err != nil {
			// TODO: better handle error
			panic(err)
		}
	}()

	return channel
}
