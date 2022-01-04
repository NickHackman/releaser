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
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "releaser",
	Short: "Create releases for a whole GitHub organization with ease",
	Long: `A CLI application that will take the diff of all repositories 
in a GitHub organization and create releases for only the repositories
that have changes. The description associated with the release is
a Go Sprig template that will be used for all descriptions.

Version Methods:

Method to determine the new version.

Methods:
major        Increment Major version 1.3.0 -> 2.0.0
minor        Increment Minor version 1.3.0 -> 1.4.0
patch        Increment Patch version 1.3.0 -> 1.3.1

Template:

Top Level Variables:

{{ .RepositoryName }}                  hello-world	
{{ .RepositoryOwner }}                 octocat
{{ .RepositoryURL }}                   https://github.com/octocat/hello-world	
{{ .RepositoryDescription }}           Example description
{{ .RepositoryDefaultBranch }}         main
{{ .Commits }}                         List of commits

Commit:

{{ .Sha }}                             Unique identifier for commit
{{ .URL }}                             URL to commit
{{ .Summary }}                         First line of the commit message
{{ .Message }}                         Full commit message (includes newlines)

Author/Committer:
{{ .AuthorUsername }}                  octocat (GitHub Username)
{{ .AuthorName }}                      octocat (Commit Name)
{{ .AuthorEmail }}                     octocat@github.com
{{ .AuthorDate }} 
{{ .AuthorURL }}                       https://github.com/octocat

Templates also include Sprig functions: https://masterminds.github.io/sprig/strings.html

Template Example:

{{ range .Commits }}
{{ substr 0 8 .Sha }} committed by {{ .CommitterUsername }} and authored by {{ .AuthorUsername }} {{ .Summary }}
{{ end }}

Examples:

Log into github.com:

releaser login

Log into GitHub Enterprise:

releaser login --auth.url git.enterprise.com

Specify how to determine the version:

releaser --version.change minor

Bypass the UI entirely and create releases:

releaser --org example --repositories example1,example2,example3
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.InitViper(cfgFile, cmd); err != nil {
			if CreatedConfigErr, ok := err.(config.CreatedConfigErr); ok {
				fmt.Println(CreatedConfigErr.Error())
				os.Exit(0)
			} else {
				cobra.CheckErr(err)
			}
		}

		config, err := config.Load()
		cobra.CheckErr(err)

		if err := config.CheckAuth(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		gh, err := github.New().Host(config.Host).Token(config.Token).Build()
		cobra.CheckErr(err)

		// if token is provided fetch user's Username
		if config.Username == "" {
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			defer cancel()

			user, err := gh.User(ctx)
			cobra.CheckErr(err)

			config.Username = user.GetLogin()
		}

		cobra.CheckErr(
			tui.Execute(gh, config),
		)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/releaser/config.yaml)")
	rootCmd.PersistentFlags().String("host", "github.com", "Hostname of GitHub or GitHub Enterprise")
	rootCmd.Flags().String("token", "", "GitHub Oauth Token")
	rootCmd.Flags().StringP("org", "o", "", "GitHub organization to create releases")
	rootCmd.Flags().String("template", "", "Go template that is the default message for all releases")
	rootCmd.Flags().DurationP("timeout", "t", time.Minute, "Timeout duration to wait for GitHub to respond before exiting")
	rootCmd.Flags().StringP("branch", "b", "", "Branch to create releases on (defaults to Repository's default branch)")
	rootCmd.Flags().String("version.change", "", "Method to determine the new version based off the previous (major, minor, patch)")
	rootCmd.Flags().StringSlice("repositories", make([]string, 0), "Repositories to release (if this flag is provided then noninteractive UI)")
}
