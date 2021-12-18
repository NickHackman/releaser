package repositories

import (
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/charmbracelet/lipgloss"
)

var (
	listTitleStyle     = lipgloss.NewStyle().Padding(1).Background(colors.Title).Bold(true)
	viewportTitleStyle = lipgloss.NewStyle().Padding(1).MarginLeft(1).Background(colors.Title).Bold(true)

	viewportStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, false, false, true).PaddingLeft(1).BorderForeground(colors.Selected)
	listStyle     = lipgloss.NewStyle().PaddingRight(1).Width(minTerminalWidth / 2).MaxWidth(minTerminalWidth / 2)
)
