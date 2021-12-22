/*
Copyright © 2021 Nick Hackman <snickhackman@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/cli/browser"
	oauth "github.com/cli/oauth/webapp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	githubClientId     = "f1afcf6b3e2a5972dfcf"
	githubClientSecret = "f79d57e38718ac9accd306432da4c623875d301a"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log into GitHub",
	Long: `Authenticate with GitHub to perform other operations.

Examples:

tagger login

tagger login --auth.url git.enterprise.com`,
	Run: func(cmd *cobra.Command, args []string) {
		token, err := githubOauthFlow()
		cobra.CheckErr(err)

		config := viper.GetViper()
		config.Set("token", token)
		err = config.WriteConfig()
		cobra.CheckErr(err)
	},
}

func githubOauthFlow() (string, error) {
	flow, err := oauth.InitFlow()
	if err != nil {
		return "", fmt.Errorf("failed to initialize Oauth flow: %v", err)
	}

	params := oauth.BrowserParams{
		ClientID:    githubClientId,
		RedirectURI: "http://127.0.0.1/callback",
		Scopes:      []string{"repo", "read:org"},
		AllowSignup: true,
	}

	url := viper.GetString("auth.url")
	authUrl := fmt.Sprintf("%s/login/oauth/authorize", url)
	browserUrl, err := flow.BrowserURL(authUrl, params)
	if err != nil {
		return "", fmt.Errorf("failed to set browser URL: %v", err)
	}

	go func() {
		_ = flow.StartServer(nil)
	}()

	if err = browser.OpenURL(browserUrl); err != nil {
		return "", fmt.Errorf("failed to open browser: %v", err)
	}

	accessTokenUrl := fmt.Sprintf("%s/login/oauth/access_token", url)

	httpClient := http.DefaultClient
	accessToken, err := flow.AccessToken(httpClient, accessTokenUrl, githubClientSecret)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub Oauth token: %v", err)
	}

	return accessToken.Token, nil
}

func init() {
	loginCmd.Flags().String("auth.url", "https://github.com/", "GitHub url (default is https://github.com/)")
	cobra.CheckErr(viper.BindPFlag("auth.url", loginCmd.Flags().Lookup("auth.url")))

	rootCmd.AddCommand(loginCmd)
}
