package connections

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/connections/internal/metadata"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type scraper struct {
	mb *metadata.MetricsBuilder
}

func newScraper(cfg *Config) (scraperhelper.Scraper, error) {
	s := &scraper{
		mb: metadata.NewMetricsBuilder(cfg.Metrics),
	}
	return scraperhelper.NewScraper(typeStr, s.scrape)
}

func (s *scraper) scrape(ctx context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()
	now := pdata.NewTimestampFromTime(time.Now())

	// TODO: Consider counting unique clients (by IP) as well as total connections
	count, err := countConnectedClients(ctx)
	if err != nil {
		return md, err
	}

	metrics := md.ResourceMetrics().AppendEmpty().
		InstrumentationLibraryMetrics().AppendEmpty().
		Metrics()

	s.mb.RecordNfsConnectionsDataPoint(now, count)
	s.mb.Emit(metrics)
	return md, nil
}

func countConnectedClients(ctx context.Context) (int64, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ss",
		"--no-header", "--oneline", "--numeric",
		"--tcp", "--udp",
		"state", "established",
		"sport", "2049",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			err = fmt.Errorf("command terminated with exit code %d\n%s", exit.ExitCode(), stderr.String())
		}
		return 0, err
	}

	var count int64
	s := bufio.NewScanner(&stdout)
	for s.Scan() {
		count++
	}

	err = s.Err()
	if err != nil {
		return 0, err
	}

	return count, nil
}
