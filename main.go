package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/actions-go/toolkit/core"
	"github.com/rfratto/depcheck/tracker"
)

func main() {
	var (
		repoPath   string
		configPath string
		dryRun     bool
	)

	f := flag.NewFlagSet("dependency-tracker", flag.ExitOnError)
	f.StringVar(&repoPath, "repository", ".", "repository to check dependencies for")
	f.StringVar(&configPath, "config-path", ".github/dependencies.yml", "config file for the dependency tracker")
	f.BoolVar(&dryRun, "dry-run", false, "don't actually create the issues")

	// Load in values that may be passed in via GitHub. This should be done
	// *after* declaring the flags (which may define defaults) but *before*
	// parsing the flags (which may override these values).
	repoPath = core.GetInputOrDefault("repository", repoPath)
	configPath = core.GetInputOrDefault("config_path", configPath)
	dryRun = core.GetBoolInput("dry_run")

	if err := f.Parse(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}

	cfg, err := tracker.LoadConfig(filepath.Join(repoPath, configPath))
	if err != nil {
		log.Fatalln(err)
	}

	t := tracker.New(cfg, repoPath)
	deps, err := t.CheckOutdated(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	if len(deps) == 0 {
		return
	}

	fmt.Printf("Out of date dependencies:\n\n")

	for _, dep := range deps {
		fmt.Printf("\tName:      %s\n", dep.Name)
		fmt.Printf("\tVersion:   %s\n", dep.CurrentVersion)
		fmt.Printf("\tAvailable: %s\n", dep.LatestVersion)
		fmt.Println()
	}
}
