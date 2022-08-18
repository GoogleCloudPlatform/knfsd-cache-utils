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

package log

import (
	"io"
	"log"
)

var (
	debug = log.New(io.Discard, "DEBUG: ", 0)
	info  = log.New(log.Writer(), "INFO: ", 0)
	error = log.New(log.Writer(), "ERROR: ", 0)
	fatal = log.New(log.Writer(), "ERROR: ", 0)
)

func EnableDebug() {
	debug.SetOutput(log.Writer())
}

func Debug(v ...interface{}) {
	debug.Print(v...)
}

func Debugf(format string, v ...interface{}) {
	debug.Printf(format, v...)
}

func Info(v ...interface{}) {
	info.Print(v...)
}

func Infof(format string, v ...interface{}) {
	info.Printf(format, v...)
}

func Error(v ...interface{}) {
	error.Print(v...)
}

func Errorf(format string, v ...interface{}) {
	error.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	fatal.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	fatal.Fatalf(format, v...)
}
