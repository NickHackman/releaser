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

func (r Model) Init() tea.Cmd {
	return awaitCmd(r.channel, r.config.TemplateString)
}

func (r *Model) updatePreview() {
	currentItem := r.list.SelectedItem()
	current, ok := currentItem.(repository.Item)
	if !ok {
		return
	}

	r.preview.SetContent(current.Preview)
}

// handleEditTemplate opens the user's $EDITOR (or if none vim) and after they save/quit
// reads the template and applies the new template to every repository in the list.
func (r *Model) handleEditTemplate() []tea.Cmd {
	var cmds []tea.Cmd

	newTemplate, err := edit.Content(r.config.TemplateString, edit.TemplateInstructions)
	if err != nil {
		return append(cmds, r.list.NewStatusMessage(fmt.Sprintf("Error: %v", err)))
	}

	r.config.TemplateString = newTemplate

	for i, item := range r.list.Items() {
		current, ok := item.(repository.Item)
		if !ok {
			continue
		}

		newItem := repository.Item{R: current.R, Preview: template.Preview(current.R, newTemplate)}
		// SetItems doubles the amount of items in the List
		cmds = append(cmds, r.list.SetItem(i, newItem))
	}

	return cmds
}

func (r *Model) SetSize(width, height int) {
	r.config.Width, r.config.Height = width, height

	// Status bars take up full width of screen
	r.progress.Width = width
	r.list.Help.Width = width

	statusHeight := lipgloss.Height(r.statusView())
	listWidth := max(width, minTerminalWidth)
	r.list.SetSize(listWidth, height-statusHeight-1)

	previewTitleHeight := lipgloss.Height(r.previewTitleView())
	r.preview.Height = height - statusHeight - previewTitleHeight - 1 // Subtract one for newline between previewTitle and preview
}

func (r Model) countSelected() int {
	var selected int

	items := r.list.Items()
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

func (r Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := r.updateSubmodels(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.SetSize(msg.Width, msg.Height)
	case progress.FrameMsg:
		progressModel, cmd := r.progress.Update(msg)
		r.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case repository.RefreshPreviewCmd:
		r.updatePreview()
	case repository.Item:
		r.repos++

		index := len(r.list.Items()) - 1
		percent := float64(r.repos) / float64(msg.R.Total)
		cmds = append(cmds, awaitCmd(r.channel, r.config.TemplateString), r.progress.SetPercent(percent), r.list.InsertItem(index, msg))

		// Loaded first repository, update the preview
		if r.repos == 1 {
			cmds = append(cmds, repository.RefreshPreview)
		}
	case tea.KeyMsg:
		if r.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, r.keys.More):
			// Force reset size
			r.SetSize(r.config.Width, r.config.Height)
		case key.Matches(msg, r.keys.Template):
			r.handleEditTemplate()
			cmds = append(cmds, repository.RefreshPreview)
		case key.Matches(msg, r.keys.Publish):
			if r.countSelected() == 0 {
				break
			}

			r.handlePublish()
			return r, tea.Quit
		}
	}

	switch r.countSelected() {
	case 0:
		r.keys.Publish.SetEnabled(false)
	default:
		r.keys.Publish.SetEnabled(true)
	}

	return r, tea.Batch(cmds...)
}

func (r *Model) updateSubmodels(msg tea.Msg) []tea.Cmd {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	currentIndex := r.list.Index()

	r.list, cmd = r.list.Update(msg)
	cmds = append(cmds, cmd)

	newIndex := r.list.Index()

	// Only update preview if the current index has changed
	if currentIndex != newIndex {
		r.updatePreview()
	}

	// Only handle mouse events
	switch msg := msg.(type) {
	case tea.MouseMsg:
		r.preview, cmd = r.preview.Update(msg)
		cmds = append(cmds, cmd)
	}

	return cmds
}

func (r *Model) handlePublish() {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	var releases []*service.RepositoryRelease
	for _, item := range r.list.Items() {
		i, ok := item.(repository.Item)
		if !ok || !i.Selected {
			continue
		}

		releases = append(releases, &service.RepositoryRelease{Name: i.R.Repo.GetName(), Version: "1.0.0", Body: i.Preview})
	}

	r.config.Releases <- r.gh.CreateReleases(ctx, r.config.Org, releases)
}

func (r Model) statusView() string {
	helpView := r.list.Styles.HelpStyle.Render(r.list.Help.View(r.list))
	return lipgloss.JoinVertical(lipgloss.Left, helpView, r.progress.View())
}

func (r Model) previewTitleView() string {
	return viewportTitleStyle.Render(previewTitle)
}

func (r Model) previewView() string {
	preview := lipgloss.NewStyle().MarginTop(1).Width(r.preview.Width).Render(r.preview.View())

	return viewportStyle.Render(lipgloss.JoinVertical(lipgloss.Left, r.previewTitleView(), preview))
}

func (r Model) View() string {
	if r.config.Width < minTerminalWidth {
		return increaseTerminalWidthMsg
	}

	top := lipgloss.JoinHorizontal(lipgloss.Left, listStyle.Render(r.list.View()), r.previewView())
	return lipgloss.JoinVertical(lipgloss.Left, top, r.statusView())
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
