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
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func readConfig(name string) (config, error) {
	f, err := os.Open(name)
	if err != nil {
		return config{}, err
	}
	defer f.Close()
	return parseConfig(f)
}

func parseConfig(r io.Reader) (config, error) {
	var err error

	cfg := config{
		lastAccess: 1 * time.Hour,
		threshold:  20,
		interval:   1 * time.Minute,
		// calculate default for quite period after parsing the config
	}

	s := bufio.NewScanner(r)
	for s.Scan() {
		cmd, arg := splitLine(s.Text())
		if cmd == "" {
			continue
		}

		switch cmd {
		case "last-access":
			cfg.lastAccess, err = time.ParseDuration(arg)

		case "threshold":
			// allow threshold to end with an optional % to match cachefilesd.conf
			cfg.threshold, err = parsePercentage(arg)

		case "interval":
			cfg.interval, err = time.ParseDuration(arg)

		case "quiet-period":
			cfg.quietPeriod, err = time.ParseDuration(arg)

		default:
			err = errors.New("unknown command")
		}

		if err != nil {
			err = fmt.Errorf("%s: %w", cmd, err)
			return cfg, err
		}
	}

	err = s.Err()
	if err != nil {
		return cfg, err
	}

	if cfg.lastAccess < 0 {
		return cfg, errors.New("age cannot be less than 0")
	}

	if cfg.threshold < 0 || cfg.threshold > 100 {
		return cfg, errors.New("threshold must be between 0% and 100%")
	}

	cfg.interval = cfg.interval.Truncate(time.Second)
	if cfg.interval < 1*time.Second {
		return cfg, errors.New("interval cannot be less than 1 second")
	}

	cfg.quietPeriod = cfg.quietPeriod.Truncate(time.Second)
	if cfg.quietPeriod == 0 {
		cfg.quietPeriod = cfg.lastAccess / 4
	}
	if cfg.quietPeriod < cfg.interval {
		cfg.quietPeriod = cfg.interval
	}

	return cfg, nil
}

func readCacheRoot(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return parseCacheRoot(f)
}

func parseCacheRoot(r io.Reader) (string, error) {
	s := bufio.NewScanner(r)

	for s.Scan() {
		cmd, arg := splitLine(s.Text())
		if cmd == "dir" {
			return arg, nil
		}
	}

	err := s.Err()
	if err == nil {
		err = errors.New("dir command not found")
	}
	return "", err
}

func splitLine(s string) (string, string) {
	s = strings.TrimSpace(s)
	i := strings.IndexRune(s, '#')
	if i >= 0 {
		s = s[:i]
	}

	i = strings.IndexFunc(s, unicode.IsSpace)
	if i < 0 {
		return s, ""
	}

	cmd := s[:i]
	arg := strings.TrimSpace(s[i:])
	return cmd, arg
}

func parsePercentage(s string) (uint64, error) {
	// % character is optional
	s = strings.TrimSuffix(s, "%")
	return strconv.ParseUint(s, 10, 0)
}
