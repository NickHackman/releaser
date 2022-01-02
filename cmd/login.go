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
	"os"

	"github.com/NickHackman/releaser/internal/config"
	"github.com/NickHackman/releaser/internal/github"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log into GitHub or GitHub Enterprise",
	Long: `Log into Github or Github Enterprise.

Examples:

releaser login

releaser --host git.enterprise.com login`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.InitViper(cfgFile, cmd); err != nil {
			if CreatedConfigErr, ok := err.(config.CreatedConfigErr); ok {
				fmt.Println(CreatedConfigErr.Error())
				os.Exit(0)
			} else {
				cobra.CheckErr(err)
			}
		}

		c, err := config.Load()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		token, err := github.Auth(c.Host)
		cobra.CheckErr(err)

		c.Token = token

		gh, err := github.New().Host(c.Host).Token(c.Token).Build()
		cobra.CheckErr(err)

		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		username, err := gh.Username(ctx)
		cobra.CheckErr(err)

		err = c.SaveHost(config.Auth{Username: username, Token: token})
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
