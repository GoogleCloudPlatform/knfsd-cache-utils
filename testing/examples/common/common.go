/*
 Copyright 2024 Google LLC

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

package common

import (
	"examples/stage"
	"examples/testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	scope "github.com/gruntwork-io/terratest/modules/test-structure"
)

type TestingT = testing.TestingT

var (
	options *terraform.Options
	outputs testing.Outputs
)

func Source(t TestingT) string {
	return outputs.GetString(t, "source")
}

func ExportMap(t TestingT) string {
	return outputs.GetString(t, "export_map")
}

func Setup(t TestingT) {
	varFiles, err := testing.FindVarFiles("setup")
	if err != nil {
		t.Fatalf("could not resolve var files for setup: %s\n", err)
	}

	terraformDir := "./common"
	options = &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"name": testing.GetTestID(t, terraformDir),
		},
		VarFiles: varFiles,
	}

	stage.RunApply(t, func() {
		terraform.Init(t, options)
		terraform.Apply(t, options)
	})

	outputs = testing.Outputs(terraform.OutputAll(t, options))
}

func TearDown(t TestingT) {
	if options == nil {
		return
	}
	stage.RunDestroy(t, func() {
		terraform.Destroy(t, options)
		scope.CleanupTestDataFolder(t, options.TerraformDir)
	})
}
