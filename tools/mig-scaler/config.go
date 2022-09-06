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
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"go.uber.org/multierr"
)

type MIGRef struct {
	Project string
	Region  string
	Zone    string
	Name    string
}

func (ref MIGRef) Validate() error {
	var err error
	err = multierr.Append(err, require("project", ref.Project))
	if ref.Region == "" && ref.Zone == "" {
		err = multierr.Append(err, errors.New("required: region and/or zone"))
	}
	err = multierr.Append(err, require("name", ref.Name))
	return err
}

func (ref MIGRef) Location() string {
	if ref.Zone == "" {
		return ref.Region
	} else {
		return ref.Zone
	}
}

func (ref MIGRef) Normalize() MIGRef {
	// Normalize the reference so that only region or zone is set, not both.
	// Zone has a higher precedence, so if zone is set, ignore region.
	var region, zone string
	if ref.Zone == "" {
		region = ref.Region
	} else {
		zone = ref.Zone
	}

	return MIGRef{
		Project: strings.ToLower(ref.Project),
		Region:  strings.ToLower(region),
		Zone:    strings.ToLower(zone),
		Name:    strings.ToLower(ref.Name),
	}
}

type Config struct {
	// Configuration settings for the workflow, this allows a single workflow
	// to control client MIGs in other projects. Generally this will be set
	// via a config file as all the client MIGs will use a common workflow.
	Workflow WorkflowConfig `ini:"workflow"`

	// The details for the client MIG to scale up
	MIG MIGConfig `ini:"mig"`

	// Format specifies the format to use for the list command. Allow specifying
	// this in the config file so a user can set a default.
	Format string `ini:"format"`

	// Include more details in the output
	Detailed bool `ini:"detailed"`

	// Only show active jobs for the list command
	ActiveOnly bool `ini:"-"`

	// --help flag has been set
	Help bool `ini:"-"`

	// positional command line arguments
	Command string   `ini:"-"`
	Args    []string `ini:"-"`
}

func (c *Config) Arg(i int) string {
	if i < 0 || i >= len(c.Args) {
		return ""
	} else {
		return c.Args[i]
	}
}

type WorkflowConfig struct {
	Project     string        `ini:"project"`
	Region      string        `ini:"region"`
	Name        string        `ini:"name"`
	MaxDuration time.Duration `ini:"max-duration"`
	MaxSize     uint32        `ini:"max-size"`
}

func (cfg WorkflowConfig) FullName() string {
	return workflowName(cfg.Project, cfg.Region, cfg.Name)
}

func (cfg *WorkflowConfig) Oldest() time.Time {
	return time.Now().Add(-cfg.MaxDuration)
}

func (cfg *WorkflowConfig) Validate() error {
	var err error

	err = multierr.Append(err, require("workflow-project", cfg.Project))
	err = multierr.Append(err, require("workflow-region", cfg.Region))
	err = multierr.Append(err, require("workflow-name", cfg.Name))

	if cfg.MaxDuration <= 0 {
		err = multierr.Append(err, errors.New("max-duration must be greater than 0"))
	}

	return err
}

type MIGConfig struct {
	Project   string        `ini:"project"`
	Region    string        `ini:"region"`
	Zone      string        `ini:"zone"`
	Increment uint32        `ini:"increment"`
	Wait      time.Duration `ini:"wait"`
	Duration  time.Duration `ini:"duration"`
}

func (cfg *MIGConfig) Ref(name string) MIGRef {
	ref := MIGRef{
		Project: cfg.Project,
		Region:  cfg.Region,
		Zone:    cfg.Zone,
		Name:    name,
	}
	return ref.Normalize()
}

func readDefaultConfig(cfg *Config) error {
	const name = "mig-scaler.conf"
	paths := make([]string, 2)

	home, _ := os.UserHomeDir()
	if home != "" {
		paths = append(paths, filepath.Join(home, ".config", name))
	}

	paths = append(paths, "/etc/"+name)

	for _, p := range paths {
		err := readConfig(cfg, p)

		switch {
		case err == nil:
			return nil
		case errors.Is(err, os.ErrNotExist):
			continue
		default:
			return err
		}
	}

	// no config file found
	return nil
}

func readConfig(cfg *Config, name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return parseConfig(cfg, f)
}

func parseConfig(cfg *Config, r io.Reader) error {
	i, err := ini.Load(r)
	if err != nil {
		return err
	}
	return i.StrictMapTo(cfg)
}

func readEnv(cfg *Config) error {
	var err error

	envString(&cfg.Workflow.Project, "MIGSCALER_WORKFLOW_PROJECT")
	envString(&cfg.Workflow.Region, "MIGSCALER_WORKFLOW_REGION")
	envString(&cfg.Workflow.Name, "MIGSCALER_WORKFLOW_NAME")
	err = multierr.Append(err, envDuration(&cfg.Workflow.MaxDuration, "MIGSCALER_MAX_DURATION"))
	err = multierr.Append(err, envUint32(&cfg.Workflow.MaxSize, "MIGSCALER_MAX_SIZE"))

	envString(&cfg.MIG.Project, "MIGSCALER_PROJECT")

	var region, zone string
	envString(&region, "MIGSCALER_REGION")
	envString(&zone, "MIGSCALER_ZONE")
	if region != "" || zone != "" {
		// Update region and zone as a pair, so that the environment variables
		// can override a zone in the config file with a region.
		cfg.MIG.Region = region
		cfg.MIG.Zone = zone
	}

	err = multierr.Append(err, envUint32(&cfg.MIG.Increment, "MIGSCALER_INCREMENT"))
	err = multierr.Append(err, envDuration(&cfg.MIG.Wait, "MIGSCALER_WAIT"))
	err = multierr.Append(err, envDuration(&cfg.MIG.Duration, "MIGSCALER_DURATION"))

	return err
}

func envString(value *string, key string) {
	if s, _ := os.LookupEnv(key); s != "" {
		*value = s
	}
}

func envUint32(value *uint32, key string) error {
	if s, _ := os.LookupEnv(key); s != "" {
		i, err := parseUint32(s)
		if err != nil {
			return err
		}
		*value = i
	}
	return nil
}

func envDuration(value *time.Duration, key string) error {
	if s, _ := os.LookupEnv(key); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("%s: invalid duration", key)
		}
		if d <= 0 {
			return fmt.Errorf("%s: duration must be greater than 0", key)
		}
		*value = d
	}
	return nil
}

func parseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		err = fmt.Errorf("must be a valid integer between 1 and %d", math.MaxUint32)
	}
	return uint32(v), err
}

func defaultString(value *string, def string) {
	if value == nil || *value == "" {
		*value = def
	}
}

func defaultDuration(value *time.Duration, def time.Duration) {
	if value == nil || *value == 0 {
		*value = def
	}
}

func truncateDuration(value *time.Duration) {
	if value == nil {
		*value = 0
	} else {
		*value = value.Truncate(time.Second)
	}
}

func require(name, value string) error {
	if value == "" {
		return fmt.Errorf("required: %s", name)
	} else {
		return nil
	}
}

func printConfigError(err error) {
	log.Printf("error: invalid configuration:")
	for _, e := range multierr.Errors(err) {
		log.Printf("  - %v", e)
	}
}
