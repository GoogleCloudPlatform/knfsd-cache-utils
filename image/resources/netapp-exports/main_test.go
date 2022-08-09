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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExports(t *testing.T) {
	json, err := os.ReadFile("testdata/responses/simple.json")
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(json)
	}))
	defer server.Close()

	s := &NetAppServer{
		Host:     "netapp.test",
		URL:      server.URL,
		User:     "test",
		Password: "test",
		TLS: &TLSConfig{
			insecure: true,
		},
	}

	err = s.validate()
	require.NoError(t, err)

	expected := strings.Join([]string{
		"netapp.test /",
		"netapp.test /archive",
		"netapp.test /assets",
		"",
	}, "\n")

	actual := new(strings.Builder)
	err = listExports(actual, s)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual.String())
}
