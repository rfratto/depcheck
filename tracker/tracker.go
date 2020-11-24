package tracker

import "context"

// Tracker can return a list of outdated dependencies.
type Tracker interface {
	// CheckOutdated should return a list of only outdated dependencies.
	CheckOutdated(ctx context.Context) ([]Dependency, error)
}

// Dependency is a named dependency with the current version being used
// and the latest version available.
type Dependency struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
}

// New creates a new Tracker that can return outdated dependencies.
func New(c *Config, repo string) Tracker {
	var trackers []Tracker
	if len(c.GoModules) > 0 {
		trackers = append(trackers, NewGoModules(repo, c.GoModules))
	}
	if len(c.GithubDeps) > 0 {
		trackers = append(trackers, NewGithub(c.GithubDeps))
	}
	return &Multi{trackers: trackers}
}

// Multi combines multiple trackers.
type Multi struct {
	trackers []Tracker
}

// CheckOutdated calls CheckOutdated for each tracker in the list.
func (m *Multi) CheckOutdated(ctx context.Context) ([]Dependency, error) {
	var deps []Dependency

	for _, t := range m.trackers {
		tDeps, err := t.CheckOutdated(ctx)
		if err != nil {
			return nil, err
		}
		deps = append(deps, tDeps...)
	}

	return deps, nil
}
