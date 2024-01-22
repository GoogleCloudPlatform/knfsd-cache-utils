/*
 Copyright 2023 Google LLC

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

package smoke_tests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	scope "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
)

const (
	KernelVersion = "6.4.0-060400-knfsd"
	TestSSHUser   = "test"
)

var cloudbuild bool = os.Getenv("CI") == "cloudbuild"

func TestEnv(t *testing.T) {
	env := os.Environ()
	for _, e := range env {
		if strings.HasPrefix(e, "SKIP_") {
			fmt.Println(e)
		}
	}
}

func TestSmoke(t *testing.T) {
	var applied bool

	defer scope.RunTestStage(t, "destroy", func() {
		terraformOptions := scope.LoadTerraformOptions(t, "terraform")
		terraform.Destroy(t, terraformOptions)
	})

	scope.RunTestStage(t, "check", func() {
		// Verify remote.test exists if running the check stage before running
		// apply. Otherwise we might just waste time running a costly apply only
		// for check to immediately fail.
		require.FileExists(t, "./remote.test")
	})

	scope.RunTestStage(t, "apply", func() {
		terraformOptions := &terraform.Options{
			TerraformDir: "terraform",
			Vars: map[string]interface{}{
				"prefix": gcp.RandomValidGcpName(),
			},
		}

		scope.SaveTerraformOptionsIfNotPresent(t, "terraform", terraformOptions)
		terraformOptions = scope.LoadTerraformOptions(t, "terraform")

		terraform.Init(t, terraformOptions)
		terraform.Apply(t, terraformOptions)
		applied = true
	})

	scope.RunTestStage(t, "check", func() {
		terraformOptions := scope.LoadTerraformOptions(t, "terraform")
		outputs := Outputs(terraform.OutputAll(t, terraformOptions))

		if applied {
			// TODO: find a better way to test if the client VM is ready.
			// Currently terraform apply will complete before the client VM is
			// ready, so the scp command fails because sshd is not yet running.
			d := 1 * time.Minute
			terraformOptions.Logger.Logf(t, "Waiting %s for client VM", d)
			time.Sleep(d)
		}

		copyRemote(t, outputs)
		executeRemote(t, outputs)
	})
}

func copyRemote(t *testing.T, outputs Outputs) {
	project := outputs.Project(t)
	zone := outputs.Zone(t)
	instance := outputs.ClientInstance(t)

	var args []string
	args = append(args,
		"compute", "scp",
		"--project", project,
		"--zone", zone,
	)
	if cloudbuild {
		args = append(args, "--internal-ip")
	}
	args = append(args,
		"./remote.test",
		fmt.Sprintf("%s@%s:./remote.test", TestSSHUser, instance),
	)
	shell.RunCommand(t, shell.Command{
		Command: "gcloud",
		Args:    args,
	})
}

func executeRemote(t *testing.T, outputs Outputs) {
	project := outputs.Project(t)
	zone := outputs.Zone(t)
	instance := outputs.ClientInstance(t)

	var args []string
	args = append(args,
		"compute", "ssh",
		"--project", project,
		"--zone", zone,
	)
	if cloudbuild {
		args = append(args, "--internal-ip")
	}
	args = append(args,
		"--command", "sudo ./remote.test",
		fmt.Sprintf("%s@%s", TestSSHUser, instance),
	)
	shell.RunCommand(t, shell.Command{
		Command: "gcloud",
		Args:    args,
	})
}

type Outputs map[string]interface{}

func (o Outputs) Project(t *testing.T) string {
	return o.GetString(t, "project")
}

func (o Outputs) Zone(t *testing.T) string {
	return o.GetString(t, "zone")
}

func (o Outputs) ClientInstance(t *testing.T) string {
	return o.GetString(t, "client_instance")
}

func (o Outputs) GetString(t *testing.T, key string) string {
	entry, ok := o[key]
	if !ok {
		require.FailNow(t, fmt.Sprintf("Required output %s was missing", key))
	}

	val, ok := entry.(string)
	if !ok {
		require.FailNow(t, fmt.Sprintf("Required output %s was not a string", key))
	}

	if val == "" {
		require.FailNow(t, fmt.Sprintf("Required output %s was empty", key))
	}

	return val
}
