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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type filter struct {
	includes []string
	excludes []string
	input    io.Reader
	output   io.Writer
	field    int
}

func (f *filter) run() error {
	// no filters, stream the input directly to the output
	if len(f.excludes) == 0 && len(f.includes) == 0 {
		_, err := io.Copy(f.output, f.input)
		return err
	}

	s := bufio.NewScanner(f.input)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}

		export := line
		if f.field > 0 {
			var err error
			export, err = extractField(line, f.field-1)
			if err != nil {
				return err
			}
		}

		if !strings.HasSuffix(export, "/") {
			export += "/"
		}

		if !match(export, f.includes, true) {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Skipped \"%s\", did not match include filter", export)
			}
			continue
		}

		if match(export, f.excludes, false) {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Skipped \"%s\", export was excluded", export)
			}
			continue
		}

		fmt.Fprintln(f.output, line)
	}

	return s.Err()
}

func match(export string, patterns []string, empty bool) bool {
	if len(patterns) == 0 {
		return empty
	}

	for _, p := range patterns {
		m, err := doublestar.Match(p, export)
		if err != nil {
			// this should not happen as the patterns were checked when loading
			panic(err)
		}
		if m {
			return true
		}
	}

	return false
}

func extractField(line string, index int) (string, error) {
	fields := strings.Fields(line)
	if index >= len(fields) {
		return "", fmt.Errorf("could not extract field %d from '%s'", index, line)
	}
	return fields[index], nil
}
