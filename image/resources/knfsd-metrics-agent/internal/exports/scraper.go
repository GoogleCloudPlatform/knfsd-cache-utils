package exports

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/convert"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/exports/internal/metadata"
	"github.com/prometheus/procfs/nfs"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type exportScraper struct {
	mb metadata.MetricsBuilder
}

func newScraper(cfg *Config) (scraperhelper.Scraper, error) {
	s := &exportScraper{
		mb: *metadata.NewMetricsBuilder(cfg.Metrics),
	}
	return scraperhelper.NewScraper(typeStr, s.scrape)
}

func (s *exportScraper) scrape(context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()
	now := pdata.NewTimestampFromTime(time.Now())

	fs, err := nfs.NewDefaultFS()
	if err != nil {
		return md, err
	}

	stats, err := fs.ServerRPCStats()
	if err != nil {
		return md, err
	}

	metrics := md.ResourceMetrics().AppendEmpty().
		InstrumentationLibraryMetrics().AppendEmpty().
		Metrics()

	s.mb.RecordNfsExportsTotalReadBytesDataPoint(now, convert.Int64(stats.InputOutput.Read))
	s.mb.RecordNfsExportsTotalWriteBytesDataPoint(now, convert.Int64(stats.InputOutput.Write))
	s.mb.Emit(metrics)
	return md, nil
}
