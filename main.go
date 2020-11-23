package main

import (
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
	)

	f := flag.NewFlagSet("dependency-tracker", flag.ExitOnError)
	f.StringVar(&repoPath, "repository", ".", "repository to check dependencies for")
	f.StringVar(&configPath, "config-path", ".github/dependencies.yml", "config file for the dependency tracker")

	// Load in values that may be passed in via GitHub. This should be done
	// *after* declaring the flags (which may define defaults) but *before*
	// parsing the flags (which may override these values).
	repoPath = core.GetInputOrDefault("repository", repoPath)
	configPath = core.GetInputOrDefault("config_path", configPath)

	if err := f.Parse(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}

	cfg, err := tracker.LoadConfig(filepath.Join(repoPath, configPath))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Label: %v\n", cfg.OutdatedLabel)
	fmt.Printf("Go Modules: %v\n", cfg.GoModules)
	fmt.Printf("Repos: %v\n", cfg.Repos)
}
