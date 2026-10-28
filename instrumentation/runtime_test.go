package instrumentation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type testProvider struct {
	metric.MeterProvider
	meter metric.Meter
}

func (p testProvider) Meter(string, ...metric.MeterOption) metric.Meter { return p.meter }

type testMeter struct {
	metric.Meter
	fail    string
	counter *testCounter
}

var instrumentError = errors.New("instrument creation failed")

func (m testMeter) Int64Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	if name == m.fail {
		return nil, instrumentError
	}
	if name == "goo.operation.calls" && m.counter != nil {
		return m.counter, nil
	}
	return m.Meter.Int64Counter(name, opts...)
}
func (m testMeter) Int64UpDownCounter(name string, opts ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	if name == m.fail {
		return nil, instrumentError
	}
	return m.Meter.Int64UpDownCounter(name, opts...)
}
func (m testMeter) Float64Histogram(name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	if name == m.fail {
		return nil, instrumentError
	}
	return m.Meter.Float64Histogram(name, opts...)
}

type testCounter struct {
	metric.Int64Counter
	zeros, observations atomic.Int64
}

func (c *testCounter) Add(_ context.Context, n int64, _ ...metric.AddOption) {
	if n == 0 {
		c.zeros.Add(1)
	} else {
		c.observations.Add(n)
	}
}
func testConfig(fail string) (Config, *testCounter) {
	provider := metricnoop.NewMeterProvider()
	meter := provider.Meter("test")
	counter, _ := meter.Int64Counter("calls")
	counts := &testCounter{Int64Counter: counter}
	return Config{MeterProvider: testProvider{provider, testMeter{meter, fail, counts}}, TracerProvider: tracenoop.NewTracerProvider()}, counts
}
func TestRuntimeInitializationErrors(t *testing.T) {
	for _, name := range []string{"goo.operation.calls", "goo.operation.duration", "goo.operation.active", "goo.instrumentation.dropped_events"} {
		t.Run(name, func(t *testing.T) {
			config, _ := testConfig(name)
			rt, err := NewRuntime(config)
			if rt != nil || !errors.Is(err, instrumentError) || !strings.Contains(err.Error(), name) {
				t.Fatalf("lost initialization error: %v %v", rt, err)
			}
		})
	}
	config, _ := testConfig("")
	config.MeterProvider = nil
	if _, err := NewRuntime(config); err == nil {
		t.Fatal("nil meter accepted")
	}
	config, _ = testConfig("")
	config.TracerProvider = nil
	if _, err := NewRuntime(config); err == nil {
		t.Fatal("nil tracer accepted")
	}
}
func TestOwnedRuntimesConcurrentAccounting(t *testing.T) {
	for range 2 {
		t.Run("isolated", func(t *testing.T) {
			t.Parallel()
			config, counts := testConfig("")
			rt, err := NewRuntime(config)
			if err != nil {
				t.Fatal(err)
			}
			var group sync.WaitGroup
			for range 100 {
				group.Go(func() {
					_, finish := rt.Operation(t.Context(), "same.operation", "internal")
					var finishes sync.WaitGroup
					for range 4 {
						finishes.Go(func() { finish(nil, false) })
					}
					finishes.Wait()
				})
			}
			group.Wait()
			if counts.zeros.Load() != 3 || counts.observations.Load() != 100 {
				t.Fatalf("seeds=%d calls=%d", counts.zeros.Load(), counts.observations.Load())
			}
		})
	}
}
func TestOutcomeInitializationBound(t *testing.T) {
	config, counts := testConfig("")
	rt, err := NewRuntime(config)
	if err != nil {
		t.Fatal(err)
	}
	for i := range maxInitializedOperations + 1 {
		_, finish := rt.Operation(t.Context(), fmt.Sprint(i), "internal")
		finish(nil, false)
	}
	if len(rt.initialized) != maxInitializedOperations || counts.zeros.Load() != 3*maxInitializedOperations || counts.observations.Load() != maxInitializedOperations+1 {
		t.Fatal("bookkeeping exceeded bound or actual observation lost")
	}
}
func TestDefaultBinding(t *testing.T) {
	if os.Getenv("GOO_TEST_DEFAULT_BINDING") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestDefaultBinding$")
		cmd.Env = append(os.Environ(), "GOO_TEST_DEFAULT_BINDING=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("binding subprocess: %v\n%s", err, out)
		}
		return
	}
	_, finish := Operation(t.Context(), "unbound", "internal")
	finish(nil, false)
	Event(t.Context(), "unbound", func() string { t.Fatal("unbound selector evaluated"); return "" }, nil)
	if BindDefault(nil) == nil || BindDefault(&Runtime{}) == nil {
		t.Fatal("invalid runtime accepted")
	}
	config, counts := testConfig("")
	rt, err := NewRuntime(config)
	if err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int64
	var group sync.WaitGroup
	for range 20 {
		group.Go(func() {
			if BindDefault(rt) == nil {
				successes.Add(1)
			}
		})
	}
	group.Wait()
	if successes.Load() != 1 {
		t.Fatal("binding not exclusive")
	}
	_, finish = Operation(t.Context(), "bound", "internal")
	finish(nil, false)
	if counts.observations.Load() != 1 {
		t.Fatal("woven bridge did not use bound runtime")
	}
	other, _ := NewRuntime(config)
	if BindDefault(other) == nil {
		t.Fatal("rebinding accepted")
	}
}
