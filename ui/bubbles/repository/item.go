package repository

import (
	"github.com/NickHackman/tagger/service"
)

type Item struct {
	R       *service.ReleaseableRepoResponse
	Preview string
}

func (oi Item) FilterValue() string {
	return oi.R.Repo.GetName()
}
