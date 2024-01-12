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
	"runtime"
	"strings"
)

var _ TestingT = (*testRunner)(nil)

// testRunner provides an alternative implementation of TestingT that can be
// used from TestMain. This is to support using various terratest methods such
// as Apply outside of a standard test.
type testRunner struct {
	name   string
	failed bool
}

// Run executes
func Run(name string, f func(t TestingT)) bool {
	r := &testRunner{name: name}
	done := make(chan struct{})
	go func() {
		defer close(done)
		f(r)
	}()
	<-done
	return !r.failed
}

// Name returns the name of the running test or benchmark.
func (r *testRunner) Name() string {
	return r.name
}

func (r *testRunner) log(s string) {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	fmt.Print(s)
}

func (r *testRunner) Log(args ...interface{}) {
	r.log(fmt.Sprint(args...))
}

func (r *testRunner) Logf(format string, args ...interface{}) {
	r.log(fmt.Sprintf(format, args...))
}

// Fail marks the function as having failed but continues execution.
func (r *testRunner) Fail() {
	r.failed = true
}

// FailNow marks the function as having failed and stops its execution
// by calling runtime.Goexit (which then runs all deferred calls in the
// current goroutine).
// Execution will continue at the next test or benchmark.
// FailNow must be called from the goroutine running the
// test or benchmark function, not from other goroutines
// created during the test. Calling FailNow does not stop
// those other goroutines.
func (r *testRunner) FailNow() {
	r.Fail()
	runtime.Goexit()
}

// Fatal is equivalent to Log followed by FailNow.
func (r *testRunner) Fatal(args ...interface{}) {
	r.Log(args...)
	r.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (r *testRunner) Fatalf(format string, args ...interface{}) {
	r.Logf(format, args...)
	r.FailNow()
}

// Error is equivalent to Log followed by Fail.
func (r *testRunner) Error(args ...interface{}) {
	r.Log(args...)
	r.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (r *testRunner) Errorf(format string, args ...interface{}) {
	r.Logf(format, args...)
	r.Fail()
}
