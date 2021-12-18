package repository

import (
	"github.com/NickHackman/tagger/tui/colors"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle       = lipgloss.NewStyle().Bold(true)
	descriptionStyle = lipgloss.NewStyle().Faint(true).MaxWidth(75)
	urlStyle         = lipgloss.NewStyle().Underline(true).Foreground(colors.Url)
	selectedStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(colors.Selected)
	unselectedStyle  = lipgloss.NewStyle().PaddingLeft(1)
)
