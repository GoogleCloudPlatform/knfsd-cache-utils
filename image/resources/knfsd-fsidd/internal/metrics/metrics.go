package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

var meter = global.Meter("fsid")

var (
	dimensionless = instrument.WithUnit(unit.Dimensionless)
	milliseconds  = instrument.WithUnit(unit.Milliseconds)
)

var (
	requestCount    = counter("fsid.request.count", dimensionless)
	requestDuration = duration("fsid.request.duration", milliseconds)
	requestRetries  = int64Histogram("fsid.request.retries", dimensionless)

	operationCount    = counter("fsid.operation.count", dimensionless)
	operationDuration = duration("fsid.operation.duration", milliseconds)

	sqlQueryCount    = counter("fsid.sql.query.count", dimensionless)
	sqlQueryDuration = duration("fsid.sql.query.duration", milliseconds)
)

func Request(ctx context.Context, command, result string, retries int64, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("command", command),
		attribute.String("result", result),
	}
	requestCount.Add(ctx, 1, attrs...)
	requestDuration.Record(ctx, ms(duration), attrs...)
	requestRetries.Record(ctx, retries, attrs...)
}

func Operation(ctx context.Context, command, result string, retry int64, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("command", command),
		attribute.String("result", result),
		attribute.Int64("retry", retry),
	}
	operationCount.Add(ctx, 1, attrs...)
	operationDuration.Record(ctx, ms(duration), attrs...)
}

func SQLOperation(ctx context.Context, query, result string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("query", query),
		attribute.String("result", result),
	}
	sqlQueryCount.Add(ctx, 1, attrs...)
	sqlQueryDuration.Record(ctx, ms(duration), attrs...)
}

func counter(name string, opts ...instrument.Int64Option) instrument.Int64Counter {
	m, err := meter.Int64Counter(name, opts...)
	if err != nil {
		otel.Handle(err)
	}
	return m
}

func duration(name string, opts ...instrument.Float64Option) instrument.Float64Histogram {
	m, err := meter.Float64Histogram(name, opts...)
	if err != nil {
		otel.Handle(err)
	}
	return m
}

func int64Histogram(name string, opts ...instrument.Int64Option) instrument.Int64Histogram {
	m, err := meter.Int64Histogram(name, opts...)
	if err != nil {
		otel.Handle(err)
	}
	return m
}

func ms(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}
