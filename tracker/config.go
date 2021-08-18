package tracker

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
)

var DefaultConfig = Config{
	IssueTitleTemplate: "Update {{.Name}} to {{.LatestVersion}}",
	IssueTextTemplate:  "An update for `{{.Name}}` (version `{{.LatestVersion}}`) is now available. Version `{{.CurrentVersion}}` is currently in use.",
	OutdatedLabel:      "outdated-dependency",
}

// Config represents the tracker configuration.
type Config struct {
	// IssueRepository is the repo to create issues in. If empty, defaults to
	// GITHUB_REPOSITORY and fails if neither is set.
	IssueRepository string `yaml:"issue_repository"`

	// IssueTitleTemplate and IssueTextTemplate are go text/templates that will be
	// used for creating the issue titles and content, respectively.
	IssueTitleTemplate string `yaml:"issue_title_template"`
	IssueTextTemplate  string `yaml:"issue_text_template"`

	// OutdatedLabel is the label to attach to created issues.
	OutdatedLabel string `yaml:"outdated_label"`

	// GoModules are a list of go module dependencies to check.
	GoModules []GoModule `yaml:"go_modules"`

	// GithubDeps are a list of github repos to check.
	GithubDeps []GithubDependency `yaml:"github_repos"`
}

// DependencyOptions are options for individual dependencies.
type DependencyOptions struct {
	// IgnoreVersionPattern is a pattern that allows you to ignore specific
	// versions matching the given string.
	IgnoreVersionPattern *Regexp `yaml:"ignore_version_pattern"`
}

// Regexp is a regex that can be unmarshaled from a string.
type Regexp regexp.Regexp

// Matches returns true if s matches the regular expression.
func (r *Regexp) Matches(s string) bool {
	if r == nil {
		return false
	}
	return (*regexp.Regexp)(r).MatchString(s)
}

func (r *Regexp) UnmarshalYAML(f func(v interface{}) error) error {
	var s string
	if err := f(&s); err != nil {
		return err
	}

	c, err := regexp.Compile(s)
	if err != nil {
		return fmt.Errorf("invalid regex: %w", err)
	}
	*r = Regexp(*c)
	return err
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
