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

package convert

// Int64() converts a uint64 to an int64 by wrapping overflows back round to zero
// For example:
// int(MaxInt64)     == 9223372036854775807
// int(MaxInt64 + 1) == 0
// int(MaxInt64 + 2) == 1
//
// There will be a small discontinuity from the reset, but most graph
// systems treat a large change in a monotonic counter as a counter reset.
// The alternative would be an inability to report any more metrics for this
// for this counter until the counter eventually overflows back to zero on its
// own.
func Int64(metric uint64) int64 {
	// mask off the uppermost bit (the sign bit for an int64) before converting
	const mask = ^(uint64(1) << 63)
	return int64(metric & mask)
}
