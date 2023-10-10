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
	"encoding/json"
	"os"
	"testing"

	"github.com/prometheus/procfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMounts(t *testing.T) {
	fs, err := procfs.NewFS("./testdata/proc/")
	require.NoError(t, err)

	proc, err := fs.Proc(1)
	require.NoError(t, err)

	mounts, err := readMounts(proc, "/srv/nfs/")
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/expected/mounts.json")
	require.NoError(t, err)

	actual, err := json.MarshalIndent(&mounts, "", "  ")
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), string(actual))
}

func TestReadMountStats(t *testing.T) {
	fs, err := procfs.NewFS("./testdata/proc/")
	require.NoError(t, err)

	proc, err := fs.Proc(1)
	require.NoError(t, err)

	mounts, err := readMountStats(proc, "/srv/nfs/")
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/expected/mountstats.json")
	require.NoError(t, err)

	actual, err := json.MarshalIndent(&mounts, "", "  ")
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), string(actual))
}
