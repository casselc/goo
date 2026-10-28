package scenario

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/instrumentation"
	"github.com/casselc/goo/verification/weaving/example"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"os"
	"path/filepath"
	"testing"
)

func attr(attrs []attribute.KeyValue, key string) string {
	for _, a := range attrs {
		if string(a.Key) == key {
			return a.Value.AsString()
		}
	}
	return ""
}
func TestGooAnnotationsEmitActualOTel(t *testing.T) {
	manifest, err := annotations.ScanDir("../example")
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Declarations) != 4 {
		t.Fatal("incomplete manifest")
	}
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(context.Background())
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	defer mp.Shutdown(context.Background())
	rt, err := instrumentation.NewRuntime(instrumentation.Config{MeterProvider: mp, TracerProvider: tp})
	if err != nil {
		t.Fatal(err)
	}
	if err := instrumentation.BindDefault(rt); err != nil {
		t.Fatal(err)
	}
	result, err := example.Reserve(context.Background(), false)
	if err != nil || result.ID != "reservation-1" {
		t.Fatal("application result changed")
	}
	result, err = example.Reserve(context.Background(), true)
	if err == nil || result != nil {
		t.Fatal("application error changed")
	}
	func() {
		defer func() {
			if recover() != "original panic" {
				t.Error("panic was swallowed or changed")
			}
		}()
		example.Panic(context.Background())
	}()
	// Invalid selector/subject and absent recording parent diagnose drops, without changing the function's return.
	if err = example.Complete(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err = example.Complete(context.Background(), &example.Decision{ID: ""}); err != nil {
		t.Fatal(err)
	}
	if err = example.Complete(context.Background(), &example.Decision{ID: "outside"}); err != nil {
		t.Fatal(err)
	}
	spans := exporter.GetSpans()
	if len(spans) != 3 {
		t.Fatalf("expected three woven spans, got %d; build with orchestrion", len(spans))
	}
	observations := []annotations.Observation{}
	for _, span := range spans {
		observations = append(observations, annotations.Observation{
			Resource: "woven-checkout", Scope: span.InstrumentationScope.Name,
			ScopeVersion: span.InstrumentationScope.Version, SchemaURL: span.InstrumentationScope.SchemaURL,
			TraceID: span.SpanContext.TraceID().String(), SpanID: span.SpanContext.SpanID().String(),
			Name: attr(span.Attributes, "goo.operation.name"), Role: attr(span.Attributes, "goo.operation.role"),
			Outcome: attr(span.Attributes, "goo.operation.outcome"), DurationNS: uint64(span.EndTime.Sub(span.StartTime)),
		})
	}
	overview, err := manifest.Observe("woven-checkout", "actual-sdk-export", observations)
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Operations) != 2 || overview.Operations[0].Panics != 1 || overview.Operations[1].Calls != 2 || overview.Operations[1].Errors != 1 {
		t.Fatalf("wrong observed overview: %+v", overview)
	}
	events := 0
	for _, span := range spans {
		for _, event := range span.Events {
			events++
			if attr(event.Attributes, "goo.subject.id") != "reservation-1" {
				t.Fatal("wrong subject")
			}
		}
		for _, a := range span.Attributes {
			if a.Value.AsString() == "private failure detail" {
				t.Fatal("raw error exposed")
			}
		}
	}
	if events != 2 {
		t.Fatalf("events %d", events)
	}
	var data metricdata.ResourceMetrics
	if err = reader.Collect(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	outcomes := map[string]int64{}
	activePoints := 0
	durationCount := uint64(0)
	drops := int64(0)
	for _, scope := range data.ScopeMetrics {
		if scope.Scope.Name != manifest.Scope || scope.Scope.Version != manifest.ScopeVersion {
			t.Fatal("scope contract mismatch")
		}
		for _, m := range scope.Metrics {
			expectedUnit := ""
			for _, signal := range manifest.Signals {
				if signal.Name == m.Name {
					expectedUnit = signal.Unit
				}
			}
			if expectedUnit == "" || m.Unit != expectedUnit {
				t.Fatalf("undeclared metric/unit: %s %s", m.Name, m.Unit)
			}
			switch m.Name {
			case "goo.operation.calls":
				for _, p := range m.Data.(metricdata.Sum[int64]).DataPoints {
					out, _ := p.Attributes.Value("goo.operation.outcome")
					outcomes[out.AsString()] += p.Value
					if _, found := p.Attributes.Value("goo.subject.id"); found {
						t.Fatal("subject became metric dimension")
					}
				}
			case "goo.operation.active":
				for _, p := range m.Data.(metricdata.Sum[int64]).DataPoints {
					activePoints++
					if p.Value != 0 {
						t.Fatal("active invocation leaked")
					}
				}
			case "goo.operation.duration":
				for _, p := range m.Data.(metricdata.Histogram[float64]).DataPoints {
					durationCount += p.Count
				}
			case "goo.instrumentation.dropped_events":
				for _, p := range m.Data.(metricdata.Sum[int64]).DataPoints {
					drops += p.Value
				}
			}
		}
	}
	if outcomes["ok"] != 1 || outcomes["error"] != 1 || outcomes["panic"] != 1 || durationCount != 3 || activePoints != 2 || drops != 3 {
		t.Fatalf("outcomes=%v duration=%d active=%d drops=%d", outcomes, durationCount, activePoints, drops)
	}
	directory := os.Getenv("GOO_ANNOTATION_ARTIFACTS")
	if directory != "" {
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		normalized, err := json.MarshalIndent(observations, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(directory, "operations.json"), normalized, 0644); err != nil {
			t.Fatal(err)
		}
		payload, err := json.MarshalIndent(struct {
			Manifest any `json:"manifest"`
			Spans    any `json:"spans"`
			Metrics  any `json:"metrics"`
		}{manifest, spans, data}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(directory, "otel.json"), payload, 0644); err != nil {
			t.Fatal(err)
		}
	}

}
