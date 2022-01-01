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
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "releaser",
	Short: "Create releases for a whole GitHub organization with ease",
	Long: `A CLI application that will take the diff of all repositories 
in a GitHub organization and create releases for only the repositories
that have changes. The description associated with the release is
a Go Sprig template that will be used by default for all descriptions.

GitHub Oauth:

releaser login

releaser --url "git.enterprise.com" login

Examples:

releaser publish --org "GitHub org"

releaser --config "releaser.yml" publish --org "GitHub org"

releaser publish --org "GitHub org" --template "This is an example"`,
	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/releaser.yaml)")
	rootCmd.PersistentFlags().String("url", "https://api.github.com/", "GitHub API url")
	rootCmd.PersistentFlags().String("token", "", "GitHub Oauth Token")

	cobra.CheckErr(viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")))
	cobra.CheckErr(viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url")))
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
	err := viper.ReadInConfig()
	cobra.CheckErr(err)
}
