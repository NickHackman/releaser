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
	Width    int
}

func New() Model {
	viewport := viewport.Model{}
	viewport.SetContent(loadingMessage)

	return Model{
		viewport: viewport,
	}
}

func (m *Model) SetContent(content, branch string) {
	m.viewport.SetContent(content)
	m.branch = branch
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
	return statusBarStyle.Render(branchIcon + " " + m.branch)
}

func (m Model) View() string {
	preview := contentStyle.Width(m.Width).Render(m.viewport.View())
	body := style.Render(lipgloss.JoinVertical(lipgloss.Left, previewTitle, preview))

	return lipgloss.JoinVertical(lipgloss.Left, body, m.statusBarView())
}
