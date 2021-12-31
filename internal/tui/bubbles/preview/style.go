package preview

import (
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle     = lipgloss.NewStyle().Padding(1).MarginLeft(1).Background(colors.Title).Bold(true)
	statusBarStyle = lipgloss.NewStyle().Background(colors.Selected).MarginTop(1).Padding(0, 1).Bold(true)
	contentStyle   = lipgloss.NewStyle().MarginTop(1)
	style          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, false, false, true).PaddingLeft(1).BorderForeground(colors.Selected)
)