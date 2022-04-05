package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		cfg, err := parseConfig(strings.NewReader(`
			last-access 24h
			threshold 15%
			interval 5m
			quiet-period 4h
		`))
		require.NoError(t, err)
		assert.Empty(t, cfg.cacheRoot)
		assert.Equal(t, 24*time.Hour, cfg.lastAccess)
		assert.Equal(t, uint64(15), cfg.threshold)
		assert.Equal(t, 5*time.Minute, cfg.interval)
		assert.Equal(t, 4*time.Hour, cfg.quietPeriod)
	})

	t.Run("default quite-interval", func(t *testing.T) {
		cfg, err := parseConfig(strings.NewReader("last-access 24h"))
		require.NoError(t, err)
		assert.Equal(t, 6*time.Hour, cfg.quietPeriod)
	})

	t.Run("Empty arguments", func(t *testing.T) {
		_, err := parseConfig(strings.NewReader(`
			last-access 24h
			quiet-period
		`))
		assert.Error(t, err, "quiet-period: time: invalid duration")
	})
}

func TestParseCacheRoot(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		cacheRoot, err := parseCacheRoot(strings.NewReader(`
			tag mycache
			dir /var/cache/fscache
		`))
		assert.NoError(t, err)
		assert.Equal(t, "/var/cache/fscache", cacheRoot)
	})

	t.Run("missing dir", func(t *testing.T) {
		cacheRoot, err := parseCacheRoot(strings.NewReader(""))
		assert.Error(t, err, "missing dir command")
		assert.Equal(t, "", cacheRoot)
	})
}

func TestSplitLine(t *testing.T) {
	type test struct {
		line string
		cmd  string
		arg  string
	}

	tests := []test{
		{"", "", ""},
		{"foo", "foo", ""},
		{"foo bar", "foo", "bar"},
		{"foo bar baz", "foo", "bar baz"},

		{" ", "", ""},
		{" foo ", "foo", ""},
		{" foo bar ", "foo", "bar"},

		{"#foo", "", ""},
		{"  #foo", "", ""},
		{"#foo bar", "", ""},
		{"foo #bar", "foo", ""},
		{"foo#bar", "foo", ""},
		{"foo bar #baz", "foo", "bar"},
	}

	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			cmd, arg := splitLine(tc.line)
			assert.Equal(t, tc.cmd, cmd)
			assert.Equal(t, tc.arg, arg)
		})
	}
}
