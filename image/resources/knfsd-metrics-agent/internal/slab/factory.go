package slab

import (
	"context"
	"errors"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/slab/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const typeStr = "slabinfo"

var errWrongConfig = errors.New("config was not a slabinfo receiver config")

func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithMetrics(createMetricsReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ScraperControllerSettings: scraperhelper.DefaultScraperControllerSettings(typeStr),
		Metrics:                   metadata.DefaultMetricsSettings(),
	}
}

func createMetricsReceiver(
	ctx context.Context,
	set component.ReceiverCreateSettings,
	conf config.Receiver,
	consumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return nil, errWrongConfig
	}

	s, err := newScraper(cfg)
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
