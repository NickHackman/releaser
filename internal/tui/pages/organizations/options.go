package organizations

import "github.com/charmbracelet/bubbles/key"

type organizationsListKeyMap struct {
	// Short
	selection key.Binding
	refresh   key.Binding

	// Full
	selectionFull key.Binding
	refreshFull   key.Binding
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
	}
}
