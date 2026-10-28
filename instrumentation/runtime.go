// Package instrumentation supplies opt-in OTel hooks for goo's weaving aspects.
package instrumentation

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const Scope = "github.com/casselc/goo/instrumentation"
const Version = "0.1.0"
const SchemaURL = "goo://instrumentation/v1"

// Config selects the application's providers. Both are required; callers that
// deliberately disable a signal can supply its OTel noop provider.
type Config struct {
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
}

// Runtime owns instruments and outcome initialization. Do not copy it after use.
// Providers and their shutdown remain owned by the application.
type Runtime struct {
	tracer      trace.Tracer
	calls       metric.Int64Counter
	duration    metric.Float64Histogram
	active      metric.Int64UpDownCounter
	dropped     metric.Int64Counter
	mu          sync.Mutex
	initialized map[operationKey]*sync.Once
}
type operationKey struct{ name, role string }

const maxInitializedOperations = 1024

// NewRuntime creates instruments eagerly and returns the first creation error.
// It never installs a global provider, binds woven hooks or shuts down providers.
func NewRuntime(config Config) (*Runtime, error) {
	if config.MeterProvider == nil || config.TracerProvider == nil {
		return nil, fmt.Errorf("instrumentation: meter and tracer providers are required")
	}
	m := config.MeterProvider.Meter(Scope, metric.WithInstrumentationVersion(Version), metric.WithSchemaURL(SchemaURL))
	rt := &Runtime{tracer: config.TracerProvider.Tracer(Scope, trace.WithInstrumentationVersion(Version), trace.WithSchemaURL(SchemaURL)), initialized: make(map[operationKey]*sync.Once)}
	var err error
	rt.calls, err = m.Int64Counter("goo.operation.calls", metric.WithUnit("{invocation}"))
	if err != nil {
		return nil, fmt.Errorf("goo.operation.calls: %w", err)
	}
	rt.duration, err = m.Float64Histogram("goo.operation.duration", metric.WithUnit("s"), metric.WithExplicitBucketBoundaries(0, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30))
	if err != nil {
		return nil, fmt.Errorf("goo.operation.duration: %w", err)
	}
	rt.active, err = m.Int64UpDownCounter("goo.operation.active", metric.WithUnit("{invocation}"))
	if err != nil {
		return nil, fmt.Errorf("goo.operation.active: %w", err)
	}
	rt.dropped, err = m.Int64Counter("goo.instrumentation.dropped_events", metric.WithUnit("{event}"))
	if err != nil {
		return nil, fmt.Errorf("goo.instrumentation.dropped_events: %w", err)
	}
	return rt, nil
}

var defaultRuntime atomic.Pointer[Runtime]

// BindDefault binds woven hooks once for this process. It cannot be reset or
// replaced. Tests and manual integrations needing isolation use Runtime methods.
func BindDefault(rt *Runtime) error {
	if rt == nil || rt.tracer == nil {
		return fmt.Errorf("instrumentation: constructed runtime required")
	}
	if !defaultRuntime.CompareAndSwap(nil, rt) {
		return fmt.Errorf("instrumentation: default runtime already bound")
	}
	return nil
}

// Operation is the weaving bridge. Before BindDefault it records nothing.
func Operation(ctx context.Context, name, role string) (context.Context, func(error, bool)) {
	if rt := defaultRuntime.Load(); rt != nil {
		return rt.Operation(ctx, name, role)
	}
	return ctx, func(error, bool) {}
}

// Event is the weaving bridge; unbound calls do not evaluate selectors.
func Event(ctx context.Context, name string, subject, outcome func() string) {
	if rt := defaultRuntime.Load(); rt != nil {
		rt.Event(ctx, name, subject, outcome)
	}
}

// Operation starts an invocation whose finalizer is safe to call repeatedly
// or concurrently. Use finite operation names and roles, never request identifiers.
func (rt *Runtime) Operation(ctx context.Context, name, role string) (context.Context, func(error, bool)) {
	attrs := []attribute.KeyValue{attribute.String("goo.operation.name", name), attribute.String("goo.operation.role", role)}
	// Establish explicit observed-zero series for the finite outcome domain.
	// Do not infer zero for operations that have never run.
	rt.mu.Lock()
	key := operationKey{name, role}
	init := rt.initialized[key]
	if init == nil && len(rt.initialized) < maxInitializedOperations {
		init = &sync.Once{}
		rt.initialized[key] = init
	}
	rt.mu.Unlock()
	// Bound bookkeeping without evicting already initialized operations. Beyond
	// the limit, actual observations continue but no synthetic zero baselines are seeded.
	if init != nil {
		init.Do(func() {
			for _, outcome := range []string{"ok", "error", "panic"} {
				a := append(append([]attribute.KeyValue{}, attrs...), attribute.String("goo.operation.outcome", outcome))
				rt.calls.Add(ctx, 0, metric.WithAttributes(a...))
			}
		})
	}

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
	ctx, span := rt.tracer.Start(ctx, name, trace.WithSpanKind(kind), trace.WithAttributes(attrs...))
	started := time.Now()
	rt.active.Add(ctx, 1, metric.WithAttributes(attrs...))
	var finish sync.Once
	return ctx, func(err error, panicked bool) {
		finish.Do(func() {
			defer span.End()
			rt.active.Add(ctx, -1, metric.WithAttributes(attrs...))
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
			rt.calls.Add(ctx, 1, metric.WithAttributes(a...))
			rt.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(a...))
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

// Event adds a subject observation to the supplied recording span. Selector
// failures are diagnosed with this runtime's dropped-event counter.
func (rt *Runtime) Event(ctx context.Context, name string, subject, outcome func() string) {
	drop := func(reason string) {
		rt.dropped.Add(ctx, 1, metric.WithAttributes(attribute.String("goo.event.name", name), attribute.String("goo.drop.reason", reason)))
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
