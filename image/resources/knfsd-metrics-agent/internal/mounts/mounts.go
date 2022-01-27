package mounts

import (
	"context"
	"strings"
	"time"

	"collectd.org/meta"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/collectd"
	"github.com/prometheus/procfs"
)

type entry struct {
	seen     bool     // flag for simple mark and sweep gc of old entries
	summary  summary  // previous summary, used to calculate diff
	counters counters // collectd counters for the mount
}

type counters struct {
	ops     collectd.Gauge // operations per second
	backlog collectd.Gauge // average RPC backlog length
	read    operationCounters
	write   operationCounters
}

type operationCounters struct {
	// average round-trip time, the duration from when the client's
	// kernel sends the RPC request until the time it receives the reply.
	rtt collectd.Gauge
	// average execution time, the duration from when the NFS client
	// sends the NFS request to its kernel until the RPC request is completed.
	// This includes the RTT time.
	exe collectd.Gauge
}

var previous = make(map[string]*entry, 1000)

func clear() {
	for _, v := range previous {
		v.seen = false
	}
}

func sweep() {
	for k, v := range previous {
		if !v.seen {
			delete(previous, k)
		}
	}
}

func Report(ctx context.Context, p procfs.Proc) error {
	mounts, err := p.MountStats()
	if err != nil {
		return err
	}

	clear()
	for _, m := range mounts {
		if m.Type == "nfs" || m.Type == "nfs4" {
			if s, ok := m.Stats.(*procfs.MountStatsNFS); ok {
				reportMount(ctx, m, s)
			}
		}
	}
	sweep()

	return nil
}

func reportMount(ctx context.Context, m *procfs.Mount, s *procfs.MountStatsNFS) {
	prev, found := previous[m.Device]
	new := newSummary(s)

	if found {
		// if new.age < prev.summary.age, then the counters must have reset
		// i.e. the NFS share was re-mounted.
		// If the counters have reset then we cannot derive any useful metrics
		// on this cycle, so just treat this as if it were a new summary.

		if new.age > prev.summary.age {
			diff := diffSummary(new, prev.summary)
			reportDiff(ctx, m, prev.counters, diff)
		}

		prev.seen = true
		prev.summary = new
	} else {
		// no previous entry, just record the new value
		previous[m.Device] = newEntry(m, new)
	}
}

func newEntry(m *procfs.Mount, new summary) *entry {
	return &entry{
		seen:     true,
		summary:  new,
		counters: newCounters(m),
	}
}

func newCounters(mount *procfs.Mount) counters {
	device := mangle(mount.Device)
	m := meta.Data{
		// TODO: change this to the non-mangled device name
		// keeping this the same for now to avoid changing any existing dashboards
		collectd.MetaMountName: meta.String(device),
	}
	return counters{
		ops:     collectd.NewGauge("nfsiostat_ops_per_second", device, m),
		backlog: collectd.NewGauge("nfsiostat_rpc_backlog", device, m),
		read: operationCounters{
			rtt: collectd.NewGauge("nfsiostat_mount_read_rtt", device, m),
			exe: collectd.NewGauge("nfsiostat_mount_read_exe", device, m),
		},
		write: operationCounters{
			rtt: collectd.NewGauge("nfsiostat_mount_write_rtt", device, m),
			exe: collectd.NewGauge("nfsiostat_mount_write_exe", device, m),
		},
	}
}

func reportDiff(ctx context.Context, m *procfs.Mount, c counters, diff summary) {
	// TODO: Report counters instead of gauges
	// TODO: Improve the metrics we're reporting
	// TODO: Report on other operations, such as LOOKUPS

	// When improving the metrics collected, need to consider the volume
	// of data being produced. Perhaps make the metrics configurable.
	// Alternatively summarise some metrics, such as instead of reporting
	// every individual operation, just include a single summarised value
	// for "metadata".

	// Currently this is a direct port of the original script that used
	// nfsiostat, so is limited to the data that was output by nfsiostat.
	// Now that we have the raw data we can report a lot more metrics.
	// We can also use counters instead of calculating the diff for many of
	// the metrics and let the monitoring tool calculate the rate.

	delta := diff.age.Seconds()
	if delta <= 0 {
		// This should not happen, as any summaries that have an age smaller
		// than the previous entry are handled as a reset to the counters.
		// Just in case, skip reporting this cycle to avoid divide by zero
		// errors or spurious values.
		return
	}

	now := time.Now()
	sends := float64(diff.transport.Sends)

	var backlog float64
	if sends > 0 {
		backlog = float64(diff.transport.CumulativeBacklog) / sends
	}

	c.ops.Write(ctx, now, sends/delta)
	c.backlog.Write(ctx, now, backlog/delta)
	reportOp(ctx, now, delta, diff.read, c.read)
	reportOp(ctx, now, delta, diff.write, c.write)
}

func reportOp(
	ctx context.Context,
	now time.Time,
	delta float64,
	diff procfs.NFSOperationStats,
	c operationCounters,
) {
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

	c.rtt.Write(ctx, now, rttPerOp)
	c.exe.Write(ctx, now, exePerOp)
}

func mangle(s string) string {
	return strings.ReplaceAll(s, "/", "$")
}

func find(ops []procfs.NFSOperationStats, name string) (procfs.NFSOperationStats, bool) {
	for _, o := range ops {
		if o.Operation == name {
			return o, true
		}
	}
	return procfs.NFSOperationStats{}, false
}
