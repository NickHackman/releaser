package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/NickHackman/tagger/internal/edit"
	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/template"
	"github.com/NickHackman/tagger/internal/tui/bubbles/repository"
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/NickHackman/tagger/internal/tui/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minTerminalWidth         = 150
	increaseTerminalWidthMsg = "Increase width of terminal to display content."
	loadingMsg               = "Loading..."
	previewTitle             = "Preview"
)

type Model struct {
	list     list.Model
	progress progress.Model
	preview  viewport.Model
	keys     *keyMap

	gh      *service.GitHub
	channel <-chan *service.ReleaseableRepoResponse
	repos   int
	config  *config.Config
}

func New(gh *service.GitHub, config *config.Config) *Model {
	keys := newKeyMap()
	delegate := repository.NewDelegate()

	list := list.NewModel([]list.Item{}, delegate, 0, 0)
	list.Title = fmt.Sprintf("%s Repositories", strings.Title(config.Org))
	list.SetShowHelp(false)
	list.Styles.Title = listTitleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{delegate.Keys.Selection, keys.Template, delegate.Keys.Edit, keys.Publish}
	}
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{delegate.Keys.Selection, keys.Template, delegate.Keys.Edit, keys.Publish}
	}

	m := &Model{
		list:     list,
		progress: progress.NewModel(progress.WithoutPercentage(), progress.WithGradient(colors.ProgressStart, colors.ProgressEnd)),
		preview:  viewport.Model{},
		keys:     keys,
		gh:       gh,
		channel:  fetch(config, gh, config.Org),
		config:   config,
	}

	m.preview.SetContent(loadingMsg)
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

	m.preview.SetContent(current.Preview)
}

// handleEditTemplate opens the user's $EDITOR (or if none vim) and after they save/quit
// reads the template and applies the new template to every repository in the list.
func (m *Model) handleEditTemplate() []tea.Cmd {
	var cmds []tea.Cmd

	newTemplate, err := edit.Content(m.config.TemplateString, edit.TemplateInstructions)
	if err != nil {
		return append(cmds, m.list.NewStatusMessage(fmt.Sprintf("Error: %v", err)))
	}

	m.config.TemplateString = newTemplate

	for i, item := range m.list.Items() {
		current, ok := item.(repository.Item)
		if !ok {
			continue
		}

		newItem := repository.Item{R: current.R, Preview: template.Preview(current.R, newTemplate)}
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

	previewTitleHeight := lipgloss.Height(m.previewTitleView())
	m.preview.Height = height - statusHeight - previewTitleHeight - 1 // Subtract one for newline between previewTitle and preview
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
		percent := float64(m.repos) / float64(msg.R.Total)
		cmds = append(cmds, awaitCmd(m.channel, m.config.TemplateString), m.progress.SetPercent(percent), m.list.InsertItem(index, msg))

		// Loaded first repository, update the preview
		if m.repos == 1 {
			cmds = append(cmds, repository.RefreshPreview)
		}
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.More):
			// Force reset size
			m.SetSize(m.config.Width, m.config.Height)
		case key.Matches(msg, m.keys.Template):
			m.handleEditTemplate()
			cmds = append(cmds, repository.RefreshPreview)
		case key.Matches(msg, m.keys.Publish):
			if m.countSelected() == 0 {
				break
			}

			m.handlePublish()
			return m, tea.Quit
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

	// Only handle mouse events
	switch msg := msg.(type) {
	case tea.MouseMsg:
		m.preview, cmd = m.preview.Update(msg)
		cmds = append(cmds, cmd)
	}

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

		releases = append(releases, &service.RepositoryRelease{Name: i.R.Repo.GetName(), Version: "1.0.0", Body: i.Preview})
	}

	m.config.Releases <- m.gh.CreateReleases(ctx, m.config.Org, releases)
}

func (m Model) statusView() string {
	helpView := m.list.Styles.HelpStyle.Render(m.list.Help.View(m.list))
	return lipgloss.JoinVertical(lipgloss.Left, helpView, m.progress.View())
}

func (m Model) previewTitleView() string {
	return viewportTitleStyle.Render(previewTitle)
}

func (m Model) previewView() string {
	preview := lipgloss.NewStyle().MarginTop(1).Width(m.preview.Width).Render(m.preview.View())

	return viewportStyle.Render(lipgloss.JoinVertical(lipgloss.Left, m.previewTitleView(), preview))
}

func (m Model) View() string {
	if m.config.Width < minTerminalWidth {
		return increaseTerminalWidthMsg
	}

	top := lipgloss.JoinHorizontal(lipgloss.Left, listStyle.Render(m.list.View()), m.previewView())
	return lipgloss.JoinVertical(lipgloss.Left, top, m.statusView())
}

func fetch(config *config.Config, gh *service.GitHub, org string) <-chan *service.ReleaseableRepoResponse {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	channel, callback := gh.ReleasableReposByOrg(ctx, org)

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

		return repository.Item{R: r, Preview: template.Preview(r, templateString)}
	}
}

func max(a, b int) int {
	if a < b {
		return b
	}

	return a
}
