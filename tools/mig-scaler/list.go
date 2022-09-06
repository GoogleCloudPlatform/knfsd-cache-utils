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
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"go.uber.org/multierr"
	"google.golang.org/api/iterator"
	execpb "google.golang.org/genproto/googleapis/cloud/workflows/executions/v1"
)

type List struct {
	// list only needs to know the workflow details as it does not interact with
	// a specific MIG.
	Workflow   WorkflowConfig
	Format     Formatter
	Detailed   bool
	ActiveOnly bool
}

type Formatter func([]*Job, bool)

var formatters = map[string]Formatter{
	"list":  formatList,
	"table": formatTable,
}

func NewList(cfg *Config) (*List, error) {
	var err error

	cmd := &List{
		Workflow:   cfg.Workflow,
		Detailed:   cfg.Detailed,
		ActiveOnly: cfg.ActiveOnly,
	}

	err = multierr.Append(err, cmd.Workflow.Validate())

	format := strings.ToLower(cfg.Format)
	if f := formatters[format]; f != nil {
		cmd.Format = f
	} else {
		err = multierr.Append(err, fmt.Errorf("error: unknown output format \"%s\"", format))
	}

	if err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
}

func (cmd *List) Execute(ctx context.Context) error {
	client, err := NewClient(ctx, cmd.Workflow)
	if err != nil {
		return err
	}
	defer client.Close()

	req := &execpb.ListExecutionsRequest{
		Parent:   cmd.Workflow.FullName(),
		PageSize: 100,
		View:     execpb.ExecutionView_FULL,
	}

	var jobs []*Job
	recent := make(map[MIGRef]*Job)

	oldest := cmd.Workflow.Oldest()
	it := client.c.ListExecutions(ctx, req)
	for {
		e, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("error: could not fetch page: %v", err)
		}
		if e.StartTime.AsTime().Before(oldest) {
			break
		}

		j, err := parseJob(e)
		if err != nil {
			log.Printf("warn: could not parse job details for \"%s\": %v", e.Name, err)
			continue
		}

		if j.State == execpb.Execution_ACTIVE {
			// include all active jobs, put these at the top of the list
			jobs = append(jobs, j)
		} else if !cmd.ActiveOnly {
			// only include non-active jobs if ActiveOnly is false
			if _, found := recent[j.MIG]; !found {
				// only include the most recent non-active job for each MIG
				recent[j.MIG] = j
			}
		}
	}

	// remove any MIGs from the recent list that have an active job
	for _, j := range jobs {
		delete(recent, j.MIG)
	}

	// add any remaining non-active jobs to the job list
	for _, j := range recent {
		jobs = append(jobs, j)
	}

	// sort the jobs, active first, then by when the job was submitted
	sort.Slice(jobs, func(i, j int) bool {
		x := jobs[i]
		y := jobs[j]

		// active first
		if active(x.State) && !active(y.State) {
			return true
		}
		if !active(x.State) && active(y.State) {
			return false
		}

		// most recent first
		return x.StartTime.After(y.StartTime)
	})

	if len(jobs) == 0 {
		fmt.Println("No recent jobs found.")
	} else {
		cmd.Format(jobs, cmd.Detailed)
	}

	return nil
}

func formatTable(jobs []*Job, detailed bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	if detailed {
		// detailed is unlikely to fit on a single line unless
		fmt.Fprint(w, "ID\tProject\tRegion\tZone\tName\tTarget\tIncrement\tWait\tStatus\tStarted\tFinished\tDeadline\n")
		for _, j := range jobs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\n",
				j.ID,
				j.MIG.Project,
				j.MIG.Region,
				j.MIG.Zone,
				j.MIG.Name,
				j.TargetSize,
				j.Increment,
				j.Wait,
				j.State,
				formatTimestamp(j.StartTime),
				formatTimestamp(j.EndTime),
				j.Deadline,
			)
		}
	} else {
		// keep the output compact to try and fit within 80 columns
		fmt.Fprint(w, "Project\tLocation\tName\tTarget\tIncrement\tWait\tStatus\n")
		for _, j := range jobs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
				j.MIG.Project,
				j.MIG.Location(),
				j.MIG.Name,
				j.TargetSize,
				j.Increment,
				j.Wait,
				j.State,
			)
		}
	}

	w.Flush()
}

func formatList(jobs []*Job, detailed bool) {
	for _, j := range jobs {
		if detailed {
			fmt.Printf("ID       : %s\n", j.ID)
		}
		fmt.Printf("Project  : %s\n", j.MIG.Project)
		if detailed {
			fmt.Printf("Region   : %s\n", j.MIG.Region)
			fmt.Printf("Zone     : %s\n", j.MIG.Zone)
		} else {
			fmt.Printf("Location : %s\n", j.MIG.Location())
		}
		fmt.Printf("Name     : %s\n", j.MIG.Name)
		fmt.Printf("Target   : %d\n", j.TargetSize)
		fmt.Printf("Increment: %d\n", j.Increment)
		fmt.Printf("Wait     : %s\n", j.Wait)
		fmt.Printf("State    : %s\n", j.State)
		if detailed {
			fmt.Printf("Started  : %s\n", formatTimestamp(j.StartTime))
			fmt.Printf("Finished : %s\n", formatTimestamp(j.EndTime))
			fmt.Printf("Deadline : %s\n", formatTimestamp(j.Deadline))
		}
		fmt.Println()
	}
}

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	} else {
		return t.Local().Format(time.RFC3339)
	}
}

func active(state execpb.Execution_State) bool {
	return state == execpb.Execution_ACTIVE
}
