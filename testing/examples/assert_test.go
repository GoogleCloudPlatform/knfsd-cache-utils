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
	"examples/compute"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertMIGSize(t *testing.T, project string, instanceGroup *compute.InstanceGroup, expectedClusterSize int) {
	t.Helper()
	instances, err := instanceGroup.GetInstancesE(t, project)
	require.NoErrorf(t, err, "could not fetch instances for instance group %d", instanceGroup)
	require.Lenf(t, instances, expectedClusterSize, "expected to find %d instances", expectedClusterSize)
}

func assertServiceStatus(t *testing.T, instance *compute.Instance, serviceName string, expectedStatus string) {
	t.Helper()
	cmd := fmt.Sprintf("systemctl show -p ActiveState --value %s", serviceName)
	out := instance.Execute(t, cmd)
	assert.Equal(t, expectedStatus, strings.TrimSpace(out))
}

func assertServiceSubStatus(t *testing.T, instance *compute.Instance, serviceName string, expectedStatus string) {
	t.Helper()
	cmd := fmt.Sprintf("systemctl show -p SubState --value %s", serviceName)
	out := instance.Execute(t, cmd)
	assert.Equal(t, expectedStatus, strings.TrimSpace(out))
}
