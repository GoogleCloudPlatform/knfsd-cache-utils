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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
)

func main() {
	factories, err := components()
	if err != nil {
		log.Fatalf("failed to build components: %v", err)
	}

	params := service.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "knfsd-metrics-agent",
			Description: "KNFSD Metrics Agent",
		},
		Factories: factories,
	}

	cmd := service.NewCommand(params)
	err = cmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
