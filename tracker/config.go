package tracker

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config represents the tracker configuration.
type Config struct {
	// IssueRepository is the repo to create issues in. If empty, defaults to
	// GITHUB_REPOSITORY and fails if neither is set.
	IssueRepository string `yaml:"issue_repository"`

	// OutdatedLabel is the label to attach to created issues.
	OutdatedLabel string `yaml:"outdated_label"`

	// GoModules are a list of go module dependencies to check.
	GoModules []string `yaml:"go_modules"`

	// GithubDeps are a list of github repos to check.
	GithubDeps []GithubDependency
}

// UnmarshalYAML unmarshals the Config with defaults applied.
func (c *Config) UnmarshalYAML(f func(v interface{}) error) error {
	type config Config
	type fullConfig struct {
		config      `yaml:",inline"`
		GithubRepos []string `yaml:"github_repos"`
	}
	var val fullConfig
	if err := f(&val); err != nil {
		return err
	}

	*c = Config(val.config)

	for _, r := range val.GithubRepos {
		depParts := strings.SplitN(r, " ", 2)
		if len(depParts) != 2 {
			return fmt.Errorf("invalid dependency %s: expected format '[owner]/[name] [version]'", r)
		}
		c.GithubDeps = append(c.GithubDeps, GithubDependency{
			Project: depParts[0],
			Version: depParts[1],
		})
	}

	return c.SetDefaults()
}

// SetDefaults applies default values to fields in the config.
func (c *Config) SetDefaults() error {
	if c.IssueRepository == "" {
		c.IssueRepository = os.Getenv("GITHUB_REPOSITORY")
	}
	if c.IssueRepository == "" {
		return fmt.Errorf("either GITHUB_REPOSITORY must be set in environment or issue_repository must be set in config")
	}

	return nil
}

// LoadConfig loads the Config via a file. The file is expected to be YAML.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}

	var c Config
	err = yaml.NewDecoder(f).Decode(&c)
	return &c, err
}
