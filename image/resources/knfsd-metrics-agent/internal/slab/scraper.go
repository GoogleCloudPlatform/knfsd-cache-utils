package slab

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/slab/internal/metadata"
	"github.com/prometheus/procfs"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type slabScraper struct {
	fs procfs.FS
	mb *metadata.MetricsBuilder
}

func newScraper(cfg *Config) (scraperhelper.Scraper, error) {
	s := &slabScraper{
		mb: metadata.NewMetricsBuilder(cfg.Metrics),
	}
	return scraperhelper.NewScraper(
		typeStr,
		s.scrape,
		scraperhelper.WithStart(s.start))
}

func (s *slabScraper) start(context.Context, component.Host) error {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return err
	}

	// Verify we can scrape slabinfo, most common reason this will fail is
	// because of permissions (needs to be root).
	_, err = fs.SlabInfo()
	// TODO: check for transient errors?
	if err != nil {
		return err
	}

	s.fs = fs
	return nil
}

func (s *slabScraper) scrape(ctx context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()

	info, err := s.fs.SlabInfo()
	if err != nil {
		return md, err
	}

	now := pdata.NewTimestampFromTime(time.Now())
	metrics := md.ResourceMetrics().AppendEmpty().
		InstrumentationLibraryMetrics().AppendEmpty().
		Metrics()

	dentry := find(info.Slabs, "dentry")
	if dentry != nil {
		s.mb.RecordSlabDentryCacheActiveObjectsDataPoint(now, dentry.ObjActive)
		s.mb.RecordSlabDentryCacheObjsizeDataPoint(now, dentry.ObjSize)
	}

	nfs := find(info.Slabs, "nfs_inode_cache")
	if nfs != nil {
		s.mb.RecordSlabNfsInodeCacheActiveObjectsDataPoint(now, nfs.ObjActive)
		s.mb.RecordSlabNfsInodeCacheObjsizeDataPoint(now, nfs.ObjSize)
	}

	s.mb.Emit(metrics)
	return md, nil
}

func find(slabs []*procfs.Slab, name string) *procfs.Slab {
	for _, s := range slabs {
		if s.Name == name {
			return s
		}
	}
	return nil
}
