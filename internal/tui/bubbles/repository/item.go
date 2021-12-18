package repository

import (
	"github.com/NickHackman/tagger/internal/service"
)

type Item struct {
	R        *service.ReleaseableRepoResponse
	Preview  string
	Selected bool
}

func (oi Item) FilterValue() string {
	return oi.R.Repo.GetName()
}
