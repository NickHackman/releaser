package github

import (
	"fmt"
	"net/http"

	"github.com/cli/browser"
	oauth "github.com/cli/oauth/webapp"
)

const (
	githubClientID = "f1afcf6b3e2a5972dfcf"
	// Safe to embed secret as it can be extracted from a released artifact.
	//
	// Following pattern from https://github.com/cli/cli/pull/492.
	// nolint:gosec // safe to embed secret as it can be extracted from a released artifact.
	githubClientSecret = "f79d57e38718ac9accd306432da4c623875d301a"
)

func Auth(host string) (string, error) {
	flow, err := oauth.InitFlow()
	if err != nil {
		return "", fmt.Errorf("failed to initialize Oauth flow: %v", err)
	}

	params := oauth.BrowserParams{
		ClientID:    githubClientID,
		RedirectURI: "http://127.0.0.1/callback",
		Scopes:      []string{"repo", "read:org"},
		AllowSignup: true,
	}

	baseURL := fmt.Sprintf("https://%s", host)
	authURL := fmt.Sprintf("%s/login/oauth/authorize", baseURL)
	browserURL, err := flow.BrowserURL(authURL, params)
	if err != nil {
		return "", fmt.Errorf("failed to set browser URL: %v", err)
	}

	go func() {
		_ = flow.StartServer(nil)
	}()

	if err = browser.OpenURL(browserURL); err != nil {
		return "", fmt.Errorf("failed to open browser: %v", err)
	}

	accessTokenURL := fmt.Sprintf("%s/login/oauth/access_token", baseURL)

	httpClient := http.DefaultClient
	accessToken, err := flow.AccessToken(httpClient, accessTokenURL, githubClientSecret)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub Oauth token: %v", err)
	}

	return accessToken.Token, nil
}
