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
