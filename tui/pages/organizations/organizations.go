package organizations

import (
	"context"

	"github.com/NickHackman/tagger/service"
	"github.com/NickHackman/tagger/tui/bubbles/org"
	"github.com/NickHackman/tagger/tui/colors"
	"github.com/NickHackman/tagger/tui/pages/repositories"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	list     list.Model
	progress progress.Model
	keys     *organizationsListKeyMap

	gh      *service.GitHub
	ctx     context.Context
	channel <-chan *service.OrgResponse
	orgs    int
}

func New(ctx context.Context, gh *service.GitHub) *Model {
	listKeys := newOrganizationsListKeyMap()

	list := list.NewModel([]list.Item{}, org.Delegate{}, 0, 0)
	list.Title = "GitHub Organizations"
	list.Styles.Title = orgListTitleStyle
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.selectionFull,
			listKeys.refreshFull,
			listKeys.quit,
		}
	}
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{listKeys.selection, listKeys.refresh}
	}

	progress := progress.NewModel(progress.WithoutPercentage(), progress.WithGradient(colors.ProgressStart, colors.ProgressEnd))

	org := &Model{gh: gh, list: list, progress: progress, keys: listKeys, ctx: ctx}
	org.channel = fetch(ctx, gh)

	return org
}

func fetch(ctx context.Context, gh *service.GitHub) <-chan *service.OrgResponse {
	channel, callback := gh.Orgs(ctx)

	go func() {
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
		o.progress.Width = msg.Width

		_, h := lipgloss.Size(o.progress.View())
		o.list.SetSize(msg.Width, msg.Height-h)
	case orgInsertCmd:
		o.orgs++

		index := len(o.list.Items()) - 1
		item := org.New(msg.Org)

		percent := float64(o.orgs) / float64(msg.Total)
		cmds = append(cmds, awaitCmd(o.channel), o.progress.SetPercent(percent), o.list.InsertItem(index, item))
	case progress.FrameMsg:
		progressModel, cmd := o.progress.Update(msg)
		o.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		if o.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, o.keys.quit):
			return o, tea.Quit
		case key.Matches(msg, o.keys.selection):
			organization, ok := o.list.SelectedItem().(org.Item)
			if !ok {
				// TODO: something is horribly wrong
				return o, nil
			}

			repositories := repositories.New(o.ctx, o.gh, organization.GetLogin(), o.list.Width(), o.list.Height())
			return repositories, repositories.Init()
		case key.Matches(msg, o.keys.refresh):
			o.orgs = 0
			o.channel = fetch(o.ctx, o.gh)
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
		cmd, ok := <-channel
		if !ok {
			return nil
		}

		return orgInsertCmd(cmd)
	}
}
