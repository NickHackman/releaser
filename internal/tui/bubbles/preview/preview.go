package preview

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	branchIcon     = ""
	loadingMessage = "Loading..."
)

var (
	previewTitle = titleStyle.Render("Preview")
)

type Model struct {
	viewport viewport.Model
	branch   string
	version  string
	Width    int
}

func New() Model {
	viewport := viewport.Model{}
	viewport.SetContent(loadingMessage)

	return Model{
		viewport: viewport,
	}
}

func (m *Model) SetContent(content, branch, version string) {
	m.viewport.SetContent(content)
	m.branch = branch
	m.version = version
}

func (m *Model) SetLoading() {
	m.viewport.SetContent(loadingMessage)
	m.branch = ""
	m.version = ""
}

func (m *Model) SetSize(width, height int) {
	m.Width = width

	m.viewport.Height = height - lipgloss.Height(previewTitle) - lipgloss.Height(m.statusBarView())
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// Only handle mouse events
	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) statusBarView() string {
	version := statusStyle.Render(m.version)
	branch := statusStyle.Render(branchIcon + " " + m.branch)

	sep := lipgloss.NewStyle().
		Width(m.Width - lipgloss.Width(version) - lipgloss.Width(branch)).
		Render("")

	bar := lipgloss.JoinHorizontal(lipgloss.Left, version, sep, branch)

	return statusBarStyle.Render(bar)
}

func (m Model) View() string {
	preview := contentStyle.Width(m.Width).MaxHeight(m.viewport.Height + 1).Render(m.viewport.View())
	body := style.Render(lipgloss.JoinVertical(lipgloss.Left, previewTitle, preview))

	return lipgloss.JoinVertical(lipgloss.Left, body, m.statusBarView())
}
