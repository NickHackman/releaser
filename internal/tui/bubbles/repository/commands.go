package repository

import tea "github.com/charmbracelet/bubbletea"

type RefreshPreviewCmd struct{}

func RefreshPreview() tea.Msg {
	return RefreshPreviewCmd{}
}
