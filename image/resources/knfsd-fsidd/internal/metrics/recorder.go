package metrics

import (
	"context"
	"time"
)

type attemptCounter int64

func (c attemptCounter) retries() int64 {
	if c > 0 {
		return int64(c - 1)
	} else {
		return int64(c)
	}
}

type RequestRecorder struct {
	start    time.Time
	command  string
	attempts attemptCounter
}

type OperationRecorder struct {
	start   time.Time
	command string
	attempt attemptCounter
}

func StartRequest(command string) RequestRecorder {
	return RequestRecorder{
		start:   time.Now(),
		command: command,
	}
}

func (rec *RequestRecorder) StartOperation() OperationRecorder {
	rec.attempts++
	return OperationRecorder{
		start:   time.Now(),
		command: rec.command,
		attempt: rec.attempts,
	}
}

func (rec *RequestRecorder) End(ctx context.Context, result string) {
	Request(ctx, rec.command, result, rec.attempts.retries(), time.Since(rec.start))
}

func (rec *OperationRecorder) End(ctx context.Context, result string) {
	Operation(ctx, rec.command, result, rec.attempt.retries(), time.Since(rec.start))
}
