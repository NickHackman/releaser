package service

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type GitHubBuilder struct {
	url   string
	token string
}

func NewGitHub() *GitHubBuilder {
	return &GitHubBuilder{}
}

func (ghb *GitHubBuilder) Url(url string) *GitHubBuilder {
	ghb.url = url
	return ghb
}

func (ghb *GitHubBuilder) Token(token string) *GitHubBuilder {
	ghb.token = token
	return ghb
}

func (ghb *GitHubBuilder) Build() (*GitHub, error) {
	if ghb.token == "" {
		return nil, errors.New("failed to authenticate missing GitHub Oauth token.\nRun `tagger login`")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghb.token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	var err error
	client.BaseURL, err = url.Parse(ghb.url)
	if err != nil {
		return nil, err
	}

	return &GitHub{client: client}, nil
}
