package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type filterTest struct {
	inputs   string
	excludes string
	includes string
	expected string
	field    int
}

func (ft filterTest) run(t *testing.T) {
	var err error
	var f filter

	expected, err := os.ReadFile(ft.expected)
	require.NoError(t, err)

	f.includes, err = loadPatterns(ft.includes)
	require.NoError(t, err)

	f.excludes, err = loadPatterns(ft.excludes)
	require.NoError(t, err)

	input, err := os.Open(ft.inputs)
	require.NoError(t, err)
	defer input.Close()
	f.input = input

	output := &strings.Builder{}
	f.output = output
	f.field = ft.field

	err = f.run()
	require.NoError(t, err)
	assert.Equal(t, string(expected), output.String())
}

func TestEmpty(t *testing.T) {
	ft := filterTest{
		inputs:   "testdata/simple/inputs",
		includes: "",
		excludes: "",
		expected: "testdata/simple/inputs",
	}
	ft.run(t)
}

func TestIncludes(t *testing.T) {
	ft := filterTest{
		inputs:   "testdata/simple/inputs",
		includes: "testdata/simple/patterns",
		excludes: "",
		expected: "testdata/simple/expected-include",
	}
	ft.run(t)
}

func TestExcludes(t *testing.T) {
	ft := filterTest{
		inputs:   "testdata/simple/inputs",
		includes: "",
		excludes: "testdata/simple/patterns",
		expected: "testdata/simple/expected-exclude",
	}
	ft.run(t)
}

func TestExcludeParent(t *testing.T) {
	ft := filterTest{
		inputs:   "testdata/exclude-parent/inputs",
		includes: "",
		excludes: "testdata/exclude-parent/excludes",
		expected: "testdata/exclude-parent/expected",
	}
	ft.run(t)
}

func TestFields(t *testing.T) {
	ft := filterTest{
		inputs:   "testdata/fields/inputs",
		includes: "testdata/fields/includes",
		excludes: "",
		expected: "testdata/fields/expected",
		field:    2,
	}
	ft.run(t)
}

func TestCombined(t *testing.T) {
	ft := filterTest{
		inputs:   "testdata/combined/inputs",
		includes: "testdata/combined/includes",
		excludes: "testdata/combined/excludes",
		expected: "testdata/combined/expected",
	}
	ft.run(t)
}
