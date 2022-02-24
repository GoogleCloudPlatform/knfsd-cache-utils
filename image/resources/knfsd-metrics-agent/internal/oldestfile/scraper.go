package oldestfile

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/oldestfile/internal/metadata"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

type oldestFileScraper struct {
	mb        *metadata.MetricsBuilder
	cachePath string
	last      oldestFile
}

type oldestFile struct {
	path  string
	mtime time.Time
}

func newScraper(cfg *Config, logger *zap.Logger) (scraperhelper.Scraper, error) {
	s := &oldestFileScraper{
		mb:        metadata.NewMetricsBuilder(cfg.Metrics),
		cachePath: cfg.CachePath,
	}
	return scraperhelper.NewScraper(
		typeStr,
		s.scrape,
	)
}

func (s *oldestFileScraper) scrape(ctx context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()
	age, err := s.findOldest(ctx)
	if err != nil {
		return md, nil
	}

	metrics := md.ResourceMetrics().AppendEmpty().
		InstrumentationLibraryMetrics().AppendEmpty().
		Metrics()

	now := pdata.NewTimestampFromTime(time.Now())
	s.mb.RecordFscacheOldestFileDataPoint(now, int64(age.Seconds()))
	s.mb.Emit(metrics)

	return md, nil
}

func (s *oldestFileScraper) findOldest(ctx context.Context) (time.Duration, error) {
	oldest, err := s.findOldestFile(ctx)
	if err != nil {
		s.last = oldestFile{}
		return 0, err
	}
	s.last = oldest

	if oldest.mtime.IsZero() {
		return 0, nil
	}

	now := time.Now()
	age := now.Sub(oldest.mtime)
	if age < 0 {
		age = 0
	}

	return age, nil
}

func (s *oldestFileScraper) findOldestFile(ctx context.Context) (oldestFile, error) {
	// optimistic check if the oldest file from a previous scrape still exists
	if s.last.path != "" {
		stat, err := os.Stat(s.last.path)
		if err == nil && stat.ModTime() == s.last.mtime {
			// assume the file is still the oldest
			return s.last, nil
		}
	}

	count := 0
	found := oldestFile{}
	err := filepath.WalkDir(s.cachePath, func(path string, d fs.DirEntry, err error) error {
		// Avoiding checking the context on every single file. This is because
		// checking the context has to lock a mutex.
		// No heuristics for a good value here, so just chose 100 arbitrarily.
		count++
		if count > 100 {
			count = 0
			if err := ctx.Err(); err != nil {
				// abort walking the tree with the context's error
				return err
			}
		}

		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			// if there's an error querying file, just skip the file
			return nil
		}

		mtime := info.ModTime()
		if mtime.IsZero() {
			return nil
		}

		if found.mtime.IsZero() || mtime.Before(found.mtime) {
			found = oldestFile{
				path:  path,
				mtime: mtime,
			}
		}

		return nil
	})
	return found, err
}
