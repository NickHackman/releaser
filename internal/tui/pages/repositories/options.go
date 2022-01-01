package repositories

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Publish       key.Binding
	RefreshRepos  key.Binding
	Open          key.Binding
	RefreshConfig key.Binding
	More          key.Binding
	Quit          key.Binding
}

func newKeyMap() *keyMap {
	keys := &keyMap{
		Publish: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "create release(s)"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open URL"),
		),
		RefreshRepos: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh repositories"),
		),
		RefreshConfig: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh config"),
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
