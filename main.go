package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/actions-go/toolkit/core"
	"github.com/google/go-github/v32/github"
	"github.com/rfratto/depcheck/tracker"
	"golang.org/x/oauth2"
)

func main() {
	var (
		repoPath      string
		configPath    string
		githubToken   string
		dryRun        bool
		closeOutdated bool
	)

	f := flag.NewFlagSet("dependency-tracker", flag.ExitOnError)
	f.StringVar(&repoPath, "repository", ".", "repository to check dependencies for")
	f.StringVar(&configPath, "config-path", ".github/depcheck.yml", "config file for the dependency tracker")
	f.StringVar(&githubToken, "github-token", "", "github token to use")
	f.BoolVar(&dryRun, "dry-run", false, "don't actually create the issues")
	f.BoolVar(&closeOutdated, "close-oudated", true, "close oudated issues after creating a new one")

	// Load in values that may be passed in via GitHub. This should be done
	// *after* declaring the flags (which may define defaults) but *before*
	// parsing the flags (which may override these values).
	repoPath = core.GetInputOrDefault("repository", repoPath)
	configPath = core.GetInputOrDefault("config-path", configPath)
	dryRun = boolOrDefault("dry-run", dryRun)
	closeOutdated = boolOrDefault("close-oudated", closeOutdated)
	githubToken = getGithubToken()

	if err := f.Parse(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(context.Background(), ts)
	cli := github.NewClient(tc)

	cfg, err := tracker.LoadConfig(filepath.Join(repoPath, configPath))
	if err != nil {
		log.Fatalln(err)
	}

	t := tracker.New(cfg, repoPath, cli)
	deps, err := t.CheckOutdated(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	if len(deps) == 0 {
		return
	}

	fmt.Printf("Issues will be created in %s\n", cfg.IssueRepository)
	fmt.Printf("Out of date dependencies:\n\n")

	for _, dep := range deps {
		fmt.Printf("\tName:      %s\n", dep.Name)
		fmt.Printf("\tVersion:   %s\n", dep.CurrentVersion)
		fmt.Printf("\tAvailable: %s\n", dep.LatestVersion)
		fmt.Println()
	}

	if dryRun {
		return
	}

	creator, err := tracker.NewIssueCreator(cfg, cli)
	if err != nil {
		log.Fatalln(err)
	}

	for _, dep := range deps {
		iss, err := creator.CreateIssue(context.Background(), dep)
		if err != nil {
			log.Printf("failed to create issue for %s: %s", dep.Name, err)
			continue
		}

		if !closeOutdated {
			continue
		}
		err = creator.CloseOutdated(context.Background(), iss, dep)
		if err != nil {
			log.Printf("failed to closed outdated issues for %s: %s", dep.Name, err)
			continue
		}
	}
}

// Taken from
// https://github.com/actions-go/toolkit/blob/2e1e0898191c8feac91a13ea9acaf06c811fcaf4/github/github.go#L20
func getGithubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	for _, input := range []string{"github-token", "token"} {
		if t, ok := core.GetInput(input); ok {
			return t
		}
	}
	return ""
}

func boolOrDefault(name string, defaultValue bool) bool {
	defaultStr := "false"
	if defaultValue {
		defaultStr = "true"
	}
	return strings.ToLower(core.GetInputOrDefault(name, defaultStr)) == "true"
}
