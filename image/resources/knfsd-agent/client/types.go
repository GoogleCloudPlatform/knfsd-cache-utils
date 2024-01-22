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
	"encoding"
	"time"
)

var (
	_ encoding.TextMarshaler   = (*Duration)(nil)
	_ encoding.TextUnmarshaler = (*Duration)(nil)
)

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte((time.Duration)(d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	*d = Duration(v)
	return err
}
