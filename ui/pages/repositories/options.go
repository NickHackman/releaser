package repositories

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type keyMap struct {
	// Short
	Selection    key.Binding
	SaveTemplate key.Binding
	Template     key.Binding
	Edit         key.Binding
	Version      key.Binding
	Quit         key.Binding

	// List builtins
	down                 key.Binding
	up                   key.Binding
	prevPage             key.Binding
	nextPage             key.Binding
	goToStart            key.Binding
	goToEnd              key.Binding
	filter               key.Binding
	clearFilter          key.Binding
	cancelWhileFiltering key.Binding
	acceptWhileFiltering key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		// Short
		Selection: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("<space>", "toggle select"),
		),
		SaveTemplate: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save template"),
		),
		Template: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "edit template"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Version: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "edit version"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),

		// List Builtin
		up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		prevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		nextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		goToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		goToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		clearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),

		// Filtering
		cancelWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		acceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),
	}
}

func (km *keyMap) IsListBuiltin(msg tea.KeyMsg) bool {
	return key.Matches(msg, km.down) ||
		key.Matches(msg, km.up) ||
		key.Matches(msg, km.prevPage) ||
		key.Matches(msg, km.nextPage) ||
		key.Matches(msg, km.goToStart) ||
		key.Matches(msg, km.goToEnd) ||
		key.Matches(msg, km.filter) ||
		key.Matches(msg, km.clearFilter) ||
		key.Matches(msg, km.cancelWhileFiltering) ||
		key.Matches(msg, km.acceptWhileFiltering)
}

func (km *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.up, km.down, km.filter, km.Selection, km.SaveTemplate, km.Template, km.Version, km.Edit, km.Quit}
}

func (km *keyMap) ShortHelpFilter() []key.Binding {
	return []key.Binding{km.acceptWhileFiltering, km.cancelWhileFiltering, km.Quit}
}

func (km *keyMap) ShortHelpFilterApplied() []key.Binding {
	return []key.Binding{km.up, km.down, km.clearFilter, km.Selection, km.SaveTemplate, km.Template, km.Version, km.Edit, km.Quit}
}

func (km *keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
