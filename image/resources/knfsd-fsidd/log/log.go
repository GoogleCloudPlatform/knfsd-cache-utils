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
	"io"
	"log"
	"os"
)

var (
	Debug = log.New(io.Discard, "DEBUG: ", 0)
	Info  = log.New(os.Stderr, "INFO: ", 0)
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
