/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 */

// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

// MetricSettings provides common settings for a particular metric.
type MetricSettings struct {
	Enabled bool `mapstructure:"enabled"`
}

// MetricsSettings provides settings for slabinfo metrics.
type MetricsSettings struct {
	SlabDentryCacheActiveObjects   MetricSettings `mapstructure:"slab.dentry_cache.active_objects"`
	SlabDentryCacheObjsize         MetricSettings `mapstructure:"slab.dentry_cache.objsize"`
	SlabNfsInodeCacheActiveObjects MetricSettings `mapstructure:"slab.nfs_inode_cache.active_objects"`
	SlabNfsInodeCacheObjsize       MetricSettings `mapstructure:"slab.nfs_inode_cache.objsize"`
}

func DefaultMetricsSettings() MetricsSettings {
	return MetricsSettings{
		SlabDentryCacheActiveObjects: MetricSettings{
			Enabled: true,
		},
		SlabDentryCacheObjsize: MetricSettings{
			Enabled: true,
		},
		SlabNfsInodeCacheActiveObjects: MetricSettings{
			Enabled: true,
		},
		SlabNfsInodeCacheObjsize: MetricSettings{
			Enabled: true,
		},
	}
}

type metricSlabDentryCacheActiveObjects struct {
	data     pdata.Metric   // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills slab.dentry_cache.active_objects metric with initial data.
func (m *metricSlabDentryCacheActiveObjects) init() {
	m.data.SetName("slab.dentry_cache.active_objects")
	m.data.SetDescription("Dentry Cache Active Objects")
	m.data.SetUnit("1")
	m.data.SetDataType(pdata.MetricDataTypeGauge)
}

func (m *metricSlabDentryCacheActiveObjects) recordDataPoint(start pdata.Timestamp, ts pdata.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricSlabDentryCacheActiveObjects) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricSlabDentryCacheActiveObjects) emit(metrics pdata.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricSlabDentryCacheActiveObjects(settings MetricSettings) metricSlabDentryCacheActiveObjects {
	m := metricSlabDentryCacheActiveObjects{settings: settings}
	if settings.Enabled {
		m.data = pdata.NewMetric()
		m.init()
	}
	return m
}

type metricSlabDentryCacheObjsize struct {
	data     pdata.Metric   // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills slab.dentry_cache.objsize metric with initial data.
func (m *metricSlabDentryCacheObjsize) init() {
	m.data.SetName("slab.dentry_cache.objsize")
	m.data.SetDescription("Dentry Cache Object Size")
	m.data.SetUnit("1")
	m.data.SetDataType(pdata.MetricDataTypeGauge)
}

func (m *metricSlabDentryCacheObjsize) recordDataPoint(start pdata.Timestamp, ts pdata.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricSlabDentryCacheObjsize) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricSlabDentryCacheObjsize) emit(metrics pdata.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricSlabDentryCacheObjsize(settings MetricSettings) metricSlabDentryCacheObjsize {
	m := metricSlabDentryCacheObjsize{settings: settings}
	if settings.Enabled {
		m.data = pdata.NewMetric()
		m.init()
	}
	return m
}

type metricSlabNfsInodeCacheActiveObjects struct {
	data     pdata.Metric   // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills slab.nfs_inode_cache.active_objects metric with initial data.
func (m *metricSlabNfsInodeCacheActiveObjects) init() {
	m.data.SetName("slab.nfs_inode_cache.active_objects")
	m.data.SetDescription("NFS inode Cache Cache Active Objects")
	m.data.SetUnit("1")
	m.data.SetDataType(pdata.MetricDataTypeGauge)
}

func (m *metricSlabNfsInodeCacheActiveObjects) recordDataPoint(start pdata.Timestamp, ts pdata.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricSlabNfsInodeCacheActiveObjects) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricSlabNfsInodeCacheActiveObjects) emit(metrics pdata.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricSlabNfsInodeCacheActiveObjects(settings MetricSettings) metricSlabNfsInodeCacheActiveObjects {
	m := metricSlabNfsInodeCacheActiveObjects{settings: settings}
	if settings.Enabled {
		m.data = pdata.NewMetric()
		m.init()
	}
	return m
}

type metricSlabNfsInodeCacheObjsize struct {
	data     pdata.Metric   // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills slab.nfs_inode_cache.objsize metric with initial data.
func (m *metricSlabNfsInodeCacheObjsize) init() {
	m.data.SetName("slab.nfs_inode_cache.objsize")
	m.data.SetDescription("NFS inode Cache Object Size")
	m.data.SetUnit("1")
	m.data.SetDataType(pdata.MetricDataTypeGauge)
}

func (m *metricSlabNfsInodeCacheObjsize) recordDataPoint(start pdata.Timestamp, ts pdata.Timestamp, val int64) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntVal(val)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricSlabNfsInodeCacheObjsize) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricSlabNfsInodeCacheObjsize) emit(metrics pdata.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricSlabNfsInodeCacheObjsize(settings MetricSettings) metricSlabNfsInodeCacheObjsize {
	m := metricSlabNfsInodeCacheObjsize{settings: settings}
	if settings.Enabled {
		m.data = pdata.NewMetric()
		m.init()
	}
	return m
}

// MetricsBuilder provides an interface for scrapers to report metrics while taking care of all the transformations
// required to produce metric representation defined in metadata and user settings.
type MetricsBuilder struct {
	startTime                            pdata.Timestamp
	metricSlabDentryCacheActiveObjects   metricSlabDentryCacheActiveObjects
	metricSlabDentryCacheObjsize         metricSlabDentryCacheObjsize
	metricSlabNfsInodeCacheActiveObjects metricSlabNfsInodeCacheActiveObjects
	metricSlabNfsInodeCacheObjsize       metricSlabNfsInodeCacheObjsize
}

// metricBuilderOption applies changes to default metrics builder.
type metricBuilderOption func(*MetricsBuilder)

// WithStartTime sets startTime on the metrics builder.
func WithStartTime(startTime pdata.Timestamp) metricBuilderOption {
	return func(mb *MetricsBuilder) {
		mb.startTime = startTime
	}
}

func NewMetricsBuilder(settings MetricsSettings, options ...metricBuilderOption) *MetricsBuilder {
	mb := &MetricsBuilder{
		startTime:                            pdata.NewTimestampFromTime(time.Now()),
		metricSlabDentryCacheActiveObjects:   newMetricSlabDentryCacheActiveObjects(settings.SlabDentryCacheActiveObjects),
		metricSlabDentryCacheObjsize:         newMetricSlabDentryCacheObjsize(settings.SlabDentryCacheObjsize),
		metricSlabNfsInodeCacheActiveObjects: newMetricSlabNfsInodeCacheActiveObjects(settings.SlabNfsInodeCacheActiveObjects),
		metricSlabNfsInodeCacheObjsize:       newMetricSlabNfsInodeCacheObjsize(settings.SlabNfsInodeCacheObjsize),
	}
	for _, op := range options {
		op(mb)
	}
	return mb
}

// Emit appends generated metrics to a pdata.MetricsSlice and updates the internal state to be ready for recording
// another set of data points. This function will be doing all transformations required to produce metric representation
// defined in metadata and user settings, e.g. delta/cumulative translation.
func (mb *MetricsBuilder) Emit(metrics pdata.MetricSlice) {
	mb.metricSlabDentryCacheActiveObjects.emit(metrics)
	mb.metricSlabDentryCacheObjsize.emit(metrics)
	mb.metricSlabNfsInodeCacheActiveObjects.emit(metrics)
	mb.metricSlabNfsInodeCacheObjsize.emit(metrics)
}

// RecordSlabDentryCacheActiveObjectsDataPoint adds a data point to slab.dentry_cache.active_objects metric.
func (mb *MetricsBuilder) RecordSlabDentryCacheActiveObjectsDataPoint(ts pdata.Timestamp, val int64) {
	mb.metricSlabDentryCacheActiveObjects.recordDataPoint(mb.startTime, ts, val)
}

// RecordSlabDentryCacheObjsizeDataPoint adds a data point to slab.dentry_cache.objsize metric.
func (mb *MetricsBuilder) RecordSlabDentryCacheObjsizeDataPoint(ts pdata.Timestamp, val int64) {
	mb.metricSlabDentryCacheObjsize.recordDataPoint(mb.startTime, ts, val)
}

// RecordSlabNfsInodeCacheActiveObjectsDataPoint adds a data point to slab.nfs_inode_cache.active_objects metric.
func (mb *MetricsBuilder) RecordSlabNfsInodeCacheActiveObjectsDataPoint(ts pdata.Timestamp, val int64) {
	mb.metricSlabNfsInodeCacheActiveObjects.recordDataPoint(mb.startTime, ts, val)
}

// RecordSlabNfsInodeCacheObjsizeDataPoint adds a data point to slab.nfs_inode_cache.objsize metric.
func (mb *MetricsBuilder) RecordSlabNfsInodeCacheObjsizeDataPoint(ts pdata.Timestamp, val int64) {
	mb.metricSlabNfsInodeCacheObjsize.recordDataPoint(mb.startTime, ts, val)
}

// Reset resets metrics builder to its initial state. It should be used when external metrics source is restarted,
// and metrics builder should update its startTime and reset it's internal state accordingly.
func (mb *MetricsBuilder) Reset(options ...metricBuilderOption) {
	mb.startTime = pdata.NewTimestampFromTime(time.Now())
	for _, op := range options {
		op(mb)
	}
}

// Attributes contains the possible metric attributes that can be used.
var Attributes = struct {
}{}

// A is an alias for Attributes.
var A = Attributes
