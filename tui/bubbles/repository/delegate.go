package repository

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

type Delegate struct{}

func (d Delegate) Height() int {
	return 3
}

func (d Delegate) Spacing() int {
	return 2
}

func (d Delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d Delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	genericItem, ok := listItem.(Item)
	if !ok {
		return
	}

	item := genericItem.R.Repo

	var output strings.Builder
	output.WriteString(titleStyle.Render(item.GetName()))

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
