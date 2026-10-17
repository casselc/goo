// Package instrumentation supplies opt-in OTel hooks for goo's weaving aspects.
package instrumentation

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"sync"
	"time"
)

const Scope = "github.com/casselc/goo/instrumentation"
const Version = "0.1.0"
const SchemaURL = "goo://instrumentation/v1"

var once sync.Once
var initializedOutcomes sync.Map
var calls metric.Int64Counter
var duration metric.Float64Histogram
var active metric.Int64UpDownCounter
var dropped metric.Int64Counter

func instruments() {
	once.Do(func() {
		m := otel.Meter(Scope, metric.WithInstrumentationVersion(Version), metric.WithSchemaURL(SchemaURL))
		calls, _ = m.Int64Counter("goo.operation.calls", metric.WithUnit("{invocation}"))
		duration, _ = m.Float64Histogram("goo.operation.duration", metric.WithUnit("s"), metric.WithExplicitBucketBoundaries(0, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30))
		active, _ = m.Int64UpDownCounter("goo.operation.active", metric.WithUnit("{invocation}"))
		dropped, _ = m.Int64Counter("goo.instrumentation.dropped_events", metric.WithUnit("{event}"))
	})
}
func Operation(ctx context.Context, name, role string) (context.Context, func(error, bool)) {
	instruments()
	attrs := []attribute.KeyValue{attribute.String("goo.operation.name", name), attribute.String("goo.operation.role", role)}
	// Establish explicit observed-zero series for the finite outcome domain.
	// Do not infer zero for operations that have never run.
	init, _ := initializedOutcomes.LoadOrStore(struct{ name, role string }{name, role}, &sync.Once{})
	init.(*sync.Once).Do(func() {
		for _, outcome := range []string{"ok", "error", "panic"} {
			a := append(append([]attribute.KeyValue{}, attrs...), attribute.String("goo.operation.outcome", outcome))
			calls.Add(ctx, 0, metric.WithAttributes(a...))
		}
	})
	kind := trace.SpanKindInternal
	switch role {
	case "server":
		kind = trace.SpanKindServer
	case "client":
		kind = trace.SpanKindClient
	case "producer":
		kind = trace.SpanKindProducer
	case "consumer":
		kind = trace.SpanKindConsumer
	}
	ctx, span := otel.Tracer(Scope, trace.WithInstrumentationVersion(Version), trace.WithSchemaURL(SchemaURL)).Start(ctx, name, trace.WithSpanKind(kind), trace.WithAttributes(attrs...))
	started := time.Now()
	active.Add(ctx, 1, metric.WithAttributes(attrs...))
	var finish sync.Once
	return ctx, func(err error, panicked bool) {
		finish.Do(func() {
			defer span.End()
			active.Add(ctx, -1, metric.WithAttributes(attrs...))
			outcome := "ok"
			if panicked {
				outcome = "panic"
				span.SetStatus(codes.Error, "operation panic")
				span.SetAttributes(attribute.String("error.type", "panic"))
			} else if err != nil {
				outcome = "error"
				span.SetStatus(codes.Error, "operation error")
				span.SetAttributes(attribute.String("error.type", reflect.TypeOf(err).String()))
			}
			a := append(append([]attribute.KeyValue{}, attrs...), attribute.String("goo.operation.outcome", outcome))
			calls.Add(ctx, 1, metric.WithAttributes(a...))
			duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(a...))
			span.SetAttributes(attribute.String("goo.operation.outcome", outcome))
		})
	}
}
func read(get func() string) (value string, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return get(), true
}

type subjectScopeKey struct{}
type subjectScope struct{ namespace, tenant string }

// WithSubjectScope declares the namespace and tenant used for subsequent subject
// events. It does not establish durability or global identity across services.
func WithSubjectScope(ctx context.Context, namespace, tenant string) context.Context {
	return context.WithValue(ctx, subjectScopeKey{}, subjectScope{namespace, tenant})
}
func Event(ctx context.Context, name string, subject, outcome func() string) {
	instruments()
	drop := func(reason string) {
		dropped.Add(ctx, 1, metric.WithAttributes(attribute.String("goo.event.name", name), attribute.String("goo.drop.reason", reason)))
	}
	id, ok := read(subject)
	if !ok {
		drop("selector-panic")
		return
	}
	if id == "" || len(id) > 256 {
		drop("invalid-subject")
		return
	}
	value, ok := read(outcome)
	if !ok || len(value) > 128 {
		drop("invalid-outcome")
		return
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		drop("no-recording-span")
		return
	}
	attrs := []attribute.KeyValue{attribute.String("goo.event.name", name), attribute.String("goo.subject.id", id)}
	if value != "" {
		attrs = append(attrs, attribute.String("goo.subject.outcome", value))
	}
	if scope, found := ctx.Value(subjectScopeKey{}).(subjectScope); found {
		if scope.namespace == "" || scope.tenant == "" || len(scope.namespace) > 128 || len(scope.tenant) > 128 {
			drop("invalid-subject-scope")
			return
		}
		attrs = append(attrs, attribute.String("goo.subject.namespace", scope.namespace), attribute.String("goo.subject.tenant", scope.tenant))
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}
