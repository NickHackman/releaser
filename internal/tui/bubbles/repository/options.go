package repository

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Selection key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		Selection: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("<space>", "toggle select"),
		),
	}
}
