package tracker

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"golang.org/x/mod/semver"
)

// Github checks for outdated dependencies on Github projects.
type Github struct {
	check []GithubDependency
}

// GithubDependency is a dependency on a Github project.
type GithubDependency struct {
	Project string
	Version string
}

// NewGithub creates a new Github tracker.
func NewGithub(check []GithubDependency) *Github {
	return &Github{check: check}
}

// CheckOutdated will return the list of go module dependencies that can be updated.
func (c *Github) CheckOutdated(ctx context.Context) ([]Dependency, error) {
	var outdated []Dependency

	client := github.NewClient(nil)

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

		tags, _, err := client.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
			Page:    0,
			PerPage: 1,
		})
		if err != nil {
			return nil, fmt.Errorf("couldn't get tags for %s: %w", d.Project, err)
		}
		if len(tags) == 0 {
			return nil, fmt.Errorf("%s has no tags", d.Project)
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
