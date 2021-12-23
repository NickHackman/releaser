package repositories

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Selection key.Binding
	Publish   key.Binding
	Template  key.Binding
	Edit      key.Binding
	More      key.Binding
	Quit      key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		// Short
		Selection: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("<space>", "toggle select"),
		),
		Publish: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "publish tags"),
		),
		Template: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "edit template"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
		),
		More: key.NewBinding(
			key.WithKeys("?"),
		),
	}
}
