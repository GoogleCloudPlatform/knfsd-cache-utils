package log

import (
	"context"
	"io"
	"log"
	"os"
)

var (
	Debug = log.New(io.Discard, "DEBUG: ", 0)
	Warn  = log.New(os.Stderr, "WARN: ", 0)
	Error = log.New(os.Stderr, "ERROR: ", 0)
)

type contextKey string

const idKey contextKey = "id"

func EnableDebug() {
	Debug.SetOutput(os.Stderr)
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
