package organizations

import (
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui/bubbles/organization"
	tea "github.com/charmbracelet/bubbletea"
)

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
