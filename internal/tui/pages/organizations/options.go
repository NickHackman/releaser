package organizations

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Selection key.Binding
	Refresh   key.Binding
	Open      key.Binding
}

func newOrganizationsListKeyMap() *keyMap {
	return &keyMap{
		Selection: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open URL"),
		),
	}
}
