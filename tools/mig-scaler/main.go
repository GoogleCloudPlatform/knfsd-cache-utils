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
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/pflag"

	execpb "google.golang.org/genproto/googleapis/cloud/workflows/executions/v1"
)

type Job struct {
	ID         string
	StartTime  time.Time
	EndTime    time.Time
	State      execpb.Execution_State
	MIG        MIGRef
	TargetSize uint32
	Increment  uint32
	Wait       time.Duration
	Deadline   time.Time
}

type JobPayload struct {
	// MaxUint32 is less than javascript's MAX_SAFE_INTEGER so there's no need
	// to use a string for the json representation of uint32
	Project    string    `json:"project"`
	Region     string    `json:"region"`
	Zone       string    `json:"zone"`
	Name       string    `json:"name"`
	TargetSize uint32    `json:"target_size"`
	Increment  uint32    `json:"increment"`
	Wait       uint32    `json:"wait"`
	Deadline   time.Time `json:"deadline"`
}

type Command interface {
	Execute(ctx context.Context) error
}

func main() {
	var err error
	log.SetFlags(0)

	cfg := new(Config)

	f := pflag.NewFlagSet("mig-scaler", pflag.ContinueOnError)

	// setup flags before reading the config files , otherwise the pflag package
	// will overwrite the config with the default values
	f.StringVar(&cfg.Workflow.Project, "workflow-project", "", "")
	f.StringVar(&cfg.Workflow.Region, "workflow-region", "", "")
	f.StringVar(&cfg.Workflow.Name, "workflow-name", "mig-scaler", "")
	f.Uint32Var(&cfg.Workflow.MaxSize, "workflow-max-size", 100, "")
	f.DurationVar(&cfg.Workflow.MaxDuration, "workflow-max-duration", 8*time.Hour, "")

	f.StringVar(&cfg.MIG.Project, "project", "", "")
	f.StringVar(&cfg.MIG.Region, "region", "", "")
	f.StringVar(&cfg.MIG.Zone, "zone", "", "")

	// default rate of 10 machines per minute
	// if the clients have 32 cores this will start 9,600 cores in 30 minutes.
	f.Uint32Var(&cfg.MIG.Increment, "increment", 10, "")
	f.DurationVar(&cfg.MIG.Wait, "wait", 1*time.Minute, "")
	f.DurationVar(&cfg.MIG.Duration, "duration", 0, "")

	f.StringVar(&cfg.Format, "format", "table", "")
	f.BoolVar(&cfg.Detailed, "detailed", false, "")
	f.BoolVar(&cfg.ActiveOnly, "active", false, "")

	f.BoolVarP(&cfg.Help, "help", "h", false, "")

	// read the config file before parsing the command line arguments so
	// that the command line arguments override any config values
	err = readDefaultConfig(cfg)
	if err != nil {
		log.Printf("error: could not read config: %s", err)
		os.Exit(2)
	}

	// override values from the config file with environment variables
	readEnv(cfg)

	// command line arguments overrides all other sources
	err = f.Parse(os.Args[1:])

	// grab the positional args before checking for an error so that they can
	// be used by ShowUsage
	cfg.Args = f.Args()
	if len(cfg.Args) >= 1 {
		cfg.Command = cfg.Args[0]
		cfg.Args = cfg.Args[1:]
	}
	if err != nil {
		log.Printf("error: %+v", err)
		showUsage(cfg)
		os.Exit(2)
	}

	if f.Lookup("region").Changed && !f.Lookup("zone").Changed {
		// if the region was set, and the zone was not set, clear the zone
		// so that setting --region=value on the command line overrides zone
		// from the environment or config
		cfg.MIG.Zone = ""
	}

	// Default the workflow location to the MIG location if the workflow was
	// not set. This allows running with just --project=... --region=... without
	// an ini file.
	defaultString(&cfg.Workflow.Project, cfg.MIG.Project)
	defaultString(&cfg.Workflow.Region, cfg.MIG.Region)

	// Likewise, default the job's max duration to the workflow's max duration
	// if the job's max duration was not set.
	defaultDuration(&cfg.MIG.Duration, cfg.Workflow.MaxDuration)

	// Truncate durations to the nearest second
	truncateDuration(&cfg.Workflow.MaxDuration)
	truncateDuration(&cfg.MIG.Wait)
	truncateDuration(&cfg.MIG.Duration)

	cmd, err := CreateCommand(cfg)
	if err != nil {
		printConfigError(err)
		showUsage(cfg)
		os.Exit(2)
	}

	ctx := context.Background()
	err = cmd.Execute(ctx)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}
}

func CreateCommand(cfg *Config) (Command, error) {
	name := cfg.Command
	if cfg.Help {
		// If the help flag has been set, show the help instead, but do not
		// change the command so that the help can use it for the topic.
		name = "help"
	}

	switch name {
	case "":
		return nil, errors.New("command not specified")
	case "list":
		return NewList(cfg)
	case "scale":
		return NewScale(cfg)
	case "cancel":
		return NewCancel(cfg)
	case "help":
		return NewHelp(cfg), nil
	default:
		return nil, fmt.Errorf("unknown command: \"%s\"", name)
	}
}
