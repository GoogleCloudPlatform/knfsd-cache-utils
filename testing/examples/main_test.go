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

package main

import (
	"examples/common"
	"examples/compute"
	"examples/stage"
	"examples/testing"
	"os"
	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/terraform"
	scope "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
)

const ExampleDir = "../../examples"

var cloudbuild bool = os.Getenv("CI") == "cloudbuild"

func TestMain(m *testing.M) {
	if cloudbuild {
		compute.UseInternalIP = true
	}

	var code int
	failed := !testing.Run("TestMain", func(t testing.TestingT) {
		code = RunTests(m, t)
	})
	if failed && code == 0 {
		code = 1
	}
	os.Exit(code)
}

func RunTests(m *testing.M, t testing.TestingT) (code int) {
	defer common.TearDown(t)
	common.Setup(t)
	return m.Run()
}

func TestBasicExample(t *testing.T) {
	t.Parallel()
	runExampleTest(t, "basic", func(o testing.Outputs) {
		project := o.Project(t)
		zone := o.Zone(t)
		instanceGroupName := o.ProxyMIG(t)
		instanceGroup := compute.FetchZonalInstanceGroup(t, project, zone, instanceGroupName)

		assertMIGSize(t, project, instanceGroup, 1)
		instanceGroup.ForAllInstances(t, func(i *compute.Instance) {
			assertServiceStatus(t, i, "fsidd", "inactive")
			assertServiceStatus(t, i, "knfsd-fsidd", "inactive")
		})
	})
}

func TestStandardExample(t *testing.T) {
	t.Parallel()
	runExampleTest(t, "standard", func(o testing.Outputs) {
		project := o.Project(t)
		zone := o.Zone(t)
		instanceGroupName := o.ProxyMIG(t)
		instanceGroup := compute.FetchZonalInstanceGroup(t, project, zone, instanceGroupName)

		assertMIGSize(t, project, instanceGroup, 3)
		instanceGroup.ForAllInstances(t, func(i *compute.Instance) {
			assertServiceStatus(t, i, "fsidd", "inactive")
			assertServiceStatus(t, i, "knfsd-fsidd", "active")
			assertServiceSubStatus(t, i, "knfsd-fsidd", "running")
		})
	})
}

type ValidateFunc func(testing.Outputs)

func runExampleTest(t *testing.T, example string, validate ValidateFunc) {
	varFiles, err := testing.FindVarFiles(example)
	require.NoError(t, err)

	terraformDir := filepath.Join(ExampleDir, example)
	terraformOptions := &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"name":       testing.GetTestID(t, terraformDir),
			"export_map": common.ExportMap(t),
		},
		VarFiles: varFiles,
	}

	defer stage.RunDestroy(t, func() {
		terraform.Destroy(t, terraformOptions)
		scope.CleanupTestDataFolder(t, terraformDir)
	})

	stage.RunApply(t, func() {
		terraform.Init(t, terraformOptions)
		terraform.Apply(t, terraformOptions)
	})

	stage.RunValidate(t, func() {
		outputs := testing.Outputs(terraform.OutputAll(t, terraformOptions))
		validate(outputs)
	})
}
