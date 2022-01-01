package repository

import (
	"github.com/NickHackman/releaser/internal/github"
	"github.com/google/uuid"
)

type Item struct {
	*github.ReleaseableRepoResponse `yaml:"-"`
	Selected                        bool `yaml:"-"`
	Preview                         string
	Branch                          string
	Version                         string
	ID                              uuid.UUID
}

func (i Item) FilterValue() string {
	return i.Repo.GetName()
}

func (i Item) Select() Item {
	return Item{
		ReleaseableRepoResponse: i.ReleaseableRepoResponse,
		Preview:                 i.Preview,
		Branch:                  i.Branch,
		Selected:                !i.Selected,
		Version:                 i.Version,
	}
}
