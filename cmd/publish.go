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

	"github.com/NickHackman/tagger/service"
	"github.com/NickHackman/tagger/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish tags/releases to repositories in a GitHub organization",
	Long: `Publish tags/releases to repositories in a GitHub organization

Template Variables:

TODO: exhaustive list of template variables

In addition to the above variables that are injected automatically, Tagger uses Sprig and makes all functions available

Sprig Documentation: https://masterminds.github.io/sprig/.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := viper.GetString("token")
		url := viper.GetString("url")
		timeout := viper.GetDuration("timeout")

		gh, err := service.NewGitHub().Url(url).Token(token).Build()
		if err != nil {
			return err
		}

		return ui.Execute(gh, timeout)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	rootCmd.Flags().StringP("org", "o", "", "GitHub organization to create tags")
	rootCmd.Flags().String("template", "", "Go template that is the default message for all tags/releases")
	rootCmd.Flags().DurationP("timeout", "t", time.Minute, "Timeout duration to wait for GitHub to respond before exiting (default 1m)")

	viper.BindPFlag("template", rootCmd.Flags().Lookup("template"))
	viper.BindPFlag("org", rootCmd.Flags().Lookup("org"))
	viper.BindPFlag("timeout", rootCmd.Flags().Lookup("timeout"))
}
