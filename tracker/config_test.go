package tracker

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParseRegex(t *testing.T) {
	input := `"^$"`
	expect := (*Regexp)(regexp.MustCompile("^$"))

	var actual Regexp
	err := yaml.Unmarshal([]byte(input), &actual)
	require.NoError(t, err)
	require.Equal(t, expect, &actual)
}
