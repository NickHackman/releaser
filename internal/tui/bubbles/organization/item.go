package organization

import (
	"github.com/NickHackman/tagger/internal/service"
)

type Item struct {
	*service.OrgResponse
}

func (i Item) FilterValue() string {
	return i.Org.GetLogin()
}
