package organization

import (
	"github.com/NickHackman/tagger/internal/service"
)

type Item struct {
	R *service.OrgResponse
}

func (oi Item) FilterValue() string {
	return oi.R.Org.GetLogin()
}
