package tracker

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParseGoModule(t *testing.T) {
	tt := []struct {
		input  string
		expect GoModule
	}{
		{
			input: `"github.com/prometheus/prometheus"`,
			expect: GoModule{
				Name: "github.com/prometheus/prometheus",
			},
		},
		{
			input: `{
				"name": "github.com/prometheus/prometheus",
				"ignore_version_pattern": "foo",
			}`,
			expect: GoModule{
				Name: "github.com/prometheus/prometheus",
				Options: DependencyOptions{
					IgnoreVersionPattern: (*Regexp)(regexp.MustCompile("foo")),
				},
			},
		},
	}

	for _, tc := range tt {
		var actual GoModule
		err := yaml.Unmarshal([]byte(tc.input), &actual)
		require.NoError(t, err)
		require.Equal(t, tc.expect, actual)
	}
}
