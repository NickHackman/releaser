package organizations

import (
	"bytes"

	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/bubbles/organization"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"
)

type errorCmd error

type loadedOrganizationCmd struct {
	Item  organization.Item
	Total int
}

func loadOrganizationsCmd(channel <-chan *github.OrgResponse) tea.Cmd {
	return func() tea.Msg {
		org, ok := <-channel
		if !ok {
			return nil
		}

		return loadedOrganizationCmd{
			Item: organization.Item{Login: org.Org.Login,
				Description: org.Org.Description,
				URL:         org.Org.URL,
			}, Total: org.Total}
	}
}

func (m *Model) openURLCmd() tea.Cmd {
	return func() tea.Msg {
		item, ok := m.list.SelectedItem().(organization.Item)
		if !ok {
			return nil
		}

		var output bytes.Buffer
		browser.Stdout = &output

		if err := browser.OpenURL(item.URL); err != nil {
			return errorCmd(err)
		}

		return nil
	}
}
