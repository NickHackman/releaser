package repositories

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Publish      key.Binding
	RefreshRepos key.Binding
	Open         key.Binding
	Refresh      key.Binding
	More         key.Binding
	ToggleAll    key.Binding
	Quit         key.Binding
}

func newKeyMap() *keyMap {
	keys := &keyMap{
		Publish: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "create release(s)"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle all repositories"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open URL"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
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
