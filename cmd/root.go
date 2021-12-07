/*
Copyright © 2021 Nick Hackman

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
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var org string
var template string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tagger",
	Short: "Create tags for a whole GitHub organization with ease",
	Long: `A CLI application that will take the diff of all repositories 
in a GitHub organization and create Tags/Releases for only the repositories
that have changes. The description/message associated with the Tag/Release is
a Go Sprig Template that will be used by default for all messages.

Examples:

tagger --org "GitHub org"

tagger --org "GitHub org" --config "tagger.yml"

tagger --org "GitHub org" --template "This is an example"

Template Variables:

TODO: exhaustive list of template variables

In addition to the above variables that are injected automatically, Tagger uses Sprig and makes all functions available

Sprig Documentation: https://masterminds.github.io/sprig/.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/tagger.yaml)")
	rootCmd.PersistentFlags().StringVar(&org, "org", "", "GitHub organization to create tags")
	rootCmd.PersistentFlags().StringVar(&template, "template", "", "Go template that is the default message for all tags/releases")
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

		// Search config in config directory with name "tagger" (without extension).
		viper.AddConfigPath(config)
		viper.SetConfigType("yaml")
		viper.SetConfigName("tagger")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
