package organization

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	oi, ok := listItem.(Item)
	if !ok {
		return
	}

	org := oi.R.Org

	var output strings.Builder
	output.WriteString(titleStyle.Render(org.GetLogin()))

	if description := org.GetDescription(); description != "" {
		output.WriteString("\n" + descriptionStyle.Render(description))
	}

	if url := org.GetHTMLURL(); url != "" {
		output.WriteString("\n" + urlStyle.Render(url))
	}

	render := unselectedStyle.Render
	if index == m.Index() {
		render = selectedStyle.Render
	}

	fmt.Fprint(w, render(output.String()))
}
