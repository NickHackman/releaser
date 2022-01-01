package organization

import (
	"github.com/NickHackman/releaser/internal/github"
)

type Item struct {
	*github.OrgResponse
}

func (i Item) FilterValue() string {
	return i.Org.GetLogin()
}
