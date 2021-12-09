package opt

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testOpts struct {
	args        []string
	environment map[string]string
}

func makeopts(opts ...string) testOpts {
	return testOpts{
		args:        opts,
		environment: make(map[string]string, 10),
	}
}

func (args testOpts) opt(a ...string) testOpts {
	args.args = append(args.args, a...)
	return args
}

func (args testOpts) env(k, v string) testOpts {
	args.environment[k] = v
	return args
}

func (arg testOpts) lookup(name string) (string, bool) {
	val, found := arg.environment[name]
	return val, found
}

func TestOptsBool(t *testing.T) {
	type test struct {
		name     string
		input    testOpts
		expected bool
	}

	tests := []test{
		{"A", makeopts(), false},

		{"B", makeopts("-x"), true},
		{"C", makeopts("-x=true"), true},
		{"D", makeopts("-x=false"), false},

		{"E", makeopts().env("X", ""), false},
		{"F", makeopts().env("X", "true"), true},
		{"G", makeopts().env("X", "false"), false},
		// {"H", makeargs().env("X", "xyzzy"), false},

		{"I", makeopts("-x").env("X", "false"), true},
		{"J", makeopts("-x=true").env("X", "false"), true},
		{"K", makeopts("-x=false").env("X", "true"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("", flag.ContinueOnError)
			args := NewOptSet(flags)
			args.LookupEnv = tc.input.lookup

			var x bool
			args.BoolVar(&x, "x", "X", "")

			err := args.Parse(tc.input.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, x)
		})
	}
}

func TestOptsString(t *testing.T) {
	type test struct {
		name     string
		input    testOpts
		expected string
	}

	tests := []test{
		{"A", makeopts(), ""},
		{"B", makeopts("-x", "a"), "a"},
		{"E", makeopts().env("X", ""), ""},
		{"F", makeopts().env("X", "b"), "b"},
		{"I", makeopts("-x", "a").env("X", "b"), "a"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("", flag.ContinueOnError)
			args := NewOptSet(flags)
			args.LookupEnv = tc.input.lookup

			var x string
			args.StringVar(&x, "x", "X", "")

			err := args.Parse(tc.input.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, x)
		})
	}
}

func TestOptsError(t *testing.T) {
	input := makeopts().env("X", "bob")
	var x bool

	flags := flag.NewFlagSet("", flag.ContinueOnError)
	args := NewOptSet(flags)
	args.LookupEnv = input.lookup
	args.BoolVar(&x, "x", "X", "")

	err := args.Parse(input.args)
	assert.EqualError(t, err, "invalid value \"bob\" for X: parse error")
}
