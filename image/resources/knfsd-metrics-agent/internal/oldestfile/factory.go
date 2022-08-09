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

package oldestfile

import (
	"context"
	"errors"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/oldestfile/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const typeStr = "oldestfile"

var errWrongConfig = errors.New("config was not an oldestfile receiver config")

func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithMetrics(createReceiver),
	)
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			ReceiverSettings:   config.NewReceiverSettings(config.NewComponentID(typeStr)),
			CollectionInterval: 10 * time.Minute,
		},
		Metrics:   metadata.DefaultMetricsSettings(),
		CachePath: "/var/cache/fscache/cache",
	}
}

func createReceiver(
	ctx context.Context,
	set component.ReceiverCreateSettings,
	conf config.Receiver,
	consumer consumer.Metrics,

) (component.MetricsReceiver, error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return nil, errWrongConfig
	}

	s, err := newScraper(cfg, set.Logger)
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewScraperControllerReceiver(
		&cfg.ScraperControllerSettings,
		set,
		consumer,
		scraperhelper.AddScraper(s),
	)
}
