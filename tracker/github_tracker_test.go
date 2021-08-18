package tracker

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParseGithubDependency(t *testing.T) {
	tt := []struct {
		input  string
		expect GithubDependency
	}{
		{
			input: `"github.com/prometheus/prometheus v1.2.3"`,
			expect: GithubDependency{
				Project: "github.com/prometheus/prometheus",
				Version: "v1.2.3",
			},
		},
		{
			input: `{
				"project": "github.com/prometheus/prometheus",
				"version": "v1.2.3",
				"ignore_version_pattern": "foo",
			}`,
			expect: GithubDependency{
				Project: "github.com/prometheus/prometheus",
				Version: "v1.2.3",
				Options: DependencyOptions{
					IgnoreVersionPattern: (*Regexp)(regexp.MustCompile("foo")),
				},
			},
		},
	}

	for _, tc := range tt {
		var actual GithubDependency
		err := yaml.Unmarshal([]byte(tc.input), &actual)
		require.NoError(t, err)
		require.Equal(t, tc.expect, actual)
	}
}
