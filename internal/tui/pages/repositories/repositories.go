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

	ctx     context.Context
	gh      *service.GitHub
	channel <-chan *service.ReleaseableRepoResponse
	repos   int
	config  *config.Config
}

func New(ctx context.Context, gh *service.GitHub, config *config.Config) *Model {
	keys := newKeyMap()
	keys.Publish.SetEnabled(false)

	list := list.NewModel([]list.Item{}, repository.Delegate{}, 0, 0)
	list.Title = fmt.Sprintf("%s Repositories", strings.Title(config.Org))
	list.SetShowHelp(false)
	list.Styles.Title = listTitleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.Selection,
			keys.Publish,
			keys.Edit,
			keys.Template,
		}
	}

	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Selection, keys.Template, keys.Edit, keys.Publish}
	}

	m := &Model{
		list:     list,
		progress: progress.NewModel(progress.WithoutPercentage(), progress.WithGradient(colors.ProgressStart, colors.ProgressEnd)),
		preview:  viewport.Model{},
		keys:     keys,
		ctx:      ctx,
		gh:       gh,
		channel:  fetch(ctx, gh, config.Org),
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
			return cmds
		}

		newItem := repository.Item{R: current.R, Preview: constructPreview(current.R, newTemplate)}
		// SetItems doubles the amount of items in the List
		cmds = append(cmds, r.list.SetItem(i, newItem))
	}

	return cmds
}

// handleEditPreview opens the user's $EDITOR (or if none vim) and after they save/quit
// reads the new preview and applies it.
func (r *Model) handleEditPreview() []tea.Cmd {
	var cmds []tea.Cmd

	currentItem := r.list.SelectedItem()
	current, ok := currentItem.(repository.Item)
	if !ok {
		return cmds
	}

	result, err := edit.Content(current.Preview, edit.ManualEditInstructions)
	if err != nil {
		return append(cmds, r.list.NewStatusMessage(fmt.Sprintf("Error: %v", err)))
	}

	currentIndex := r.list.Index()
	newItem := repository.Item{R: current.R, Preview: result}
	return append(cmds, r.list.SetItem(currentIndex, newItem))
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
	case repository.Item:
		r.repos++

		index := len(r.list.Items()) - 1
		percent := float64(r.repos) / float64(msg.R.Total)
		cmds = append(cmds, awaitCmd(r.channel, r.config.TemplateString), r.progress.SetPercent(percent), r.list.InsertItem(index, msg))

		// Loaded first repository, update the preview
		if r.repos == 1 {
			r.updatePreview()
		}
	case tea.KeyMsg:
		if r.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, r.keys.Quit):
			return r, tea.Quit
		case key.Matches(msg, r.keys.More):
			// Force reset size
			r.SetSize(r.config.Width, r.config.Height)
		case key.Matches(msg, r.keys.Edit):
			r.handleEditPreview()
			r.updatePreview()
		case key.Matches(msg, r.keys.Template):
			r.handleEditTemplate()
			r.updatePreview()
		case key.Matches(msg, r.keys.Selection):
			item := r.list.SelectedItem()

			current, ok := item.(repository.Item)
			if !ok {
				return r, nil
			}

			index := r.list.Index()
			cmds = append(cmds, r.list.SetItem(index, repository.Item{R: current.R, Preview: current.Preview, Selected: !current.Selected}))

			switch r.countSelected() {
			case 0:
				r.keys.Publish.SetEnabled(false)
			default:
				r.keys.Publish.SetEnabled(true)
			}
		case key.Matches(msg, r.keys.Publish):
			if r.countSelected() == 0 {
				break
			}

			r.handlePublish()
			return r, tea.Quit
		}
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

	// Only set preview if the current index has changed
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
	var releases []*service.RepositoryRelease
	for _, item := range r.list.Items() {
		i, ok := item.(repository.Item)
		if !ok || !i.Selected {
			continue
		}

		releases = append(releases, &service.RepositoryRelease{Name: i.R.Repo.GetName(), Version: "1.0.0", Body: i.Preview})
	}

	r.config.Releases <- r.gh.CreateReleases(r.ctx, r.config.Org, releases)
}

func (r Model) statusView() string {
	helpView := r.list.Styles.HelpStyle.Render(r.list.Help.View(r.list))
	return lipgloss.JoinVertical(lipgloss.Left, helpView, r.progress.View())
}

func (r Model) previewTitleView() string {
	return viewportTitleStyle.Render(previewTitle)
}

func (r Model) previewView() string {
	preview := lipgloss.NewStyle().Width(r.preview.Width).Render(r.preview.View())

	// HACK: add newline before preview to ensure it's on a line below.
	return viewportStyle.Render(lipgloss.JoinVertical(lipgloss.Left, r.previewTitleView(), "\n"+preview))
}

func (r Model) View() string {
	if r.config.Width < minTerminalWidth {
		return increaseTerminalWidthMsg
	}

	top := lipgloss.JoinHorizontal(lipgloss.Left, listStyle.Render(r.list.View()), r.previewView())
	return lipgloss.JoinVertical(lipgloss.Left, top, r.statusView())
}

func fetch(ctx context.Context, gh *service.GitHub, org string) <-chan *service.ReleaseableRepoResponse {
	channel, callback := gh.ReleasableReposByOrg(ctx, org)

	go func() {
		if err := callback(); err != nil {
			// TODO: better handle error
			panic(err)
		}
	}()

	return channel
}

func constructPreview(r *service.ReleaseableRepoResponse, templatedString string) string {
	tagTemplate := template.NewTag(r.Repo, r.Commits)
	content, err := tagTemplate.Execute(templatedString)
	if err != nil {
		content = fmt.Sprintf("%s\n\n# Error: %v", templatedString, err)
	}

	return content
}

func awaitCmd(channel <-chan *service.ReleaseableRepoResponse, templateString string) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-channel
		if !ok {
			return nil
		}

		return repository.Item{R: r, Preview: constructPreview(r, templateString)}
	}
}

func max(a, b int) int {
	if a < b {
		return b
	}

	return a
}
