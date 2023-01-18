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
	"context"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-systemd/v22/journal"
)

type any = interface{}

type Logger interface {
	Print(v ...any)
	Printf(format string, v ...any)
}

type SystemdLogger struct {
	Priority journal.Priority
}

func (l *SystemdLogger) Print(v ...any) {
	msg := fmt.Sprint(v...)
	journal.Send(msg, l.Priority, nil)
}

func (l *SystemdLogger) Printf(format string, v ...any) {
	journal.Print(l.Priority, format, v...)
}

type Discard struct{}

func (*Discard) Print(v ...any)                 {}
func (*Discard) Printf(format string, v ...any) {}

type contextKey string

const idKey contextKey = "id"

var (
	useSystemd bool
	Debug      Logger = &Discard{}
	Info       Logger
	Warn       Logger
	Error      Logger
)

func init() {
	stderrIsJournalStream, _ := journal.StderrIsJournalStream()
	useSystemd = stderrIsJournalStream && journal.Enabled()

	if useSystemd {
		Info = &SystemdLogger{journal.PriInfo}
		Warn = &SystemdLogger{journal.PriWarning}
		Error = &SystemdLogger{journal.PriErr}
	} else {
		Info = log.New(os.Stderr, "INFO: ", 0)
		Warn = log.New(os.Stderr, "WARN: ", 0)
		Error = log.New(os.Stderr, "ERROR: ", 0)
	}
}

func EnableDebug() {
	if useSystemd {
		Debug = &SystemdLogger{journal.PriDebug}
	} else {
		Debug = log.New(os.Stderr, "DEBUG: ", 0)
	}
}

func WithID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, idKey, id)
}

func ID(ctx context.Context) uint64 {
	val := ctx.Value(idKey)
	if id, ok := val.(uint64); ok {
		return id
	}
	return 0
}
