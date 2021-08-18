package tracker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/go-github/v32/github"
)

var ErrIssueNotFound = errors.New("no such issue")

// IssueCreator can create issues for a set of dependencies.
type IssueCreator struct {
	c   *Config
	cli *github.Client

	titleTmpl, bodyTmpl *template.Template
}

// NewIssueCreator creates a new issue creator.
func NewIssueCreator(c *Config, cli *github.Client) (*IssueCreator, error) {
	titleTmpl, err := template.New("issue_title").Parse(c.IssueTitleTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue title template: %w", err)
	}
	bodyTmpl, err := template.New("issue_body").Parse(c.IssueTextTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue body template: %w", err)
	}

	return &IssueCreator{
		c:   c,
		cli: cli,

		titleTmpl: titleTmpl,
		bodyTmpl:  bodyTmpl,
	}, nil
}

// Create will create issues for a dep. If an issue already exists
// (including if it is closed), then it will not be recreated. The associated
// issue is returned.
func (c *IssueCreator) CreateIssue(ctx context.Context, dep Dependency) (*github.Issue, error) {
	iss, err := c.FindIssue(ctx, dep)
	if err != nil && err != ErrIssueNotFound {
		return nil, fmt.Errorf("failed to look for existing issue: %w", err)
	} else if err == nil {
		return iss, nil
	}

	repoOwner, repoName, err := parseGithubRepo(c.c.IssueRepository)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse issue repo: %w", err)
	}

	expectedIssue, err := c.issueRequest(dep)
	if err != nil {
		return nil, fmt.Errorf("failed to generate expected issue: %w", err)
	}

	iss, _, err = c.cli.Issues.Create(ctx, repoOwner, repoName, expectedIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	return iss, nil
}

// FindIssue looks for an existing issue associated with a dependency.
func (c *IssueCreator) FindIssue(ctx context.Context, dep Dependency) (*github.Issue, error) {
	expectedIssue, err := c.issueRequest(dep)
	if err != nil {
		return nil, fmt.Errorf("failed to generate expected issue: %w", err)
	}

	query := fmt.Sprintf(
		`"%s" repo:"%s" label:"%s" in:title`,
		expectedIssue.GetTitle(),
		c.c.IssueRepository,
		c.c.OutdatedLabel,
	)
	res, _, err := c.cli.Search.Issues(ctx, query, &github.SearchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to search for issues: %w", err)
	}
	for _, iss := range res.Issues {
		if iss.GetTitle() == expectedIssue.GetTitle() {
			return iss, nil
		}
	}

	return nil, ErrIssueNotFound
}

func (c *IssueCreator) issueRequest(dep Dependency) (*github.IssueRequest, error) {
	var (
		titleBuilder strings.Builder
		bodyBuilder  strings.Builder
	)

	err := c.titleTmpl.Execute(&titleBuilder, dep)
	if err != nil {
		return nil, fmt.Errorf("failed to generate issue title: %w", err)
	}

	err = c.bodyTmpl.Execute(&bodyBuilder, dep)
	if err != nil {
		return nil, fmt.Errorf("failed to generate issue body: %w", err)
	}

	var (
		title = titleBuilder.String()
		body  = bodyBuilder.String()
	)

	return &github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &[]string{c.c.OutdatedLabel},
	}, nil
}

func parseGithubRepo(fullRepo string) (owner, repo string, err error) {
	parts := strings.SplitN(fullRepo, "/", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid repo format %s, expected owner/repo", fullRepo)
		return
	}
	return parts[0], parts[1], nil
}

// CloseOutdated closes issues for dep that are older than latest.
func (c *IssueCreator) CloseOutdated(ctx context.Context, latest *github.Issue, dep Dependency) error {
	genericDep := dep
	genericDep.LatestVersion = "*"

	genericIss, err := c.issueRequest(genericDep)
	if err != nil {
		return fmt.Errorf("failed to generate oudated issue pattern: %w", err)
	}

	repoOwner, repoName, err := parseGithubRepo(c.c.IssueRepository)
	if err != nil {
		return fmt.Errorf("couldn't parse issue repo: %w", err)
	}

	searchQuery := fmt.Sprintf(
		`"%s" repo:"%s" label:"%s" in:title is:open`,
		genericIss.GetTitle(),
		c.c.IssueRepository,
		c.c.OutdatedLabel,
	)

	res, _, err := c.cli.Search.Issues(ctx, searchQuery, &github.SearchOptions{})
	if err != nil {
		return fmt.Errorf("failed to search for issues: %w", err)
	}
	for _, iss := range res.Issues {
		if iss.GetID() == latest.GetID() || iss.GetID() == 0 {
			continue
		}

		var (
			closed  = "closed"
			comment = fmt.Sprintf("Closing in favor of #%d", latest.GetNumber())
		)

		_, _, err := c.cli.Issues.CreateComment(ctx, repoOwner, repoName, iss.GetNumber(), &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			return fmt.Errorf("failed to make comment on #%d: %w", iss.GetNumber(), err)
		}

		_, _, err = c.cli.Issues.Edit(ctx, repoOwner, repoName, iss.GetNumber(), &github.IssueRequest{
			State: &closed,
		})
		if err != nil {
			return fmt.Errorf("failed to close outdated issue #%d: %w", iss.GetNumber(), err)
		}
	}

	return nil
}
