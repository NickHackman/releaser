package repositories

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Publish  key.Binding
	Template key.Binding
	More     key.Binding
	Quit     key.Binding
}

func newKeyMap() *keyMap {
	keys := &keyMap{
		Publish: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "publish tags"),
		),
		Template: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "edit template"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
		),
		More: key.NewBinding(
			key.WithKeys("?"),
		),
	}

	keys.Publish.SetEnabled(false)

	return keys
}
