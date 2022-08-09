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
	"log"
	"net/http"
	"os"
)

const (
	apiVersion = "1.0"
)

var (
	nodeInfo nodeData
)

// Fetch information on the Knfsd node
func init() {

	// Populate Node Info
	nodeInfo = nodeData{}
	err := nodeInfo.fetch()
	if err != nil {
		log.Fatal(err)
	}

	// Create Logging Directory if it does not exist
	err = os.MkdirAll("/var/log/knfsd-agent", os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating logging directory: %s", err.Error())
	}

	// Setup Logging
	file, err := os.OpenFile("/var/log/knfsd-agent/agent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error creating logging file: %s", err.Error())
	}
	log.SetOutput(file)

}

// Define a server type
type server struct{}

func main() {

	// Create a HTTP Mux
	mux := http.NewServeMux()

	// Create Server
	s := server{}

	// Register all API Routes
	s.routes(mux)

	// Listen and Serve
	log.Println("Knfsd Agent is listening on web server port 80...")
	http.ListenAndServe(":80", mux)

}
