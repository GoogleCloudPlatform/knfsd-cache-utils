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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	excludeFile = flag.String("exclude", "", "`path` to a file that contains a list of exclude patterns")
	includeFile = flag.String("include", "", "`path` to a file that contains a list of include patterns")
	field       = flag.Int("field", 0, "sets the field (space delimited) that contains the export")
	verbose     = flag.Bool("verbose", false, "log rejected exports to stderr")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	var err error
	var f = &filter{
		input:  os.Stdin,
		output: os.Stdout,
		field:  *field,
	}

	f.excludes, err = loadPatterns(*excludeFile)
	fatal(err)

	f.includes, err = loadPatterns(*includeFile)
	fatal(err)

	err = f.run()
	fatal(err)
}

func fatal(err error) {
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
}

func loadPatterns(path string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	patterns := make([]string, 0, 10)

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		line, err = check(line)
		if err != nil {
			return nil, err
		}

		patterns = append(patterns, line)
	}

	if err = s.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

func check(pattern string) (string, error) {
	if !strings.HasPrefix(pattern, "/") {
		return "", fmt.Errorf("invalid pattern '%s': must start with '/'", pattern)
	}
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}

	_, err := doublestar.Match(pattern, "")
	if err != nil {
		return "", fmt.Errorf("invalid pattern '%s': %w", pattern, err)
	}

	return pattern, nil
}
