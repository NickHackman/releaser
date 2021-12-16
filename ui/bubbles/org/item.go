package org

import (
	"github.com/google/go-github/v41/github"
)

type Item struct {
	*github.Organization
}

func New(org *github.Organization) Item {
	return Item{org}
}

func (oi Item) FilterValue() string {
	return oi.GetLogin()
}
