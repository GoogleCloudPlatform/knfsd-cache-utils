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
	"errors"
	"strings"
)

// lastAfterDelimiter returns the string after the last occurrence of a delimiter
func lastAfterDelimiter(path string, seperator string) (string, error) {

	split := strings.Split(path, seperator)

	if len(split) == 0 {
		return "", errors.New("slice length was 0")
	}

	return split[len(split)-1], nil

}
