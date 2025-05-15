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

package main

import (
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	KERNEL_VERSION = "6.11.0-1014-gcp"
)

func TestKernelVersion(t *testing.T) {
	t.Parallel()
	version, err := proxy.GetOS()
	require.NoError(t, err)
	assert.Equal(t, KERNEL_VERSION, version.Kernel)
}

func TestStatus(t *testing.T) {
	t.Parallel()

	type Expected struct {
		name   string
		checks []string
	}

	expected := []Expected{
		{"cachefilesd", []string{"enabled", "running", "fscache mounted"}},
	}

	status, err := proxy.GetStatus()
	require.NoError(t, err)

	find := func(name string) *client.ServiceHealth {
		for _, s := range status.Services {
			if s.Name == name {
				return &s
			}
		}
		return nil
	}

	findCheck := func(s *client.ServiceHealth, name string) *client.ServiceCheck {
		for _, c := range s.Checks {
			if c.Name == name {
				return &c
			}
		}
		return nil
	}

	for _, e := range expected {
		t.Run(e.name, func(t *testing.T) {
			s := find(e.name)
			if s == nil {
				t.Fatal("status not found")
			}
			if s.Health != client.CHECK_PASS {
				t.Errorf("check failed: %s", s.Health)
			}

			for _, c := range e.checks {
				t.Run(c, func(t *testing.T) {
					c := findCheck(s, c)
					if c == nil {
						t.Fatalf("check not found")
					}
					if c.Result != client.CHECK_PASS {
						if c.Error != "" {
							t.Error(c.Error)
						} else {
							t.Errorf("check failed: %s", s.Health)
						}
					}
				})
			}
		})
	}
}

func TestProxyMountedSource(t *testing.T) {
	t.Parallel()

	mounts, err := proxy.GetMounts()
	require.NoError(t, err)
	require.NotNil(t, mounts)
	require.Len(t, mounts.Mounts, 1)

	m := mounts.Mounts[0]
	assert.Equal(t, fmt.Sprintf("%s:/files", sourceHost), m.Device)
	assert.Equal(t, "/srv/nfs/files", m.Mount)
	assert.Equal(t, "/files", m.Export)
}
