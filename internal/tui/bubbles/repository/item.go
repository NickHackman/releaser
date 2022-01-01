package repository

import (
	"context"
	"strings"

	"github.com/NickHackman/releaser/internal/edit"
	"github.com/NickHackman/releaser/internal/service"
	"github.com/NickHackman/releaser/internal/template"
	"github.com/NickHackman/releaser/internal/tui/config"
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
	Version                          string
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
		"version":     i.Version,
	}

	var result *struct {
		Branch      string
		Version     string
		Description string
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
		Version:                 result.Version,
	}, nil
}
