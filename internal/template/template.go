package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/NickHackman/tagger/internal/service"
	"github.com/google/go-github/v41/github"
)

type tag struct {
	params map[string]interface{}
}

type commit struct {
	Sha     string
	URL     string
	Summary string
	Message string

	AuthorUsername string
	AuthorName     string
	AuthorEmail    string
	AuthorDate     time.Time
	AuthorURL      string

	CommitterUsername string
	CommitterName     string
	CommitterEmail    string
	CommitterDate     time.Time
	CommitterURL      string
}

func commitFrom(c *github.RepositoryCommit) *commit {
	return &commit{
		Sha:     c.GetSHA(),
		URL:     c.GetHTMLURL(),
		Message: c.GetCommit().GetMessage(),
		Summary: strings.Split(c.GetCommit().GetMessage(), "\n")[0],

		AuthorUsername: c.GetAuthor().GetLogin(),
		AuthorURL:      c.GetAuthor().GetURL(),
		AuthorName:     c.GetCommit().GetAuthor().GetName(),
		AuthorEmail:    c.GetCommit().GetAuthor().GetEmail(),
		AuthorDate:     c.GetCommit().GetAuthor().GetDate(),

		CommitterUsername: c.GetCommitter().GetLogin(),
		CommitterURL:      c.GetCommitter().GetURL(),
		CommitterName:     c.GetCommit().GetCommitter().GetName(),
		CommitterEmail:    c.GetCommit().GetCommitter().GetEmail(),
		CommitterDate:     c.GetCommit().GetCommitter().GetDate(),
	}
}

func new(repo *github.Repository, commits []*github.RepositoryCommit) *tag {
	var templateCommits []*commit

	for _, c := range commits {
		templateCommits = append(templateCommits, commitFrom(c))
	}

	return &tag{
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

func (tag *tag) execute(templatedString string) (string, error) {
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

func Preview(r *service.ReleaseableRepoResponse, templatedString string) string {
	tagTemplate := new(r.Repo, r.Commits)
	content, err := tagTemplate.execute(templatedString)
	if err != nil {
		content = fmt.Sprintf("%s\n\n# Error: %v", templatedString, err)
	}

	return content
}
