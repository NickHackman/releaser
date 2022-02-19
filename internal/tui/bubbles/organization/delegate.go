package organization

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"
)

const (
	terminalWidth = 70
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
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	var output strings.Builder
	output.WriteString(titleStyle.Render(i.Login))

	if i.Description != "" {
		desc := truncate.StringWithTail(i.Description, terminalWidth, "...")
		output.WriteString("\n" + descriptionStyle.Render(desc))
	}

	if i.URL != "" {
		url := truncate.StringWithTail(i.URL, terminalWidth, "...")
		output.WriteString("\n" + urlStyle.Render(url))
	}

	render := unselectedStyle.MaxWidth(m.Width()).Render
	if index == m.Index() {
		render = selectedStyle.MaxWidth(m.Width()).Render
	}

	fmt.Fprint(w, render(output.String()))
}
