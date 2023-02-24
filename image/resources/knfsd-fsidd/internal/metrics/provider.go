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

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-fsidd/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Config struct {
	Enabled  bool          `ini:"enabled"`
	Endpoint string        `ini:"endpoint"`
	Insecure bool          `ini:"insecure"`
	Interval time.Duration `ini:"interval"`
}

type Provider interface {
	Shutdown(context.Context) error
}

type empty struct{}

func (empty) Shutdown(context.Context) error {
	return nil
}

func Start(ctx context.Context, cfg Config) Provider {
	var err error = nil
	if !cfg.Enabled {
		return empty{}
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName("knfsd-fsidd"),
		),
	)
	if err != nil {
		log.Warn.Printf("could not load all otel resources: %v", err)
	}

	exporter, err := newExporter(ctx, cfg)
	if err != nil {
		log.Warn.Printf("could not initialize metric exporter: %v", err)
		return empty{}
	}

	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(cfg.Interval))
	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
	)

	global.SetMeterProvider(provider)
	return provider
}

func newExporter(ctx context.Context, cfg Config) (metric.Exporter, error) {
	var opts []otlpmetricgrpc.Option
	if cfg.Endpoint != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	return exporter, err
}
