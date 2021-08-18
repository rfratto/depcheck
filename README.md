# depcheck

`depcheck` is a tool that can analyze a repository's Go and Git dependencies and
create GitHub issues for dependencies that are out of date. If you want to opt
out of updating a dependency to a new version, simply close the issue - the
issue for that dependency version won't be created again.

### Why not Dependabot?

[Dependabot](https://dependabot.com/) is extremely cool and infinitely more
powerful than `depcheck`. Depcheck was designed to meet the following needs:

1. Tracking dependencies that may be implicit (e.g., you depend on a repo but
   only use release assets, or you have a fork with changes from an upstream
   repo and want to be notified when the upstream repo changes)

2. Your dependencies frequently make breaking API changes and would find issues
   more useful than pull requests.

So unless you have these very specific and niche needs, you should try
Dependabot instead!

## Configuring

`depcheck` looks for a `.github/depcheck.yml` file from the current working
directory. Another directory can be specifed by passing the `-repository` flag.
A different relative file path other than `.github/depcheck.yml` can be
specified with the `-config-path` flag.

The `depcheck` YAML config file takes these options:

```yaml
# List of go modules to check for outdated-ness. Prometheus is listed as an
# example; all modules listed here must be used in your go.mod.
#
# You can use a regexp to ignore specific versions by passing an object
# as a dependency instead of a string.
go_modules:
  - github.com/grafana/agent
  - name: github.com/prometheus/prometheus
    ignore_version_pattern: "-rc\.\d+$" # Ignore release candidates

# List of Github repos to check for newer tags. The versions here must be listed
# and updated manually; there is no magic to determine what is being used. This
# is a fallback mechanism for checking dependencies that influence the project
# and aren't directly imported as go modules.
#
# "github.com/" may be omitted as a prefix, but the dependency name listed in
# issues will always display as prefixed with github.com/.
#
# You can use a regexp to ignore specific versions by passing an object
# as a dependency instead of a string.
github_repos:
  - github.com/rfratto/depcheck v0.1.0
  - project: github.com/prometheus/node_exporter
    version: v0.18.1
    ignore_version_pattern: "-rc\.\d+$" # Ignore release candidates

# Repository to create issues in. If empty or undefined, defaults to
# GITHUB_REPOSITORY in environment variables. Must be set either here or via the
# environment variable.
issue_repository: ''

# Label to use for tracking outdated dependencies.
outdated_label: 'outdated-dependency'

# Title of the issue to create. This title is searched for when creating a new
# issue to determine if one already exists. Uses Go's text/template to render
# out the string. .Name, .LatestVersion, and .CurrentVersion are all available
# as fields to use.
issue_title_template: |-
  Update {{.Name}} to {{.LatestVersion}}

# Body of the issue to create. Uses Go's text/template to render out the
# string. .Name, .LatestVersion, and .CurrentVersion are all available
# as fields to use.
issue_text_template: >-
  An update for `{{.Name}}` (version `{{.LatestVersion}}`) is now available.
  Version `{{.CurrentVersion}}` is currently in use.
```

## Using

`depcheck` can function as a GitHub Action. Example usage to check dependency
whenever a workflow is manually triggered:

```yaml
name: Check Dependencies
on:
  workflow_dispatch: {}
jobs:
  check:
    name: Check
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v2

    - name: Invoke action
      uses: rfratto/depcheck@main
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
```

The following inputs are available:

- `repository` coressponds to the `-repository` flag.
- `config-path` coressponds to the `-config-path` flag.
- `dry-run` coressponds to the `-dry-run` flag and will stop at printing out the
   outdated dependencies and not actually create any issues.
- `github-token` corresponds to the `-github-token` flag.
- `close-outdated` corresponds to the `-close-oudated` flag.

## Roadmap

- [ ] Jsonnet dependencies

