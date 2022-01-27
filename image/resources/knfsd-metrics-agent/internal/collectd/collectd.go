package collectd

import (
	"context"
	"log"
	"time"

	"collectd.org/api"
	"collectd.org/exec"
	"collectd.org/meta"
)

var (
	hostname = exec.Hostname()
	interval = exec.Interval()
	writer   = exec.Putval

	EnableMeta bool

	// TODO: Review the names of the metric types
	// changing these now would break existing dashboards
	MetaMetricType = "stackdriver_metric_type"
	MetaMountName  = "label:mount_name"
)

// Keeping the plugin as exec for consistency with original metrics.
// If changing this you might need to update the re-write rules in knfsd.conf
// to match if using collectd < 5.11.0
const plugin = "exec"

func UseUnixSocket(path string) {
	writer = newUnix(path)
	// plugin = "unixsock"
}

func Interval() time.Duration {
	return interval
}

type Gauge struct {
	vl *api.ValueList
}

func NewGauge(pluginInstance, typeInstance string, m meta.Data) Gauge {
	if EnableMeta {
		metricType := meta.String("custom.googleapis.com/knfsd/" + pluginInstance)
		if m == nil {
			m = meta.Data{
				MetaMetricType: metricType,
			}
		} else {
			m = m.Clone()
			if _, found := m[MetaMetricType]; !found {
				m[MetaMetricType] = metricType
			}
		}
	} else {
		m = nil
	}

	vl := &api.ValueList{
		Identifier: api.Identifier{
			Host:           hostname,
			Plugin:         plugin,
			PluginInstance: pluginInstance,
			Type:           "gauge",
			TypeInstance:   typeInstance,
		},
		Interval: interval,
		Values:   make([]api.Value, 1),
		Meta:     m,
	}

	return Gauge{vl}
}

func (g Gauge) Write(ctx context.Context, now time.Time, value float64) {
	g.vl.Time = now
	g.vl.Values[0] = api.Gauge(value)
	err := writer.Write(ctx, g.vl)
	if err != nil {
		log.Printf("WARN: %s: %s\n", g.vl.Identifier, err)
	}
}
