package main

import (
	"context"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestShutdownFlushesQueuedSpansAfterSignal(t *testing.T) {
	exporter := &recordingExporter{InMemoryExporter: tracetest.NewInMemoryExporter()}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Hour)))
	_, span := provider.Tracer("shutdown-test").Start(context.Background(), "last-request")
	span.End()
	// SDK shutdown must use a fresh context, not the cancelled signal context.
	shutdownProvider(provider.Shutdown)
	if len(exporter.flushed) != 1 || exporter.flushed[0].Name != "last-request" {
		t.Fatalf("queued span not flushed: %v", exporter.flushed)
	}
}

type recordingExporter struct {
	*tracetest.InMemoryExporter
	flushed tracetest.SpanStubs
}

func (e *recordingExporter) Shutdown(ctx context.Context) error {
	e.flushed = e.GetSpans()
	return e.InMemoryExporter.Shutdown(ctx)
}

func TestShutdownGetsIndependentLiveContexts(t *testing.T) {
	var first context.Context
	shutdownProvider(func(ctx context.Context) error {
		first = ctx
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		return context.DeadlineExceeded
	})
	shutdownProvider(func(ctx context.Context) error {
		if ctx.Err() != nil || first.Err() == nil {
			t.Fatal("provider budgets not independent")
		}
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 5*time.Second {
			t.Fatal("missing bounded flush")
		}
		return nil
	})
}
