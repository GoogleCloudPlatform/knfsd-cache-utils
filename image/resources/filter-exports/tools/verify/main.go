/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"math"

	"github.com/bmatcuk/doublestar/v4"
)

type test struct {
	shouldMatch bool
	pattern     string
	input       string
}

// Define patterns and inputs without a trailing slash.
// We will test them in all 4 combinations of trailing slash to see which set of
// combinations is closest to our desired output.
var patternsToSearch = []test{
	// exact match
	{true, "/foo", "/foo"},
	{false, "/foo", "/foo/bar"},
	{false, "/foo", "/bar"},

	// wildcard, last component
	{false, "/foo/*", "/foo"},
	{true, "/foo/*", "/foo/bar"},
	{false, "/foo/*", "/bar"},
	{false, "/foo/*", "/bar/baz"},

	// wildcard, component
	{true, "/*/bar", "/foo/bar"},
	{false, "/*/bar", "/foo/baz"},
	{false, "/*/bar", "/foo/bar/baz"},

	// recursive
	{false, "/foo/**", "/foo"},
	{true, "/foo/**", "/foo/bar"},
	{true, "/foo/**", "/foo/bar/baz"},
	{false, "/foo/**", "/bar"},
	{false, "/foo/**", "/bar/baz"},

	// wildcard + recursive
	{false, "/*/bar/**", "/foo"},
	{false, "/*/bar/**", "/foo/bar"},
	{true, "/*/bar/**", "/foo/bar/baz"},
	{true, "/*/bar/**", "/foo/bar/baz/more"},

	// pattern missing anchor
	{false, "foo", "/foo"},
	{false, "*", "/foo"},
	{true, "**", "/foo"},  // exception
	{true, "/**", "/foo"}, // **/ is equivalent to /**/
}

func main() {
	failed := make([][]test, 4)
	failed[0] = try(false, false)
	failed[1] = try(true, false)
	failed[2] = try(false, true)
	failed[3] = try(true, true)

	min := math.MaxInt
	for _, f := range failed {
		if len(f) < min {
			min = len(f)
		}
	}

	for i, f := range failed {
		fmt.Printf("Case %d has %d errors\n", i, len(f))
	}

	for i, f := range failed {
		if len(f) > min {
			continue
		}

		fmt.Println()
		fmt.Printf("Case %d's errors:\n", i)
		for _, t := range f {
			fmt.Printf("pattern: %-20s input: %-20s", t.pattern, t.input)
		}
	}
}

func try(patternSlash, inputSlash bool) (failed []test) {
	for _, t := range patternsToSearch {
		pattern := t.pattern
		if patternSlash {
			pattern += "/"
		}

		input := t.input
		if inputSlash {
			input += "/"
		}

		matched, err := doublestar.Match(pattern, input)
		if err != nil {
			log.Fatalf("pattern '%s' is not valid: %s\n", pattern, err)
		}

		if matched != t.shouldMatch {
			failed = append(failed, t)
		}
	}
	return
}
