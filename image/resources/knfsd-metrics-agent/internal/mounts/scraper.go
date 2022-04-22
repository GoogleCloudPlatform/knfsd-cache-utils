package mounts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/convert"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/mounts/internal/metadata"
	"github.com/prometheus/procfs"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

type mountScraper struct {
	cfg      *Config
	logger   *zap.Logger
	p        procfs.Proc
	mb       *metadata.MetricsBuilder
	nic      *nodeInfoClient
	previous map[string]nfsMount
}

type nfsMount struct {
	device string
	mount  string

	server string
	path   string

	// included if QueryInstanceName is true
	instance string

	stats   *procfs.MountStatsNFS
	summary summary
}

type nodeInfoClient http.Client

type op struct {
	rttPerOp float64
	exePerOp float64
}

type nodeInfo struct {
	Name string `json:"name"`
}

func newScraper(cfg *Config, logger *zap.Logger) (scraperhelper.Scraper, error) {
	s := &mountScraper{
		cfg:    cfg,
		mb:     metadata.NewMetricsBuilder(cfg.Metrics),
		nic:    createNodeInfoClient(cfg),
		logger: logger,
	}
	return scraperhelper.NewScraper(
		typeStr,
		s.scrape,
		scraperhelper.WithStart(s.start),
	)
}

func createNodeInfoClient(cfg *Config) *nodeInfoClient {
	if !cfg.QueryProxyInstance.Enabled {
		return nil
	}
	return &nodeInfoClient{
		Timeout: cfg.QueryProxyInstance.Timeout,
	}
}

func (s *mountScraper) start(context.Context, component.Host) error {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return err
	}

	p, err := fs.Self()
	if err != nil {
		return err
	}

	s.p = p
	return nil
}

func (s *mountScraper) scrape(context.Context) (pdata.Metrics, error) {
	md := pdata.NewMetrics()
	rms := md.ResourceMetrics()
	now := pdata.NewTimestampFromTime(time.Now())

	mounts, err := s.findNFSMounts()
	if err != nil {
		return md, err
	}
	s.queryInstanceNames(mounts)

	for _, mount := range mounts {
		metrics := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty().Metrics()
		s.report(mount, now, metrics)
	}
	s.track(mounts)

	return md, nil
}

func (s *mountScraper) findNFSMounts() ([]nfsMount, error) {
	mounts, err := s.p.MountStats()
	if err != nil {
		return []nfsMount{}, err
	}

	// estimate the capacity based on the previous run
	cap := len(s.previous) + 10
	if cap < 20 {
		cap = 20
	}
	nfsMounts := make([]nfsMount, 0, cap)

	for _, mount := range mounts {
		if mount.Type == "nfs" || mount.Type == "nfs4" {
			if stats, ok := mount.Stats.(*procfs.MountStatsNFS); ok {
				server, path := splitNFSDevice(mount.Device)
				nfsMounts = append(nfsMounts, nfsMount{
					device:  mount.Device,
					mount:   mount.Mount,
					server:  server,
					path:    path,
					stats:   stats,
					summary: newSummary(stats),
				})
			}
		}
	}

	return nfsMounts, nil
}

func (s *mountScraper) queryInstanceNames(mounts []nfsMount) {
	if !s.cfg.QueryProxyInstance.Enabled {
		s.logger.Debug("not resolving instance names, QueryProxyInstance is disabled")
		return
	}

	s.logger.Debug("querying instance names")

	// Emulate a set using a map, create a distinct list of servers
	servers := make(map[string]struct{}, 10)
	for _, mount := range mounts {
		if mount.server != "" {
			servers[mount.server] = struct{}{}
		}
	}

	// TODO: Consider optimising this by running queries in parallel.
	// TODO: Exponential backoff (per server) if a query keeps failing.
	instances := make(map[string]string, len(servers))
	for server := range servers {
		instance, err := s.nic.queryInstanceName(server)
		if err != nil {
			s.logger.Warn("failed to query instance name", zap.String("server", server))
		} else if instance == "" {
			s.logger.Warn("instance name resolved as empty string", zap.String("server", server))
		} else {
			s.logger.Debug("resolved proxy instance", zap.String("server", server), zap.String("instance", instance))
			instances[server] = instance
		}
	}

	for idx := range mounts {
		mount := &mounts[idx]
		if instance, ok := instances[mount.server]; ok {
			mount.instance = instance
		} else if previous, ok := s.previous[mount.device]; ok {
			// In case the lookup failed due to a transient error assume the
			// instance name has not changed since the last scrape.
			mount.instance = previous.instance
		}
	}
}

func (c *nodeInfoClient) queryInstanceName(addr string) (string, error) {
	// this shouldn't happen, but just in case do not cause a panic
	if c == nil {
		return "", nil
	}

	var err error
	url := fmt.Sprintf("http://%s/api/v1.0/nodeInfo", addr)
	resp, err := (*http.Client)(c).Get(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var info nodeInfo
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&info)
	if err != nil {
		return "", err
	}

	return info.Name, nil
}

