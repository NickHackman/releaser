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
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

const (
	minTerminalWidth         = 150
	increaseTerminalWidthMsg = "Increase width of terminal to display content."
)

type Model struct {
	list          list.Model
	progress      progress.Model
	templateView  viewport.Model
	help          help.Model
	keys          *keyMap
	width, height int

	ctx     context.Context
	gh      *service.GitHub
	channel <-chan *service.ReleaseableRepoResponse
	repos   int
}

func max(a, b int) int {
	if a < b {
		return b
	}

	return a
}

func New(ctx context.Context, gh *service.GitHub, org string, width, height int) *Model {
	listWidth := max(width, minTerminalWidth)
	listKeys := newKeyMap()

	list := list.NewModel([]list.Item{}, repository.Delegate{}, listWidth, height-2)
	list.Title = fmt.Sprintf("%s Repositories", strings.Title(org))
	list.SetShowHelp(false)
	list.Styles.Title = listTitleStyle

	help := help.NewModel()
	viewport := viewport.Model{Width: width - listWidth, Height: height - 5}
	viewport.SetContent("Loading...")

	progress := progress.NewModel(progress.WithoutPercentage(), progress.WithGradient(colors.ProgressStart, colors.ProgressEnd))
	progress.Width = width

	return &Model{
		list:         list,
		progress:     progress,
		templateView: viewport,
		width:        width,
		height:       height,
		help:         help,
		keys:         listKeys,
		ctx:          ctx,
		gh:           gh,
		channel:      fetch(ctx, gh, org),
	}
}

func (r Model) Init() tea.Cmd {
	return awaitCmd(r.channel)
}

func (r Model) ShortHelp() []key.Binding {
	switch r.list.FilterState() {
	case list.Filtering:
		return r.keys.ShortHelpFilter()
	case list.FilterApplied:
		return r.keys.ShortHelpFilterApplied()
	default:
		return r.keys.ShortHelp()
	}
}

func (r Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

func (r *Model) updatePreview() {
	currentItem := r.list.SelectedItem()
	current, ok := currentItem.(repository.Item)
	if !ok {
		return
	}

	r.templateView.SetContent(current.Preview)
}

func (r Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
		_, h := lipgloss.Size(r.renderStatus())
		r.list.SetSize(r.width, r.height-h)
		r.templateView.Height = r.height - h - 4
		r.progress.Width = r.width
		r.help.Width = r.width

	case progress.FrameMsg:
		progressModel, cmd := r.progress.Update(msg)
		r.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case repository.Item:
		r.repos++

		index := len(r.list.Items()) - 1
		percent := float64(r.repos) / float64(msg.R.Total)
		cmds = append(cmds, awaitCmd(r.channel), r.progress.SetPercent(percent), r.list.InsertItem(index, msg))

		// Loaded first repository, update the preview
		if r.repos == 1 {
			r.updatePreview()
		}
	case tea.KeyMsg:
		if r.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case r.keys.IsListBuiltin(msg):
			break
		case key.Matches(msg, r.keys.Quit):
			return r, tea.Quit
		case key.Matches(msg, r.keys.Edit):
			currentItem := r.list.SelectedItem()
			current, ok := currentItem.(repository.Item)
			if !ok {
				return r, nil
			}

			result, err := edit.Content(current.Preview, edit.BasicInstructions)
			if err != nil {
				cmds = append(cmds, r.list.NewStatusMessage("Error: "+err.Error()))
				break
			}

			currentIndex := r.list.Index()
			newItem := repository.Item{R: current.R, Preview: result}
			cmds = append(cmds, r.list.SetItem(currentIndex, newItem))
		case key.Matches(msg, r.keys.Template):
			templatedString := viper.GetString("template")
			newTemplate, err := edit.Content(templatedString, edit.TemplateInstructions)
			if err != nil {
				cmds = append(cmds, r.list.NewStatusMessage("Error: "+err.Error()))
				break
			}
			viper.Set("template", newTemplate)

			for i, item := range r.list.Items() {
				current, ok := item.(repository.Item)
				if !ok {
					return r, nil
				}

				newItem := repository.Item{R: current.R, Preview: constructPreview(current.R, newTemplate)}
				// SetItems doubles the amount of items in the List
				cmds = append(cmds, r.list.SetItem(i, newItem))
			}
		case key.Matches(msg, r.keys.SaveTemplate):
			return r, tea.Quit
		case key.Matches(msg, r.keys.Version):
			return r, tea.Quit
		}

		r.updatePreview()
	}

	currentIndex := r.list.Index()
	r.list, cmd = r.list.Update(msg)
	cmds = append(cmds, cmd)
	newIndex := r.list.Index()

	if currentIndex != newIndex {
		r.updatePreview()
	}

	r.templateView, cmd = r.templateView.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r Model) renderStatus() string {
	return lipgloss.JoinVertical(lipgloss.Left, r.help.View(r), r.progress.View())
}

func (r Model) View() string {
	if r.width < minTerminalWidth {
		return increaseTerminalWidthMsg
	}

	templateView := viewportStyle.Render(lipgloss.JoinVertical(lipgloss.Left, viewportTitleStyle.Render("Preview"), "\n"+lipgloss.NewStyle().Width(r.templateView.Width).Render(r.templateView.View())))
	top := lipgloss.JoinHorizontal(lipgloss.Left, listStyle.Render(r.list.View()), templateView)

	return lipgloss.JoinVertical(lipgloss.Left, top, r.renderStatus())
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
		content = templatedString + "\n\n# Error: " + err.Error()
	}

	return content
}

func awaitCmd(channel <-chan *service.ReleaseableRepoResponse) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-channel
		if !ok {
			return nil
		}

		templatedString := viper.GetString("template")
		return repository.Item{R: r, Preview: constructPreview(r, templatedString)}
	}
}
