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
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	metadataServerURL = "http://metadata.google.internal"
)

// getMetadataValue fetches a metadata path from the GCE Metadata Server returning the output as a string
// if delimit is set to true then it will return the string after the last occurrence of /
func getMetadataValue(URI string, delimit bool) (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?alt=text", metadataServerURL, URI), nil)
	if err != nil {
		return "", err
	}

	// As per https://cloud.google.com/compute/docs/metadata/overview#parts-of-a-request
	// all metadata queries need to include "Metadata-Flavor: Google" in the
	// HTTP headers.
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received invalid HTTP response code, got %d, wanted 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if delimit {
		return lastAfterDelimiter(string(body), "/"), nil
	} else {
		return string(body), nil
	}
}
