package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/NickHackman/releaser/internal/github"
	"github.com/NickHackman/releaser/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

//go:embed releaser.example.yml
var exampleConfig []byte

const (
	TemplateFlag      = "template"
	OrgFlag           = "org"
	TimeoutFlag       = "timeout"
	BranchFlag        = "branch"
	VersionChangeFlag = "version.change"
	TokenFlag         = "token"
	HostFlag          = "host"
)

// CreatedConfigErr error returned when InitViper fails due to the config not existing
// in the first place and was created at will.
type CreatedConfigErr struct {
	Path string
}

func (cce CreatedConfigErr) Error() string {
	return fmt.Sprintf("Wrote example config file to %s", cce.Path)
}

var flags = []string{
	TemplateFlag,
	OrgFlag,
	TimeoutFlag,
	BranchFlag,
	VersionChangeFlag,
	TokenFlag,
	HostFlag,
}

const (
	configFilename = "config.yaml"
	hostsFilename  = "hosts.yaml"
)

// TerminalConfig configuration maintained for the terminal
type TerminalConfig struct {
	Width    int
	Height   int
	Releases chan<- []*github.RepositoryReleaseResponse
}

// AuthHosts format for how auth is stored inside of hosts.yaml.
//
// github.com:
//   username: NickHackman
//   token: this-is-a-fake-auth-token
type AuthHosts map[string]Auth

type Auth struct {
	Username string
	Token    string
}

type Config struct {
	Username      string
	Host          string
	Org           string
	Branch        string
	Token         string
	Template      string
	Timeout       time.Duration
	VersionChange version.Change

	AuthHosts AuthHosts
	Terminal  *TerminalConfig
}

func (c *Config) SetSize(width, height int) {
	c.Terminal.Width = width
	c.Terminal.Height = height
}

func (c *Config) Size() (int, int) {
	return c.Terminal.Width, c.Terminal.Height
}

func configDir() (string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(config, "releaser"), nil
}

func InitViper(filename string, cmd *cobra.Command) error {
	for _, flag := range flags {
		if exists := cmd.Flags().Lookup(flag); exists == nil {
			continue
		}

		if err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag)); err != nil {
			return err
		}
	}

	// config was passed via CLI args
	if filename != "" {
		viper.SetConfigFile(filename)
	} else {
		config, err := configDir()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(config, 0700); err != nil {
			return err
		}

		viper.AddConfigPath(config)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Only set the default config if the user is not providing a config via path
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && filename == "" {
			dir, err := configDir()
			if err != nil {
				return err
			}

			configPath := filepath.Join(dir, configFilename)

			if err := ioutil.WriteFile(configPath, exampleConfig, 0600); err != nil {
				return err
			}

			hostsPath := filepath.Join(dir, hostsFilename)

			f, err := os.Create(hostsPath)
			if err != nil {
				return err
			}
			defer f.Close()

			return CreatedConfigErr{Path: configPath}
		} else {
			return err
		}
	}

	return nil
}

func (c *Config) Refresh() error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	c.Branch = viper.GetString(BranchFlag)
	c.Timeout = viper.GetDuration(TimeoutFlag)
	c.Template = viper.GetString(TemplateFlag)

	change, err := version.ChangeFromString(viper.GetString(VersionChangeFlag))
	if err != nil {
		return err
	}

	c.VersionChange = change

	return nil
}

func Load() (*Config, error) {
	change, err := version.ChangeFromString(viper.GetString(VersionChangeFlag))
	if err != nil {
		return nil, err
	}

	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	hostsPath := filepath.Join(dir, hostsFilename)

	if _, err := os.Stat(hostsPath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(hostsPath)
		if err != nil {
			return nil, err
		}

		f.Close()
	}

	hostsBytes, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		return nil, err
	}

	authHosts := make(AuthHosts)
	if err := yaml.Unmarshal(hostsBytes, &authHosts); err != nil {
		return nil, err
	}

	return &Config{
		Host:          viper.GetString(HostFlag),
		Token:         viper.GetString(TokenFlag),
		Branch:        viper.GetString(BranchFlag),
		Org:           viper.GetString(OrgFlag),
		Timeout:       viper.GetDuration(TimeoutFlag),
		Template:      viper.GetString(TemplateFlag),
		VersionChange: change,
		AuthHosts:     authHosts,
		Terminal:      &TerminalConfig{},
	}, nil
}

func (c *Config) CheckAuth() error {
	if c.Token == "" {
		auth, ok := c.AuthHosts[c.Host]
		if !ok {
			return fmt.Errorf("releaser is not authorized for host %s\n\nRun:\nreleaser --host %s login", c.Host, c.Host)
		}

		c.Token = auth.Token
		c.Username = auth.Username
	}

	return nil
}

func (c *Config) SaveHost(auth Auth) error {
	c.AuthHosts[c.Host] = auth

	out, err := yaml.Marshal(c.AuthHosts)
	if err != nil {
		return err
	}

	dir, err := configDir()
	if err != nil {
		return err
	}

	hostsPath := filepath.Join(dir, hostsFilename)

	return ioutil.WriteFile(hostsPath, out, 0600)
}