func (s *mountScraper) report(mount nfsMount, now pdata.Timestamp, metrics pdata.MetricSlice) {
	s.mb.RecordNfsMountReadBytesDataPoint(now, convert.Int64(mount.stats.Bytes.ReadTotal), mount.server, mount.path, mount.instance)
	s.mb.RecordNfsMountWriteBytesDataPoint(now, convert.Int64(mount.stats.Bytes.WriteTotal), mount.server, mount.path, mount.instance)

	for _, op := range mount.stats.Operations {
		s.mb.RecordNfsMountOperationRequestsDataPoint(now, convert.Int64(op.Requests), mount.server, mount.path, mount.instance, op.Operation)
		s.mb.RecordNfsMountOperationSentBytesDataPoint(now, convert.Int64(op.BytesSent), mount.server, mount.path, mount.instance, op.Operation)
		s.mb.RecordNfsMountOperationReceivedBytesDataPoint(now, convert.Int64(op.BytesReceived), mount.server, mount.path, mount.instance, op.Operation)
		s.mb.RecordNfsMountOperationMajorTimeoutsDataPoint(now, convert.Int64(op.MajorTimeouts), mount.server, mount.path, mount.instance, op.Operation)
		s.mb.RecordNfsMountOperationErrorsDataPoint(now, convert.Int64(op.Errors), mount.server, mount.path, mount.instance, op.Operation)
	}

	// report original delta based metrics
	s.reportDelta(mount, now)

	s.mb.Emit(metrics)
}

func (s *mountScraper) reportDelta(mount nfsMount, now pdata.Timestamp) {
	// TODO: Report counters instead of gauges

	// This is a direct port of the original script that used
	// nfsiostat, so is limited to the data that was output by nfsiostat.
	// Now that we have the raw data we can report a lot more metrics.
	// We can also use counters instead of calculating the diff for many of
	// the metrics and let the monitoring tool calculate the rate.

	// This is a new mount, no previous metrics yet so cannot calculate a diff
	// on this scrape.
	prev, found := s.previous[mount.device]
	if !found {
		return
	}

	// if new.age < prev.summary.age, then the counters must have reset
	// i.e. the NFS share was re-mounted.
	// If the counters have reset then we cannot derive any useful metrics
	// on this scrape, so just treat this as if it were a new summary.
	if mount.summary.age <= prev.summary.age {
		return
	}

	diff := diffSummary(mount.summary, prev.summary)

	delta := diff.age.Seconds()
	if delta <= 0 {
		// This should not happen, as any summaries that have an age smaller
		// than the previous entry are handled as a reset to the counters.
		// Just in case, skip reporting this scrape to avoid divide by zero
		// errors or spurious values.
		return
	}

	sends := float64(diff.transport.Sends)
	var backlog float64
	if sends > 0 {
		backlog = float64(diff.transport.CumulativeBacklog) / sends
	}
	read := calc(delta, diff.read)
	write := calc(delta, diff.write)

	s.mb.RecordNfsMountOpsPerSecondDataPoint(now, sends/delta, mount.server, mount.path, mount.instance)
	s.mb.RecordNfsMountRPCBacklogDataPoint(now, backlog/delta, mount.server, mount.path, mount.instance)
	s.mb.RecordNfsMountReadExeDataPoint(now, read.exePerOp, mount.server, mount.path, mount.instance)
	s.mb.RecordNfsMountReadRttDataPoint(now, read.rttPerOp, mount.server, mount.path, mount.instance)
	s.mb.RecordNfsMountWriteExeDataPoint(now, write.exePerOp, mount.server, mount.path, mount.instance)
	s.mb.RecordNfsMountWriteRttDataPoint(now, write.rttPerOp, mount.server, mount.path, mount.instance)
}

func calc(delta float64, diff procfs.NFSOperationStats) op {
	ops := float64(diff.Requests)
	// retrans := float64(diff.Transmissions) - float64(diff.Requests)
	// kilobytes := float64(diff.BytesSent+diff.BytesReceived) / 1024
	// queuedFor := float64(diff.CumulativeQueueMilliseconds)
	rtt := float64(diff.CumulativeTotalRequestMilliseconds)
	exe := float64(diff.CumulativeTotalResponseMilliseconds)
	// errs := float64(diff.Errors)

	// var kbPerOp, retransPercent, queuedForPerOp, errsPercent float64
	var rttPerOp, exePerOp float64
	if diff.Requests > 0 {
		// kbPerOp = kilobytes / ops
		// retransPercent = (retrans * 100) / ops
		rttPerOp = rtt / ops
		exePerOp = exe / ops
		// queuedForPerOp = queuedFor / ops
		// errsPercent = (errs * 100) / ops
	}

	return op{rttPerOp, exePerOp}
}

func (s *mountScraper) track(mounts []nfsMount) {
	previous := make(map[string]nfsMount, len(mounts))
	for _, m := range mounts {
		previous[m.device] = m
	}
	s.previous = previous
}

func splitNFSDevice(s string) (server string, path string) {
	parts := strings.SplitN(s, ":", 2)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		// shouldn't happen, do our best to return something sensible
		return "", parts[0]
	default:
		return parts[0], parts[1]
	}
}
