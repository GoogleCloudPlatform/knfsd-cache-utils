package main

import (
	"fmt"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/stretchr/testify/assert"
)

type testPattern struct {
	expected bool
	pattern  string
	input    string
}

// These are not exhaustive tests as the doublestar package has its own tests.
// These test just assert that the doublestar.Match has the behaviour we want.
var testPatterns = []testPattern{
	{true, "/", "/"},
	{false, "/", "/a"},

	// Handling of trailing slashes on exports.
	// Note that /* does not match if there is a trailing slash.
	// Need to trim trailing slashes from the list of exports before filtering.
	{true, "/foo", "/foo"},
	{false, "/foo", "/foo/"},
	{true, "/*", "/foo"},
	{false, "/*", "/foo/"},

	// Handling of trailing slashes on patterns.
	// Patterns *must not* have trailing slashes, otherwise they'll never match
	// an export.
	// Need to trim trailing slashes from patterns before filtering.
	{false, "/foo/", "/foo"},
	{true, "/foo/", "/foo/"},
	{false, "/*/", "/foo"},
	{true, "/*/", "/foo/"},

	// Handling of the parents when the last component is a wildcard.
	{false, "/foo/*", "/foo"},
	{true, "/foo/*", "/foo/bar"},
	// {false, "/foo/**", "/foo"},
	{true, "/foo/**", "/foo/bar"},
	{true, "/foo/**", "/foo/bar/baz"},

	// wildcard as the root component
	{true, "/*", "/"},
	{true, "/*", "/a"},
	{false, "/*", "/a/"},
	{false, "/*", "/a/b"},

	// wildcard as the last component
	{false, "/foo/*", "/foo"},
	// {false, "/foo/*", "/foo/"},
	{true, "/foo/*", "/foo/a"},
	{false, "/foo/*", "/foo/a/"},
	{false, "/foo/*", "/foo/a/b"},
	{false, "/foo/*", "/bar"},
	{false, "/foo/*", "/bar/"},
	{false, "/foo/*", "/bar/a"},

	// wildcard as a suffix (last component)
	{true, "/foo-*", "/foo-x"},
	{false, "/foo-*", "/foo-x/"},
	{true, "/foo-*", "/foo-"},
	{false, "/foo-*", "/foo-/x"},
	{false, "/foo-*", "/foo-x/bar"},

	// wildcard as a component suffix
	{true, "/foo-*/bar", "/foo-x/bar"},
	{false, "/foo-*/bar", "/foo-x/bar/"},
	{true, "/foo-*/bar", "/foo-y/bar"},
	{false, "/foo-*/bar", "/foo-x/baz"},
	{false, "/foo-*/bar", "/foo-x/baz/bar"},
	{false, "/foo-*/bar", "/foo-x/bar/baz"},

	// wildcard as a component
	{true, "/*/bar", "/x/bar"},
	// {true, "/*/bar", "/x/bar/"},
	{true, "/*/bar", "/y/bar"},
	{false, "/*/bar", "/x/baz"},
	{false, "/*/bar", "/x/baz/bar"},
	{false, "/*/bar", "/x/bar/baz"},

	// recursive wildcard
	{true, "/foo/**", "/foo/bar"},
	{true, "/foo/**", "/foo/bar/baz"},
	{true, "/foo/**", "/foo"},
	{true, "/foo/**", "/foo/"},
	{false, "/foo/**", "/bar"},
	{false, "/foo/**", "/bar/baz"},
	{false, "/foo/**", "/bar/foo"},
	{false, "/foo/**", "/bar/foo/baz"},

	// pattern missing root /
	// Because exports are absolute paths these should never match.
	{false, "foo", "/foo"},
	{false, "*", "/foo"},
	// ** is a special case as it matches everything, this is equivalent to /**
	{true, "**", "/foo"},
}

func testSinglePattern(t *testing.T, tc testPattern) {
	actual, err := doublestar.Match(tc.pattern, tc.input)
	assert.NoError(t, err)

	if tc.expected && !actual {
		assert.Fail(t, fmt.Sprintf("Expected pattern to accept:\n"+
			"pattern: %s\n"+
			"name   : %s", tc.pattern, tc.input))

	}
	if !tc.expected && actual {
		assert.Fail(t, fmt.Sprintf("Expected pattern to reject:\n"+
			"pattern: %s\n"+
			"name   : %s", tc.pattern, tc.input))
	}
}

func TestSinglePatterns(t *testing.T) {
	for _, tc := range testPatterns {
		t.Run("", func(t *testing.T) {
			testSinglePattern(t, tc)
		})
	}
}
