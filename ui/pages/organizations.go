package pages

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/NickHackman/tagger/service"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v41/github"
)

var (
	orgListTitleStyle       = lipgloss.NewStyle().Padding(1).Background(lipgloss.Color("#00316E")).Bold(true)
	orgItemTitleStyle       = lipgloss.NewStyle().Bold(true)
	orgItemDescriptionStyle = lipgloss.NewStyle().Faint(true)
	orgItemUrlStyle         = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#6697d6"))
	orgItemSelectedStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("#08589d"))
	orgItemUnselectedStyle  = lipgloss.NewStyle().PaddingLeft(1)
)

type orgItem struct {
	*github.Organization
}

func (oi orgItem) FilterValue() string {
	return oi.GetLogin()
}

type orgDelegate struct{}

func (d orgDelegate) Height() int {
	return 2
}

func (d orgDelegate) Spacing() int {
	return 2
}

func (d orgDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d orgDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	oi, ok := listItem.(orgItem)
	if !ok {
		return
	}

	var output strings.Builder
	output.WriteString(orgItemTitleStyle.Render(oi.GetLogin()))
	if oi.GetIsVerified() {
		output.WriteString(" ✅")
	}

	if oi.GetDescription() != "" {
		output.WriteString("\n" + orgItemDescriptionStyle.Render(oi.GetDescription()))
	}

	if oi.GetHTMLURL() != "" {
		output.WriteString("\n" + orgItemUrlStyle.Render(oi.GetHTMLURL()))
	}

	render := orgItemUnselectedStyle.Render
	if index == m.Index() {
		render = orgItemSelectedStyle.Render
	}

	fmt.Fprint(w, render(output.String()))
}

type organizationsListKeyMap struct {
	// Short
	selection key.Binding
	refresh   key.Binding

	// Full
	selectionFull key.Binding
	refreshFull   key.Binding
	quit          key.Binding
}

func newOrganizationsListKeyMap() *organizationsListKeyMap {
	return &organizationsListKeyMap{
		// Short
		selection: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),

		// Full
		selectionFull: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select GitHub organization"),
		),
		refreshFull: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh GitHub organizations"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
		),
	}
}

type orgInsertCmd *service.OrgResponse

type Organizations struct {
	list     list.Model
	progress progress.Model
	keys     *organizationsListKeyMap

	gh      *service.GitHub
	ctx     context.Context
	channel <-chan *service.OrgResponse
	orgs    int
}

func NewOrganizations(ctx context.Context, gh *service.GitHub) *Organizations {
	listKeys := newOrganizationsListKeyMap()

	list := list.NewModel([]list.Item{}, orgDelegate{}, 0, 0)
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

	progress := progress.NewModel(progress.WithoutPercentage(), progress.WithGradient("#00316E", "#6697d6"))

	org := &Organizations{gh: gh, list: list, progress: progress, keys: listKeys, ctx: ctx}
	org.channel = fetchOrgs(ctx, gh)

	return org
}

func fetchOrgs(ctx context.Context, gh *service.GitHub) <-chan *service.OrgResponse {
	channel, callback := gh.Orgs(ctx)

	go func() {
		if err := callback(); err != nil {
			panic(err)
		}
	}()

	return channel
}

func (o Organizations) Init() tea.Cmd {
	return awaitCmd(o.channel)
}

func (o Organizations) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		item := orgItem{msg.Org}

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
			return o, tea.Quit
		case key.Matches(msg, o.keys.refresh):
			o.orgs = 0
			o.channel = fetchOrgs(o.ctx, o.gh)
			cmds = append(cmds, o.progress.SetPercent(0), o.list.SetItems([]list.Item{}), o.Init())
		}
	}

	o.list, cmd = o.list.Update(msg)
	cmds = append(cmds, cmd)

	return o, tea.Batch(cmds...)
}

func (o Organizations) View() string {
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
