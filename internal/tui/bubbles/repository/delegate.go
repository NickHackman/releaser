package repository

import (
	"fmt"
	"io"
	"strings"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"
)

const (
	terminalWidth = 70
)

type Delegate struct {
	Keys   *keyMap
	gh     *github.Client
	config *config.Config
}

func NewDelegate(gh *github.Client, config *config.Config) Delegate {
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.Keys.Selection):
			current, ok := m.SelectedItem().(Item)
			if !ok {
				break
			}

			index := m.Index()
			newItem := current.Select()
			return m.SetItem(index, newItem)
		}
	}

	return nil
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
		text := truncate.StringWithTail(description, terminalWidth, "...")
		output.WriteString("\n" + descriptionStyle.Render(text))
	}

	if url := item.Repo.GetHTMLURL(); url != "" {
		text := truncate.StringWithTail(url, terminalWidth, "...")
		output.WriteString("\n" + urlStyle.Render(text))
	}

	render := unselectedStyle.MaxWidth(terminalWidth).Render
	if index == m.Index() {
		render = selectedStyle.MaxWidth(terminalWidth).Render
	}

	fmt.Fprint(w, render(output.String()))
}
