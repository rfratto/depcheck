package tracker

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v48/github"
	"golang.org/x/mod/semver"
)

// Github checks for outdated dependencies on Github projects.
type Github struct {
	check []GithubDependency
	cli   *github.Client
}

// GithubDependency is a dependency on a Github project.
type GithubDependency struct {
	Project string            `yaml:"project"`
	Version string            `yaml:"version"`
	Options DependencyOptions `yaml:",inline"`
}

// UnmarshalYAML will unmarshal a string or an object into a GithubDependency.
func (d *GithubDependency) UnmarshalYAML(f func(interface{}) error) error {
	var (
		stringError error
		objectError error
	)

	// Try as a raw string
	var s string
	stringError = f(&s)
	if stringError == nil {
		return unmarshalGithubDependencyString(s, d)
	}

	// Then a whole object
	type githubDependency GithubDependency
	var v githubDependency
	objectError = f(&v)
	if objectError == nil {
		*d = GithubDependency(v)
		return nil
	}

	return fmt.Errorf(
		"could not parse Github dependency as a string (%s) or an object (%s)",
		stringError,
		objectError,
	)
}

func unmarshalGithubDependencyString(in string, dep *GithubDependency) error {
	depParts := strings.SplitN(in, " ", 2)
	if len(depParts) != 2 {
		return fmt.Errorf("invalid dependency %s: expected format '[owner]/[name] [version]'", in)
	}
	dep.Project = depParts[0]
	dep.Version = depParts[1]
	return nil
}

// NewGithub creates a new Github tracker.
func NewGithub(check []GithubDependency, cli *github.Client) *Github {
	return &Github{check: check, cli: cli}
}

// CheckOutdated will return the list of go module dependencies that can be updated.
func (c *Github) CheckOutdated(ctx context.Context) ([]Dependency, error) {
	var outdated []Dependency

	for _, d := range c.check {
		// Trim out github.com/ from the name, but it'll be added back later.
		sanitizedName := strings.TrimPrefix(d.Project, "github.com/")
		nameParts := strings.SplitN(sanitizedName, "/", 2)
		if len(nameParts) != 2 {
			return nil, fmt.Errorf("invalid project name %s", d.Project)
		}
		var (
			owner = nameParts[0]
			repo  = nameParts[1]
		)

		tags, _, err := c.cli.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
			Page:    0,
			PerPage: 1,
		})
		if err != nil {
			return nil, fmt.Errorf("couldn't get tags for %s: %w", d.Project, err)
		}
		if len(tags) == 0 {
			return nil, fmt.Errorf("%s has no tags", d.Project)
		}

		if d.Options.IgnoreVersionPattern.Matches(tags[0].GetName()) {
			continue
		}

		if semver.Compare(d.Version, tags[0].GetName()) == -1 {
			outdated = append(outdated, Dependency{
				Name:           "github.com/" + sanitizedName,
				CurrentVersion: d.Version,
				LatestVersion:  tags[0].GetName(),
			})
		}
	}

	return outdated, nil
}
