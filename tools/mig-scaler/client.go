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
	"encoding/json"
	"fmt"
	"log"
	"path"
	"time"

	exec "cloud.google.com/go/workflows/executions/apiv1"
	"google.golang.org/api/iterator"
	execpb "google.golang.org/genproto/googleapis/cloud/workflows/executions/v1"
)

type Client struct {
	c *exec.Client
	w WorkflowConfig
}

func NewClient(ctx context.Context, w WorkflowConfig) (*Client, error) {
	c, err := exec.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create workflow client: %w", err)
	}
	return &Client{c, w}, nil
}

func (client *Client) Close() error {
	return client.c.Close()
}

func (client *Client) Execute(ctx context.Context, job *JobPayload) (*execpb.Execution, error) {
	argument, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("could not serialize job payload to json: %w", err)
	}

	req := &execpb.CreateExecutionRequest{
		Parent: client.w.FullName(),
		Execution: &execpb.Execution{
			Argument: string(argument),
		},
	}

	e, err := client.c.CreateExecution(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not execute workflow: %w", err)
	}
	return e, nil
}

func (client *Client) Cancel(ctx context.Context, mig MIGRef) error {
	active, err := client.findActiveJobs(ctx, mig)
	if err != nil {
		return fmt.Errorf("could not fetch active jobs: %w", err)
	}

	if len(active) > 0 {
		log.Print("info: cancelling existing jobs for MIG")
		err = client.cancelExecutions(ctx, active)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}

	return nil
}

func (client *Client) findActiveJobs(ctx context.Context, mig MIGRef) ([]string, error) {
	var active []string
	req := &execpb.ListExecutionsRequest{
		Parent:   client.w.FullName(),
		PageSize: 100,
		View:     execpb.ExecutionView_BASIC,
	}

	oldest := client.w.Oldest()
	it := client.c.ListExecutions(ctx, req)
	for {
		e, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return active, fmt.Errorf("could not fetch page: %w", err)
		}
		if e.StartTime.AsTime().Before(oldest) {
			break
		}

		// This is N+1 queries, but we're expecting only a couple of active
		// executions, while there might a full page of executions in other
		// states. Currently assuming this will use less bandwidth than fetching
		// a full page of executions with their entire JSON request payload.
		if e.State == execpb.Execution_ACTIVE {
			e, err = client.getExecutionFull(ctx, e.Name)
			if err != nil {
				return active, fmt.Errorf("could not fetch execution \"%s\": %w", e.Name, err)
			}

			j, err := parseJob(e)
			if err != nil {
				log.Printf("warn: could not parse job details for \"%s\": %v", e.Name, err)
			} else if mig == j.MIG {
				// expects both mig and j.MIG to be normalized
				active = append(active, e.Name)
			}
		}
	}

	return active, nil
}

func (client *Client) getExecutionFull(ctx context.Context, name string) (*execpb.Execution, error) {
	req := &execpb.GetExecutionRequest{
		Name: name,
		View: execpb.ExecutionView_FULL,
	}
	return client.c.GetExecution(ctx, req)
}

func (client *Client) cancelExecutions(ctx context.Context, names []string) error {
	for _, name := range names {
		req := &execpb.CancelExecutionRequest{Name: name}
		_, err := client.c.CancelExecution(ctx, req)
		if err != nil {
			return fmt.Errorf("could not cancel existing job \"%s\": %w", name, err)
		}
		log.Printf("info: cancelled \"%s\"", name)
	}
	return nil
}

func parseJob(e *execpb.Execution) (*Job, error) {
	var payload JobPayload
	err := json.Unmarshal([]byte(e.Argument), &payload)
	if err != nil {
		return nil, err
	}

	mig := MIGRef{
		Project: payload.Project,
		Region:  payload.Region,
		Zone:    payload.Zone,
		Name:    payload.Name,
	}
	job := &Job{
		ID:         parseJobID(e.Name),
		State:      e.State,
		MIG:        mig.Normalize(),
		TargetSize: payload.TargetSize,
		Increment:  payload.Increment,
		Wait:       time.Duration(payload.Wait) * time.Second,
		Deadline:   payload.Deadline,
	}

	if e.StartTime != nil {
		job.StartTime = e.StartTime.AsTime()
	}
	if e.EndTime != nil {
		job.EndTime = e.EndTime.AsTime()
	}

	return job, nil
}

func parseJobID(name string) string {
	id := path.Base(name)
	if id == "." || id == "/" {
		id = ""
	}
	return id
}

func workflowName(project, region, workflow string) string {
	return fmt.Sprintf("projects/%s/locations/%s/workflows/%s",
		project, region, workflow)
}

func executionName(project, region, workflow, execution string) string {
	return fmt.Sprintf("projects/%s/locations/%s/workflows/%s/executions/%s",
		project, region, workflow, execution)
}
