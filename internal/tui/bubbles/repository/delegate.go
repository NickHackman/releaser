package repository

import (
	"fmt"
	"io"
	"strings"

	"github.com/NickHackman/releaser/internal/service"
	"github.com/NickHackman/releaser/internal/tui/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

type Delegate struct {
	Keys   *keyMap
	gh     *service.GitHub
	config *config.Config
}

func NewDelegate(gh *service.GitHub, config *config.Config) Delegate {
	return Delegate{
		Keys:   newKeyMap(),
		gh:     gh,
		config: config,
	}
}

func (d Delegate) Height() int {
	return 3
}

func (d Delegate) Spacing() int {
	return 2
}

func (d Delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.Keys.Edit):
			current, ok := m.SelectedItem().(Item)
			if !ok {
				break
			}

			newItem, err := current.Edit(d.gh, d.config)
			if err != nil {
				return m.NewStatusMessage(fmt.Sprintf("Error: %v", err))
			}

			index := m.Index()
			cmds = append(cmds, m.SetItem(index, newItem), RefreshPreview)
		case key.Matches(msg, d.Keys.Selection):
			current, ok := m.SelectedItem().(Item)
			if !ok {
				break
			}

			index := m.Index()
			newItem := current.Select()
			cmds = append(cmds, m.SetItem(index, newItem), RefreshPreview)
		}
	}

	return tea.Batch(cmds...)
}

func (d Delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		return
	}

	var output strings.Builder
	output.WriteString(titleStyle.Render(item.Repo.GetName()))

	if item.Selected {
		output.WriteString(checkmarkStyle.Render(" ✓"))
	}

	if description := item.Repo.GetDescription(); description != "" {
		text := strings.Split(wordwrap.String(description, 75), "\n")[0]
		output.WriteString("\n" + descriptionStyle.Render(text))
	}

	if url := item.Repo.GetHTMLURL(); url != "" {
		output.WriteString("\n" + urlStyle.Render(url))
	}

	render := unselectedStyle.Render
	if index == m.Index() {
		render = selectedStyle.Render
	}

	fmt.Fprint(w, render(output.String()))
}
