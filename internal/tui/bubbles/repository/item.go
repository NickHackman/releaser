package repository

import (
	"context"
	"strings"

	"github.com/NickHackman/tagger/internal/edit"
	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/template"
	"github.com/NickHackman/tagger/internal/tui/config"
	"gopkg.in/yaml.v3"
)

const (
	editBranchInstructions = "Editing the branch will overwrite all description changes."
)

type Item struct {
	*service.ReleaseableRepoResponse `yaml:"-"`
	Selected                         bool `yaml:"-"`
	Preview                          string
	Branch                           string
}

func (oi Item) FilterValue() string {
	return oi.Repo.GetName()
}

func (oi Item) Select() Item {
	return Item{
		ReleaseableRepoResponse: oi.ReleaseableRepoResponse,
		Preview:                 oi.Preview,
		Branch:                  oi.Branch,
		Selected:                !oi.Selected,
	}
}

func (i Item) Edit(gh *service.GitHub, config *config.Config) (Item, error) {
	branchNames := []string{"Branches:"}
	for _, branch := range i.Branches {
		name := branch.GetName()

		if name == i.Repo.GetDefaultBranch() {
			name += " *"
		}

		branchNames = append(branchNames, "  "+name) // Indent list
	}

	yamlRepresentation := map[string]interface{}{
		"description": i.Preview,
		"branch":      &yaml.Node{Kind: yaml.ScalarNode, Value: i.Branch, HeadComment: strings.Join(branchNames, "\n"), LineComment: editBranchInstructions},
	}

	var result *struct {
		Description string
		Branch      string
	}

	if err := edit.Invoke(&yamlRepresentation, &result); err != nil {
		return i, err
	}

	response := i.ReleaseableRepoResponse
	description := result.Description
	if i.Branch != result.Branch {
		var err error

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
		defer cancel()

		if response, err = gh.ReleaseableRepo(ctx, i.Repo.GetOwner().GetLogin(), i.Repo, result.Branch); err != nil {
			return i, err
		}

		description = template.Preview(response, config.TemplateString)
	}

	return Item{
		ReleaseableRepoResponse: response,
		Selected:                i.Selected,
		Preview:                 description,
		Branch:                  result.Branch,
	}, nil
}
