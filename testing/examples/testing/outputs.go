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

package testing

import (
	"fmt"

	"github.com/stretchr/testify/require"
)

type Outputs map[string]interface{}

func (o Outputs) Project(t TestingT) string {
	return o.GetString(t, "project")
}

func (o Outputs) Zone(t TestingT) string {
	return o.GetString(t, "zone")
}

func (o Outputs) Region(t TestingT) string {
	return o.GetString(t, "region")
}

func (o Outputs) ProxyMIG(t TestingT) string {
	return o.GetString(t, "proxy_instance_group")
}

func (o Outputs) GetString(t TestingT, key string) string {
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
