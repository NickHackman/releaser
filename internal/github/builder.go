package github

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Builder struct {
	url   string
	token string
}

func New() *Builder {
	return &Builder{}
}

func (ghb *Builder) URL(url string) *Builder {
	ghb.url = url
	return ghb
}

func (ghb *Builder) Token(token string) *Builder {
	ghb.token = token
	return ghb
}

func (ghb *Builder) Build() (*Client, error) {
	if ghb.token == "" {
		return nil, errors.New("failed to authenticate missing GitHub Oauth token.\nRun `releaser login`")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghb.token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	var err error
	client.BaseURL, err = url.Parse(ghb.url)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}
