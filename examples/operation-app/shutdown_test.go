package main

import (
	"context"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"sync/atomic"
	"testing"
	"time"
)

type finalMetricExporter struct {
	count   int64
	stopped bool
}

func (*finalMetricExporter) Temporality(k sdkmetric.InstrumentKind) metricdata.Temporality {
	return sdkmetric.DefaultTemporalitySelector(k)
}
func (*finalMetricExporter) Aggregation(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(k)
}
func (e *finalMetricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
				for _, p := range sum.DataPoints {
					e.count += p.Value
				}
			}
		}
	}
	return ctx.Err()
}
func (*finalMetricExporter) ForceFlush(ctx context.Context) error { return ctx.Err() }
func (e *finalMetricExporter) Shutdown(ctx context.Context) error { e.stopped = true; return ctx.Err() }
func TestShutdownExportsFinalMetrics(t *testing.T) {
	exporter := &finalMetricExporter{}
	var enabled atomic.Bool
	enabled.Store(true)
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(demoExporter{exporter, &enabled}, sdkmetric.WithInterval(time.Hour))))
	counter, err := provider.Meter("shutdown").Int64Counter("requests")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(context.Background(), 1)
	shutdownProvider(provider.Shutdown)
	if exporter.count != 1 || !exporter.stopped {
		t.Fatalf("final metric missing or exporter not stopped: %+v", exporter)
	}
}
