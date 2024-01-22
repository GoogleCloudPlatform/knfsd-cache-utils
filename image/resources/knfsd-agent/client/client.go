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
	"net/http"
	"net/url"
)

type KnfsdAgentClient struct {
	c       *http.Client
	baseURL string
}

func NewKnfsdAgentClient(c *http.Client, baseURL string) *KnfsdAgentClient {
	return &KnfsdAgentClient{c, baseURL}
}

func (c *KnfsdAgentClient) get(path string, v any) error {
	url, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return err
	}

	res, err := c.c.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(v)
}
