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
	"fmt"
	"log"
	"net/http"
)

// Custom HandlerFunc that returns the status code and any errors
type HandlerFunc func(http.ResponseWriter, *http.Request) (int, error)

// routes() holds all of the server routes
func (s *server) routes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handlerWrapper(handleProcess))                                          // Handler for /api/{version}/nodeinfo
	mux.HandleFunc(fmt.Sprintf("/api/v%s/nodeInfo", apiVersion), s.handlerWrapper(handleProcess)) // Handler for /api/{version}/nodeinfo
}

// handleWrapper is a middleware function to handle common tasks such as setting response headers
func (s *server) handlerWrapper(h HandlerFunc) http.HandlerFunc {

	// Return the hander func
	return func(w http.ResponseWriter, r *http.Request) {

		// Set the Content-Type headers
		w.Header().Set("Content-Type", "application/json")

		// Handle Request
		status, err := h(w, r)

		// Prep Error Message
		var errMsg string

		// Handle error
		if err != nil {
			w.Write([]byte("{\"message\": \"An unknown error occurred\"}"))
			errMsg = err.Error()
		}

		// Log Request
		log.Printf("%s %s %s %d %s", r.RemoteAddr, r.Method, r.URL, status, errMsg)

	}

}
