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

package mounts

//go:generate go run github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen --experimental-gen metadata.yaml

import (
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/mounts/internal/metadata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Metrics                                 metadata.MetricsSettings `mapstructure:"metrics"`

	// Query the knfsd-agent service to find out which instance a client is
	// connected to, and include an instance attribute in the metrics.
	// This is used to enrich the client metrics with the name of the instance
	// the client is connected to, as the source attribute will only indicate
	// the IP of the load balancer.
	// The load balancer is assumed to have session affinity based upon client
	// IP only, so that all the connections from a client to the same IP will
	// use the same instance.
	QueryProxyInstance QueryProxyInstanceConfig `mapstructure:"query_proxy_instance"`
}

type QueryProxyInstanceConfig struct {
	Enabled bool                             `mapstructure:"enabled"`
	Timeout time.Duration                    `mapstructure:"timeout"`
	Exclude QueryProxyInstanceExcludesConfig `mapstructure:"exclude"`
}

type QueryProxyInstanceExcludesConfig struct {
	Servers    []string `mapstructure:"servers"`
	LocalPaths []string `mapstructure:"local_paths"`
}
