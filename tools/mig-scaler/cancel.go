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
	"fmt"

	"go.uber.org/multierr"
)

type Cancel struct {
	Workflow WorkflowConfig
	MIG      MIGRef
}

func NewCancel(cfg *Config) (*Cancel, error) {
	var err error

	// expecting 1 arguments: cancel <name>
	if len(cfg.Args) != 1 {
		return nil, fmt.Errorf("cancel expects 1 arguments but %d were provided", len(cfg.Args))
	}

	name := cfg.Arg(0)
	cmd := &Cancel{
		Workflow: cfg.Workflow,
		MIG:      cfg.MIG.Ref(name),
	}

	err = multierr.Append(err, cmd.Workflow.Validate())
	err = multierr.Append(err, cmd.MIG.Validate())

	if err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
}

func (cmd *Cancel) Execute(ctx context.Context) error {
	client, err := NewClient(ctx, cmd.Workflow)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Cancel(ctx, cmd.MIG)
}
