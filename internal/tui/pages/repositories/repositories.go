package repositories

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/NickHackman/tagger/internal/edit"
	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/template"
	"github.com/NickHackman/tagger/internal/tui/bubbles/preview"
	"github.com/NickHackman/tagger/internal/tui/bubbles/repository"
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/NickHackman/tagger/internal/tui/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"gopkg.in/yaml.v3"
)

const (
	minTerminalWidth         = 150
	increaseTerminalWidthMsg = "Increase width of terminal to display content."
)

type Model struct {
	list     list.Model
	progress progress.Model
	preview  preview.Model
	keys     *keyMap

	gh      *service.GitHub
	channel <-chan *service.ReleaseableRepoResponse
	repos   int
	config  *config.Config
}

func New(gh *service.GitHub, config *config.Config) *Model {
	keys := newKeyMap()
	delegate := repository.NewDelegate(gh, config)

	list := list.NewModel([]list.Item{}, delegate, 0, 0)
	list.Title = fmt.Sprintf("%s Repositories", strings.Title(config.Org))
	list.SetShowHelp(false)
	list.Styles.Title = listTitleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			delegate.Keys.Selection,
			keys.Template,
			keys.Open,
			delegate.Keys.Edit,
			keys.Publish,
			keys.Refresh,
		}
	}

	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			delegate.Keys.Selection,
			keys.Template,
			keys.Open,
			delegate.Keys.Edit,
			keys.Publish,
			keys.Refresh,
		}
	}

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
	return awaitCmd(m.channel, m.config.TemplateString)
}

func (m *Model) updatePreview() {
	currentItem := m.list.SelectedItem()
	current, ok := currentItem.(repository.Item)
	if !ok {
		return
	}

	m.preview.SetContent(current.Preview, current.Branch)
}

// handleEditTemplate opens the user's $EDITOR (or if none vim) and after they save/quit
// reads the template and applies the new template to every repository in the list.
func (m *Model) handleEditTemplate() []tea.Cmd {
	var cmds []tea.Cmd

	yamlRepresentation := map[string]interface{}{
		"template": &yaml.Node{
			Value:       m.config.TemplateString,
			Kind:        yaml.ScalarNode,
			HeadComment: m.config.TemplateInstructions,
		},
	}

	var result *struct {
		Template string
	}

	if err := edit.Invoke(&yamlRepresentation, &result); err != nil {
		return append(cmds, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", err)))
	}

	if m.config.TemplateString == result.Template {
		return cmds
	}

	m.config.TemplateString = result.Template

	for i, item := range m.list.Items() {
		current, ok := item.(repository.Item)
		if !ok {
			continue
		}

		newItem := repository.Item{
			ReleaseableRepoResponse: current.ReleaseableRepoResponse,
			Preview:                 template.Preview(current.ReleaseableRepoResponse, result.Template),
			Branch:                  current.Branch,
		}

		// SetItems doubles the amount of items in the List
		cmds = append(cmds, m.list.SetItem(i, newItem))
	}

	return cmds
}

func (m *Model) SetSize(width, height int) {
	m.config.Width, m.config.Height = width, height

	// Status bars take up full width of screen
	m.progress.Width = width
	m.list.Help.Width = width

	statusHeight := lipgloss.Height(m.statusView())
	listWidth := max(width, minTerminalWidth)
	m.list.SetSize(listWidth, height-statusHeight-1)
	m.preview.SetSize(width, height-statusHeight-1)
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := m.updateSubmodels(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case repository.RefreshPreviewCmd:
		m.updatePreview()
	case repository.Item:
		m.repos++

		index := len(m.list.Items()) - 1
		percent := float64(m.repos) / float64(msg.Total)
		cmds = append(cmds,
			awaitCmd(m.channel, m.config.TemplateString),
			m.progress.SetPercent(percent),
			m.list.InsertItem(index, msg),
			// Refresh preview every time, since the current item may change
			repository.RefreshPreview,
		)

		// TODO: when opening the editor to edit the preview or the template in some cases where the user enters input into the terminal
		// Bubbletea will crash with an error of "Could not decode rune", which seems to be as a result of multiple runes being entered
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.More):
			// Force reset size
			m.SetSize(m.config.Width, m.config.Height)
		case key.Matches(msg, m.keys.Open):
			item, ok := m.list.SelectedItem().(repository.Item)
			if !ok {
				break
			}

			var output bytes.Buffer
			browser.Stdout = &output

			var statusMsg string
			if err := browser.OpenURL(item.Repo.GetHTMLURL()); err != nil {
				statusMsg = "Error: " + err.Error()
			} else {
				statusMsg = output.String()
			}

			cmds = append(cmds, m.list.NewStatusMessage(statusMsg))
		case key.Matches(msg, m.keys.Template):
			m.handleEditTemplate()
			cmds = append(cmds, repository.RefreshPreview)
		case key.Matches(msg, m.keys.Publish):
			if m.countSelected() == 0 {
				break
			}

			m.handlePublish()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			m.repos = 0
			m.channel = fetch(m.config, m.gh, m.config.Org)
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
		m.updatePreview()
	}

	var previewModel tea.Model
	previewModel, cmd = m.preview.Update(msg)
	m.preview = previewModel.(preview.Model)

	cmds = append(cmds, cmd)

	return cmds
}

func (m *Model) handlePublish() {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	var releases []*service.RepositoryRelease
	for _, item := range m.list.Items() {
		i, ok := item.(repository.Item)
		if !ok || !i.Selected {
			continue
		}

		releases = append(releases, &service.RepositoryRelease{Name: i.Repo.GetName(), Version: "1.0.0", Body: i.Preview})
	}

	m.config.Releases <- m.gh.CreateReleases(ctx, m.config.Org, releases)
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

func fetch(config *config.Config, gh *service.GitHub, org string) <-chan *service.ReleaseableRepoResponse {
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

func awaitCmd(channel <-chan *service.ReleaseableRepoResponse, templateString string) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-channel
		if !ok {
			return nil
		}

		return repository.Item{
			ReleaseableRepoResponse: r,
			Preview:                 template.Preview(r, templateString),
			Branch:                  r.Branch,
		}
	}
}

func max(a, b int) int {
	if a < b {
		return b
	}

	return a
}
