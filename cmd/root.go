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
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/tui"
	"github.com/NickHackman/releaser/internal/version"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

//go:embed releaser.example.yml
var exampleConfig []byte

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

releaser login

releaser login --auth.url git.enterprise.com

releaser --org example --version.change minor
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/releaser.yaml)")
	rootCmd.Flags().String("url", "https://api.github.com/", "GitHub API url")
	rootCmd.Flags().String("token", "", "GitHub Oauth Token")
	rootCmd.Flags().StringP("org", "o", "", "GitHub organization to create releases")
	rootCmd.Flags().String("template", "", "Go template that is the default message for all releases")
	rootCmd.Flags().DurationP("timeout", "t", time.Minute, "Timeout duration to wait for GitHub to respond before exiting")
	rootCmd.Flags().StringP("branch", "b", "", "Branch to create releases on (defaults to Repository's default branch)")
	rootCmd.Flags().String("version.change", "", "Method to determine the new version based off the previous (major, minor, patch)")

	cobra.CheckErr(viper.BindPFlag("template", rootCmd.Flags().Lookup("template")))
	cobra.CheckErr(viper.BindPFlag("org", rootCmd.Flags().Lookup("org")))
	cobra.CheckErr(viper.BindPFlag("timeout", rootCmd.Flags().Lookup("timeout")))
	cobra.CheckErr(viper.BindPFlag("branch", rootCmd.Flags().Lookup("branch")))
	cobra.CheckErr(viper.BindPFlag("version.change", rootCmd.Flags().Lookup("version.change")))
	cobra.CheckErr(viper.BindPFlag("token", rootCmd.Flags().Lookup("token")))
	cobra.CheckErr(viper.BindPFlag("url", rootCmd.Flags().Lookup("url")))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		config, err := os.UserConfigDir()
		cobra.CheckErr(err)

		// Search config in config directory with name "releaser" (without extension).
		viper.AddConfigPath(config)
		viper.SetConfigType("yaml")
		viper.SetConfigName("releaser")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Only set the default config if the user is not providing a config via path
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && cfgFile == "" {
			config, err := os.UserConfigDir()
			cobra.CheckErr(err)

			defaultPath := filepath.Join(config, "releaser.yaml")

			err = ioutil.WriteFile(defaultPath, exampleConfig, 0600)
			cobra.CheckErr(err)

			fmt.Printf("Wrote example config file to %s\n", defaultPath)
			os.Exit(0)
		} else {
			cobra.CheckErr(err)
		}
	}
}
