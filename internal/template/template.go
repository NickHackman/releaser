package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/google/go-github/v41/github"
)

type Tag struct {
	params map[string]interface{}
}

type commit struct {
	Sha     string
	Url     string
	Summary string
	Message string

	AuthorUsername string
	AuthorName     string
	AuthorEmail    string
	AuthorDate     time.Time
	AuthorUrl      string

	CommitterUsername string
	CommitterName     string
	CommitterEmail    string
	CommitterDate     time.Time
	CommitterUrl      string
}

func commitFrom(c *github.RepositoryCommit) *commit {
	return &commit{
		Sha:     c.GetSHA(),
		Url:     c.GetHTMLURL(),
		Message: c.GetCommit().GetMessage(),
		Summary: strings.Split(c.GetCommit().GetMessage(), "\n")[0],

		AuthorUsername: c.GetAuthor().GetLogin(),
		AuthorUrl:      c.GetAuthor().GetURL(),
		AuthorName:     c.GetCommit().GetAuthor().GetName(),
		AuthorEmail:    c.GetCommit().GetAuthor().GetEmail(),
		AuthorDate:     c.GetCommit().GetAuthor().GetDate(),

		CommitterUsername: c.GetCommitter().GetLogin(),
		CommitterUrl:      c.GetCommitter().GetURL(),
		CommitterName:     c.GetCommit().GetCommitter().GetName(),
		CommitterEmail:    c.GetCommit().GetCommitter().GetEmail(),
		CommitterDate:     c.GetCommit().GetCommitter().GetDate(),
	}
}

func NewTag(repo *github.Repository, commits []*github.RepositoryCommit) *Tag {
	var templateCommits []*commit

	for _, c := range commits {
		templateCommits = append(templateCommits, commitFrom(c))
	}

	return &Tag{
		params: map[string]interface{}{
			"RepositoryName":          repo.GetName(),
			"RepositoryOwner":         repo.GetOwner().GetLogin(),
			"RepositoryUrl":           repo.GetHTMLURL(),
			"RepositoryDescription":   repo.GetDescription(),
			"RepositoryDefaultBranch": repo.GetDefaultBranch(),
			"Commits":                 templateCommits,
		},
	}
}

func (tag *Tag) Execute(templatedString string) (string, error) {
	sf := sprig.TxtFuncMap()

	t, err := template.New("template").Funcs(sf).Parse(templatedString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, tag.params); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}
