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
