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
	"testing"

	terratest "github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type M = testing.M
type T = testing.T

// Verify that the real testing.T type is compatible with this interface
var _ TestingT = (*testing.T)(nil)

// TestingT provides an interface for testing.T that is compatible with both
// terratest and stretchr testify.
type TestingT interface {
	require.TestingT
	assert.TestingT
	terratest.TestingT
}
