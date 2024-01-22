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
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func QueryAttribute(name string) (string, error) {
	url := fmt.Sprintf("http://metadata.google.internal/computeMetadata/v1/instance/attributes/%s?alt=text", name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", nil
	}

	req.Header.Add("Metadata-Flavor", "Google")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New(res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(body)), nil
}
