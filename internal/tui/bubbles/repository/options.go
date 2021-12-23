package repository

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Selection key.Binding
	Template  key.Binding
	Edit      key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		// Short
		Selection: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("<space>", "toggle select"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
	}
}
