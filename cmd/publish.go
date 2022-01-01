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
	"time"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui"
	"github.com/NickHackman/releaser/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish releases to repositories in a GitHub organization",
	Long: `Publish releases to repositories in a GitHub organization

Version Methods:

Method to determine the new version.

Methods:
major        Increment Major version 1.3.0 -> 2.0.0
minor        Increment Minor version 1.3.0 -> 1.4.0
patch        Increment Patch version 1.3.0 -> 1.3.1

Template Instructions

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

Example:

{{ range .Commits }}
{{ substr 0 8 .Sha }} committed by {{ .CommitterUsername }} and authored by {{ .AuthorUsername }} {{ .Summary }}
{{ end }}
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := viper.GetString("token")
		url := viper.GetString("url")

		gh, err := github.New().URL(url).Token(token).Build()
		if err != nil {
			return err
		}

		change, err := version.ChangeFromString(viper.GetString("version.change"))
		if err != nil {
			return err
		}

		config := &config.Config{
			Branch:         viper.GetString("branch"),
			Org:            viper.GetString("org"),
			Timeout:        viper.GetDuration("timeout"),
			TemplateString: viper.GetString("template"),
			VersionChange:  change,
		}

		return tui.Execute(gh, config)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringP("org", "o", "", "GitHub organization to create releases")
	publishCmd.Flags().String("template", "", "Go template that is the default message for all releases")
	publishCmd.Flags().DurationP("timeout", "t", time.Minute, "Timeout duration to wait for GitHub to respond before exiting")
	publishCmd.Flags().StringP("branch", "b", "", "Branch to create releases on (defaults to Repository's default branch)")
	publishCmd.Flags().String("version.change", "", "Method to determine the new version based off the previous (major, minor, patch)")

	cobra.CheckErr(viper.BindPFlag("template", publishCmd.Flags().Lookup("template")))
	cobra.CheckErr(viper.BindPFlag("org", publishCmd.Flags().Lookup("org")))
	cobra.CheckErr(viper.BindPFlag("timeout", publishCmd.Flags().Lookup("timeout")))
	cobra.CheckErr(viper.BindPFlag("branch", publishCmd.Flags().Lookup("branch")))
	cobra.CheckErr(viper.BindPFlag("version.change", publishCmd.Flags().Lookup("version.change")))
}
