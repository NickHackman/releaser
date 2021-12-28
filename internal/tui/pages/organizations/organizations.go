package organizations

import (
	"context"

	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/tui/bubbles/organization"
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/NickHackman/tagger/internal/tui/config"
	"github.com/NickHackman/tagger/internal/tui/pages/repositories"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listTitle = "GitHub Organizations"
)

type Model struct {
	list     list.Model
	progress progress.Model
	keys     *organizationsListKeyMap

	gh      *service.GitHub
	channel <-chan *service.OrgResponse
	orgs    int

	config *config.Config
}

func New(gh *service.GitHub, config *config.Config) *Model {
	listKeys := newOrganizationsListKeyMap()

	list := list.NewModel([]list.Item{}, organization.Delegate{}, 0, 0)
	list.Title = listTitle
	list.Styles.Title = orgListTitleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.selectionFull,
			listKeys.refreshFull,
		}
	}
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{listKeys.selection, listKeys.refresh}
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

func fetch(config *config.Config, gh *service.GitHub) <-chan *service.OrgResponse {
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

func (o Model) Init() tea.Cmd {
	return awaitCmd(o.channel)
}

func (o Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.config.Width, o.config.Height = msg.Width, msg.Height
		o.progress.Width = msg.Width

		_, h := lipgloss.Size(o.progress.View())
		o.list.SetSize(msg.Width, msg.Height-h)
	case organization.Item:
		o.orgs++

		index := len(o.list.Items()) - 1
		percent := float64(o.orgs) / float64(msg.R.Total)
		cmds = append(cmds, awaitCmd(o.channel), o.progress.SetPercent(percent), o.list.InsertItem(index, msg))
	case progress.FrameMsg:
		progressModel, cmd := o.progress.Update(msg)
		o.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		if o.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, o.keys.selection):
			organization, ok := o.list.SelectedItem().(organization.Item)
			if !ok {
				return o, nil
			}

			o.config.Org = organization.R.Org.GetLogin()

			repositories := repositories.New(o.gh, o.config)
			return repositories, repositories.Init()
		case key.Matches(msg, o.keys.refresh):
			o.orgs = 0
			o.channel = fetch(o.config, o.gh)
			cmds = append(cmds, o.progress.SetPercent(0), o.list.SetItems([]list.Item{}), o.Init())
		}
	}

	o.list, cmd = o.list.Update(msg)
	cmds = append(cmds, cmd)

	return o, tea.Batch(cmds...)
}

func (o Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, o.list.View(), o.progress.View())
}

func awaitCmd(channel <-chan *service.OrgResponse) tea.Cmd {
	return func() tea.Msg {
		org, ok := <-channel
		if !ok {
			return nil
		}

		return organization.Item{R: org}
	}
}
