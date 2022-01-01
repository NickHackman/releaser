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
	"time"

	"github.com/NickHackman/tagger/internal/service"
	"github.com/NickHackman/tagger/internal/tui"
	"github.com/NickHackman/tagger/internal/tui/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed template-instructions.txt
var templateInstructions string

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish tags/releases to repositories in a GitHub organization",
	Long: `Publish tags/releases to repositories in a GitHub organization

Version Methods:

Method to determine the new version.

Methods:
major        Increment Major version 1.3.0 -> 2.0.0
minor        Increment Minor version 1.3.0 -> 1.4.0
patch        Increment Patch version 1.3.0 -> 1.3.1

` + templateInstructions,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := viper.GetString("token")
		url := viper.GetString("url")

		gh, err := service.NewGitHub().URL(url).Token(token).Build()
		if err != nil {
			return err
		}

		config := &config.Config{
			Branch:               viper.GetString("branch"),
			Org:                  viper.GetString("org"),
			Timeout:              viper.GetDuration("timeout"),
			TemplateString:       viper.GetString("template"),
			TemplateInstructions: templateInstructions,
		}

		return tui.Execute(gh, config)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringP("org", "o", "", "GitHub organization to create tags")
	publishCmd.Flags().String("template", "", "Go template that is the default message for all tags/releases")
	publishCmd.Flags().DurationP("timeout", "t", time.Minute, "Timeout duration to wait for GitHub to respond before exiting")
	publishCmd.Flags().StringP("branch", "b", "", "Branch to create releases on (defaults to Repository's default branch)")

	cobra.CheckErr(viper.BindPFlag("template", publishCmd.Flags().Lookup("template")))
	cobra.CheckErr(viper.BindPFlag("org", publishCmd.Flags().Lookup("org")))
	cobra.CheckErr(viper.BindPFlag("timeout", publishCmd.Flags().Lookup("timeout")))
	cobra.CheckErr(viper.BindPFlag("branch", publishCmd.Flags().Lookup("branch")))
}
