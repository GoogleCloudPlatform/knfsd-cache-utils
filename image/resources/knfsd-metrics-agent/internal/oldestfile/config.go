package oldestfile

//go:generate go run github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen --experimental-gen metadata.yaml

import (
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/oldestfile/internal/metadata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Metrics                                 metadata.MetricsSettings `mapstructure:"metrics"`
	CachePath                               string                   `mapstructure:"cache_path"`
}
