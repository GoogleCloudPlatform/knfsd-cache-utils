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

package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONHandler(t *testing.T) {
	type Body struct {
		Message string `json:"msg"`
	}

	execute := func(handler http.Handler, method string) (string, *http.Response) {
		req := httptest.NewRequest(method, "/foo", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		body, _ := io.ReadAll(res.Body)
		return string(body), res
	}

	t.Run("response", func(t *testing.T) {
		handler := JSONHandler(func(*http.Request) (*Body, error) {
			return &Body{"Hello World"}, nil
		})
		body, res := execute(handler, http.MethodGet)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "{\n  \"msg\": \"Hello World\"\n}", body)
	})

	t.Run("nil", func(t *testing.T) {
		handler := JSONHandler(func(*http.Request) (*Body, error) {
			return nil, nil
		})
		body, res := execute(handler, http.MethodGet)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Equal(t, `{"message": "An unknown error occurred"}`, body)
	})

	t.Run("error", func(t *testing.T) {
		handler := JSONHandler(func(*http.Request) (*Body, error) {
			return nil, errors.New("handler generated error")
		})
		body, res := execute(handler, http.MethodGet)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Equal(t, `{"message": "An unknown error occurred"}`, body)
	})

	t.Run("invalid method", func(t *testing.T) {
		// handlers only allow GET requests
		handler := JSONHandler(func(*http.Request) (*Body, error) {
			return &Body{"Hello World"}, nil
		})
		body, res := execute(handler, http.MethodPost)
		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
		assert.Equal(t, `{"message": "method not allowed"}`, body)
	})

	t.Run("HEAD request", func(t *testing.T) {
		// should also support HEAD requests for any endpoint that supports GET
		handler := JSONHandler(func(r *http.Request) (*Body, error) {
			return &Body{Message: "Hello World"}, nil
		})
		body, res := execute(handler, http.MethodHead)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Empty(t, body)
	})
}
