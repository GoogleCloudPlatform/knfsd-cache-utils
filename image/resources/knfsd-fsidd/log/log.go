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
