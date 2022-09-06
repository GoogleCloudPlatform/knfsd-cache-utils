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
	"math"
	"time"

	"go.uber.org/multierr"
)

// This the the absolute upper bound for the wait parameter as job.Wait
// represents the wait in seconds as a uint32.
const MaxWait = math.MaxUint32 * time.Second

// Estimate a worst case scenario for how long a single cycle of the workflow
// will take. This includes the HTTP request time, and overhead from the sleep
// command (sleep waits a minimum of the wait time, but can be longer).
const CycleTime = 10 * time.Second

type Scale struct {
	Workflow   WorkflowConfig
	MIG        MIGRef
	TargetSize uint32
	Increment  uint32
	Wait       time.Duration
	Duration   time.Duration
}

func NewScale(cfg *Config) (*Scale, error) {
	var err error

	// expecting 2 arguments: scale <name> <target>
	if len(cfg.Args) != 2 {
		return nil, fmt.Errorf("scale expects 2 arguments but %d were provided", len(cfg.Args))
	}

	name := cfg.Arg(0)
	target, err := parseUint32(cfg.Arg(1))
	if err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}

	cmd := &Scale{
		Workflow:   cfg.Workflow,
		MIG:        cfg.MIG.Ref(name),
		TargetSize: target,
		Increment:  cfg.MIG.Increment,
		Wait:       cfg.MIG.Wait,
		Duration:   cfg.MIG.Duration,
	}

	err = multierr.Append(err, cmd.Workflow.Validate())
	err = multierr.Append(err, cmd.MIG.Validate())

	if cmd.TargetSize > cfg.Workflow.MaxSize {
		err = multierr.Append(err, fmt.Errorf(
			"target cannot be greater than workflow-max-target (%d)",
			cfg.Workflow.MaxSize,
		))
	}

	if cmd.Duration <= 0 {
		err = multierr.Append(err, errors.New("duration must be greater than 0"))
	} else if cmd.Duration > cmd.Workflow.MaxDuration {
		err = multierr.Append(err, fmt.Errorf(
			"duration cannot be greater than workflow-max-duration (%s)",
			cmd.Workflow.MaxDuration,
		))
	}

	if cmd.Wait <= 0 {
		err = multierr.Append(err, errors.New("wait must be greater than 0"))
	}

	if cmd.Increment == 0 {
		// apply a default increment of 5%
		err = multierr.Append(err, errors.New("increment must by greater than 0"))
	}

	if err == nil {
		// Only try to check if the increment/wait is valid for the duration
		// if there's no error so far. If there is an error then the config
		// values are not trustworthy so this calculation could be nonsense.
		minIncrement := cmd.EstimateMinIncrement()
		maxWait := cmd.EstimateMaxWait()

		if cmd.Increment < minIncrement {
			err = multierr.Append(err, fmt.Errorf("increment is too small, based on wait and duration the minimum increment is %d", minIncrement))
		}
		if maxWait > 0 && cmd.Wait > maxWait {
			err = multierr.Append(err, fmt.Errorf("wait is too long, based on the increment and duration the maximum wait is %s", maxWait))
		}
	}

	if err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
}

func (cmd *Scale) Execute(ctx context.Context) error {
	var err error

	client, err := NewClient(ctx, cmd.Workflow)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.Cancel(ctx, cmd.MIG)
	if err != nil {
		return err
	}

	job := &JobPayload{
		Project:    cmd.MIG.Project,
		Name:       cmd.MIG.Name,
		Region:     cmd.MIG.Region,
		Zone:       cmd.MIG.Zone,
		TargetSize: cmd.TargetSize,
		Increment:  cmd.Increment,
		Wait:       uint32(cmd.Wait.Seconds()),
		Deadline:   time.Now().Add(cmd.Duration),
	}

	e, err := client.Execute(ctx, job)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Printf("info: job started \"%s\"", e.Name)
	// output the full name of the workflow execution to stdout to support scripts
	fmt.Println(e.Name)
	return nil
}

// EstimateMinIncrement calculates the minimum increment size required bring the
// MIG to it's target size based on the wait and duration.
// EstimateMinIncrement assumes the MIGs initial size is zero.
func (cmd *Scale) EstimateMinIncrement() uint32 {
	// Estimate the theoretical number of cycles required, assuming no errors.
	// Include some overhead (CycleTime) to account for the time required for
	// the workflow to execute a cycle.
	cycles := uint64(cmd.Duration / (cmd.Wait + CycleTime))

	// The workflow contains an additional final cycle after the final
	// request to check that the target size has been reached.
	cycles += 1

	min := uint64(cmd.TargetSize) / cycles
	if min > math.MaxUint32 {
		// if we actually reach this value we've received some very weird inputs
		return math.MaxUint32
	} else {
		return uint32(min)
	}
}

// EstimateMaxWait calculates the maximum wait allowed to bring the MIG to
// it's target size based on the increment and duration.
// EstimateMaxWait assumes the MIGs initial size is zero.
func (cmd *Scale) EstimateMaxWait() time.Duration {
	// integer division rounding up, e.g. 5 / 2 = 3
	cycles := IntDiv(cmd.TargetSize, cmd.Increment)

	// The workflow contains an additional final cycle after the final
	// request to check that the target size has been reached.
	if cycles < math.MaxUint32 {
		cycles += 1
	}

	maxWait := cmd.Duration / time.Duration(cycles)

	// Subtract some overhead (CycleTime) to account for the time required for
	// the workflow to execute a cycle.
	maxWait -= CycleTime

	if maxWait < 0 {
		return 0
	} else {
		return maxWait.Truncate(time.Second)
	}
}

// IntDiv divides two uint32 values, rounding up
func IntDiv(x, y uint32) uint32 {
	// promote to uint64 to ensure the add can never overflow
	return uint32((uint64(x) + uint64(y) - 1) / uint64(y))
}
