package verification

import (
	"context"
	"testing"

	"github.com/casselc/goo/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOwnedRuntimeSDKIsolation(t *testing.T) {
	for _, count := range []int{2, 3} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			t.Parallel()
			reader := sdkmetric.NewManualReader()
			mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			defer mp.Shutdown(context.Background())
			exporter := tracetest.NewInMemoryExporter()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			defer tp.Shutdown(context.Background())
			rt, err := instrumentation.NewRuntime(instrumentation.Config{MeterProvider: mp, TracerProvider: tp})
			if err != nil {
				t.Fatal(err)
			}
			for range count {
				ctx, finish := rt.Operation(t.Context(), "same.name", "server")
				rt.Event(ctx, "accepted", func() string { return "subject" }, func() string { return "ready" })
				finish(nil, false)
				finish(nil, true)
			}
			spans := exporter.GetSpans()
			if len(spans) != count {
				t.Fatalf("cross-runtime span count: %d", len(spans))
			}
			for _, span := range spans {
				if span.InstrumentationScope.Name != instrumentation.Scope || len(span.Events) != 1 {
					t.Fatal("scope or event lost")
				}
			}
			var data metricdata.ResourceMetrics
			if err := reader.Collect(t.Context(), &data); err != nil {
				t.Fatal(err)
			}
			calls := int64(0)
			outcomes := map[string]bool{}
			for _, scope := range data.ScopeMetrics {
				for _, m := range scope.Metrics {
					if m.Name == "goo.operation.calls" {
						for _, p := range m.Data.(metricdata.Sum[int64]).DataPoints {
							calls += p.Value
							value, _ := p.Attributes.Value("goo.operation.outcome")
							outcomes[value.AsString()] = true
						}
					}
				}
			}
			if calls != int64(count) || len(outcomes) != 3 {
				t.Fatalf("cross-runtime calls or zero baseline missing: %d %v", calls, outcomes)
			}
		})
	}
}
