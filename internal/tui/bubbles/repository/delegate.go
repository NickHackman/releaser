package repository

import (
	"fmt"
	"io"
	"strings"

	"github.com/NickHackman/tagger/internal/edit"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

type Delegate struct {
	Keys *keyMap
}

func NewDelegate() Delegate {
	return Delegate{
		Keys: newKeyMap(),
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
			currentItem := m.SelectedItem()
			current, ok := currentItem.(Item)
			if !ok {
				break
			}

			result, err := edit.Content(current.Preview, edit.ManualEditInstructions)
			if err != nil {
				cmds = append(cmds, m.NewStatusMessage(fmt.Sprintf("Error: %v", err)))
			}

			currentIndex := m.Index()
			newItem := Item{R: current.R, Preview: result}
			cmds = append(cmds, m.SetItem(currentIndex, newItem), RefreshPreview)
		case key.Matches(msg, d.Keys.Selection):
			item := m.SelectedItem()

			current, ok := item.(Item)
			if !ok {
				break
			}

			index := m.Index()
			cmds = append(cmds, m.SetItem(index, Item{R: current.R, Preview: current.Preview, Selected: !current.Selected}), RefreshPreview)
		}
	}

	return tea.Batch(cmds...)
}

func (d Delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	genericItem, ok := listItem.(Item)
	if !ok {
		return
	}

	item := genericItem.R.Repo

	var output strings.Builder
	output.WriteString(titleStyle.Render(item.GetName()))

	if genericItem.Selected {
		output.WriteString(checkmarkStyle.Render(" ✓"))
	}

	if item.GetDescription() != "" {
		text := strings.Split(wordwrap.String(item.GetDescription(), 75), "\n")[0]
		output.WriteString("\n" + descriptionStyle.Render(text))
	}

	if item.GetHTMLURL() != "" {
		output.WriteString("\n" + urlStyle.Render(item.GetHTMLURL()))
	}

	render := unselectedStyle.Render
	if index == m.Index() {
		render = selectedStyle.Render
	}

	fmt.Fprint(w, render(output.String()))
}
