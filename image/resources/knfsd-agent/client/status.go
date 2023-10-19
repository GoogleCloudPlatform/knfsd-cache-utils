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

package client

import (
	"encoding/json"
	"fmt"
)

type Check int

var (
	_ fmt.Stringer     = (*Check)(nil)
	_ json.Marshaler   = (*Check)(nil)
	_ json.Unmarshaler = (*Check)(nil)
)

const (
	CHECK_UNKNOWN Check = 0
	CHECK_PASS    Check = -1
	CHECK_WARN    Check = -2
	CHECK_FAIL    Check = -3
)

func (c Check) String() string {
	switch c {
	case CHECK_PASS:
		return "PASS"
	case CHECK_WARN:
		return "WARN"
	case CHECK_FAIL:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

func (c *Check) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"PASS"`:
		*c = CHECK_PASS
	case `"WARN"`:
		*c = CHECK_WARN
	case `"FAIL"`:
		*c = CHECK_FAIL
	default:
		*c = CHECK_UNKNOWN
	}
	return nil
}

func (c Check) MarshalJSON() ([]byte, error) {
	var str string
	switch c {
	case CHECK_PASS:
		str = `"PASS"`
	case CHECK_WARN:
		str = `"WARN"`
	case CHECK_FAIL:
		str = `"FAIL"`
	default:
		// Shouldn't happen, though if we return an error it will prevent
		// marshalling the status response, and then the client won't receive
		// any details other than a generic 500 Internal Server Error.
		str = "UNKNOWN"
	}
	return []byte(str), nil
}

type ServiceHealth struct {
	Name   string         `json:"name"`
	Health Check          `json:"health"`
	Checks []ServiceCheck `json:"checks"`
	Log    string         `json:"log"`
}

type ServiceCheck struct {
	Name   string `json:"name"`
	Result Check  `json:"result"`
	Error  string `json:"error"`
}

type StatusResponse struct {
	Services []ServiceHealth `json:"services"`
}

func (c *KnfsdAgentClient) GetStatus() (*StatusResponse, error) {
	var v *StatusResponse
	err := c.get("api/v1/status", &v)
	return v, err
}
