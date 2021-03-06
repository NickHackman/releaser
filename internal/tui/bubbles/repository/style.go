package repository

import (
	"github.com/NickHackman/releaser/internal/tui/colors"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle       = lipgloss.NewStyle().Bold(true)
	descriptionStyle = lipgloss.NewStyle().Faint(true).MaxWidth(75)
	urlStyle         = lipgloss.NewStyle().Underline(true).Foreground(colors.URL)
	selectedStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(colors.Selected)
	unselectedStyle  = lipgloss.NewStyle().PaddingLeft(1)
	checkmarkStyle   = lipgloss.NewStyle().Foreground(colors.Title).Bold(true)
)
