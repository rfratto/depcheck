package tracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// GoModules checks for outdated dependencies for Go modules.
type GoModules struct {
	module string
	check  []GoModule
}

// GoModule is an individual module to check.
type GoModule struct {
	Name    string            `yaml:"name"`
	Options DependencyOptions `yaml:",inline"`
}

// UnmarshalYAML unmarshals a GoModule. The value can either be a string
// or the GoModule struct.
func (m *GoModule) UnmarshalYAML(f func(v interface{}) error) error {
	var (
		stringError error
		objectError error
	)

	// Try as a raw string
	var s string
	stringError = f(&s)
	if stringError == nil {
		m.Name = s
		return nil
	}

	// Then a whole object
	type goModule GoModule
	var v goModule
	objectError = f(&v)
	if objectError == nil {
		*m = GoModule(v)
		return nil
	}

	return fmt.Errorf(
		"could not parse go module as a string (%s) or an object (%s)",
		stringError,
		objectError,
	)
}

// NewGoModules creates a new GoModules tracker.
func NewGoModules(module string, check []GoModule) *GoModules {
	return &GoModules{
		module: module,
		check:  check,
	}
}

// CheckOutdated will return the list of go module dependencies that can be updated.
func (c *GoModules) CheckOutdated(ctx context.Context) ([]Dependency, error) {
	var (
		moduleNames = make([]string, len(c.check))
		moduleMap   = make(map[string]GoModule)
	)
	for i := 0; i < len(c.check); i++ {
		moduleNames[i] = c.check[i].Name
		moduleMap[c.check[i].Name] = c.check[i]
	}

	var outdated []Dependency
	goArgs := append([]string{"list", "-mod=readonly", "-json", "-u", "-m"}, moduleNames...)

	var out bytes.Buffer

	cmd := exec.Command("go", goArgs...)
	cmd.Stdout = &out
	cmd.Dir = c.module

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed running go command: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(out.Bytes()))
	for {
		var dep goListModule
		if err := dec.Decode(&dep); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error decoding dependency: %w", err)
		}

		if dep.Error != nil {
			return nil, fmt.Errorf("error loading module %s: %w", dep.Path, dep.Error)
		}

		// No update available, skip it
		if dep.Update == nil {
			continue
		}

		// Check to see if we're ignoring this version
		if ref, ok := moduleMap[dep.Path]; ok {
			if ref.Options.IgnoreVersionPattern.Matches(dep.Update.Version) {
				continue
			}
		}

		outdated = append(outdated, Dependency{
			Name:           dep.Path,
			CurrentVersion: dep.Version,
			LatestVersion:  dep.Update.Version,
		})
	}

	return outdated, nil
}

// goListModule is the structure returned by "go list -json -u -m"
// This struct was copied from the help page for "go help list".
type goListModule struct {
	Path      string             // module path
	Version   string             // module version
	Versions  []string           // available module versions (with -versions)
	Replace   *goListModule      // replaced by this module
	Time      *time.Time         // time version was created
	Update    *goListModule      // available update, if any (with -u)
	Main      bool               // is this the main module?
	Indirect  bool               // is this module only an indirect dependency of main module?
	Dir       string             // directory holding files for this module, if any
	GoMod     string             // path to go.mod file used when loading this module, if any
	GoVersion string             // go version used in module
	Error     *goListModuleError // error loading module
}

// This struct was copied from the help page for "go help list".
type goListModuleError struct {
	Err string // the error itself
}

func (e *goListModuleError) Error() string { return e.Err }
