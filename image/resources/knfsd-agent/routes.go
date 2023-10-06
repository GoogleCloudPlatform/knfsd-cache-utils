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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// JSONHandlerFunc is an adapter for functions that return either data or an
// error. The data is JSON encoded and sent as the response to the HTTP request.
//
// All JSON responses should be a JSON object (not a plain string, array, etc)
// so to keep the logic of ServeHTTP simple force the returned value to be a
// pointer. This way ServeHTTP does not need to worry about pointer vs value
// when calling json.MarshalIndent.
type JSONHandlerFunc[T any] func(*http.Request) (*T, error)

// JSONHandler is an adapter to support automatic type inference of T when
// instantiating JSONHandlerFunc[T] because Go does not support type inference
// when instantiating types.
//
// For example, to create a JSONHandlerFunc for handleNodeInfo you would need
// to specify T; `JSONHanderFunc[nodeData](handleNodeInfo)`. Using this function
// Go can infer T automatically; `JSONHandler(handleNodeInfo)`.
func JSONHandler[T any](handler JSONHandlerFunc[T]) http.Handler {
	return handler
}

func (handler JSONHandlerFunc[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statusCode, body, err := handler.Execute(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
	logRequest(r, statusCode, err)
}

// Execute calls the handler and formats the result as JSON.
func (handler JSONHandlerFunc[T]) Execute(r *http.Request) (statusCode int, body []byte, err error) {
	statusCode = http.StatusOK

	if r.Method == http.MethodHead {
		// In case a client relies on sending a HEAD request (e.g. for CORS
		// support) just send a basic empty response.
		// We're not going to try and generate any other headers such as
		// Content-Length as many of the handlers scrape live details on
		// request, and HEAD is supposed to be cheap to call.
		return http.StatusOK, []byte{}, nil
	}

	// Endpoints only support GET requests.
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		statusCode = http.StatusMethodNotAllowed
		body = []byte("{\"message\": \"method not allowed\"}")
		err = errors.New("method not allowed")
		return statusCode, body, err
	}

	result, err := handler(r)
	if err == nil {
		if result == nil {
			// The status endpoints only deal with queries, so there's no valid
			// reason for an endpoint to return 204 No Content. Thus if a handler
			// method returns (nil, nil) then something has gone wrong.
			err = errors.New("handler did not return any content")
		} else {
			// Convert the result to json if there was no error.
			body, err = json.MarshalIndent(result, "", "  ")
		}
	}

	// If there was an error from either the handler, or JSON conversion return
	// a generic error message to the client and log the real error message.
	if err != nil {
		statusCode = http.StatusInternalServerError
		body = []byte("{\"message\": \"An unknown error occurred\"}")
	}

	return statusCode, body, err
}

func logRequest(r *http.Request, statusCode int, err error) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	log.Printf("%s %s %s %d %s", r.RemoteAddr, r.Method, r.URL, statusCode, errMsg)
}

func registerRoutes(mux *http.ServeMux) {
	mux.Handle("/", JSONHandler(handleNodeInfo))
	mux.Handle(fmt.Sprintf("/api/v%s/nodeInfo", apiVersion), JSONHandler(handleNodeInfo))
}
