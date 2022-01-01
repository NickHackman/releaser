package repositories

import (
	"github.com/NickHackman/tagger/internal/tui/colors"
	"github.com/charmbracelet/lipgloss"
)

var (
	listTitleStyle = lipgloss.NewStyle().Padding(1).Background(colors.Title).Bold(true)
	listStyle      = lipgloss.NewStyle().PaddingRight(1).Width(listWidth).MaxWidth(listWidth)
)
