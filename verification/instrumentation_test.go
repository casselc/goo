package verification

import (
	"context"
	"fmt"
	"github.com/casselc/goo/instrumentation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"hegel.dev/go/hegel"
	"testing"
)

type invocation struct {
	end  func(error, bool)
	done bool
}
type operationMachine struct {
	reader      *metric.ManualReader
	invocations []invocation
}

func (m *operationMachine) RuleStart(tc hegel.TestCase) {
	_, end := instrumentation.Operation(context.Background(), "property.balance", "internal")
	m.invocations = append(m.invocations, invocation{end: end})
}
func (m *operationMachine) finish(tc hegel.TestCase, mode string) {
	if len(m.invocations) == 0 {
		return
	}
	i := hegel.Draw(tc, hegel.Integers(0, len(m.invocations)-1))
	entry := &m.invocations[i]
	var err error
	if mode == "error" {
		err = fmt.Errorf("controlled error")
	}
	entry.end(err, mode == "panic")
	entry.done = true
	if mode == "double" {
		entry.end(nil, false)
	}
}
func (m *operationMachine) RuleFinish(tc hegel.TestCase)         { m.finish(tc, "ok") }
func (m *operationMachine) RuleError(tc hegel.TestCase)          { m.finish(tc, "error") }
func (m *operationMachine) RulePanic(tc hegel.TestCase)          { m.finish(tc, "panic") }
func (m *operationMachine) RuleRepeatedFinish(tc hegel.TestCase) { m.finish(tc, "double") }
func (m *operationMachine) InvariantActive(tc hegel.TestCase) {
	var snapshot metricdata.ResourceMetrics
	if err := m.reader.Collect(context.Background(), &snapshot); err != nil {
		panic(err)
	}
	expected := int64(0)
	for _, entry := range m.invocations {
		if !entry.done {
			expected++
		}
	}
	observed := int64(0)
	outcomes := map[string]bool{}
	for _, scope := range snapshot.ScopeMetrics {
		for _, signal := range scope.Metrics {
			if signal.Name == "goo.operation.calls" {
				for _, point := range signal.Data.(metricdata.Sum[int64]).DataPoints {
					v, _ := point.Attributes.Value("goo.operation.name")
					if v.AsString() == "property.balance" {
						o, _ := point.Attributes.Value("goo.operation.outcome")
						outcomes[o.AsString()] = true
					}
				}
			}
			if signal.Name == "goo.operation.active" {
				for _, point := range signal.Data.(metricdata.Sum[int64]).DataPoints {
					observed += point.Value
				}
			}
		}
	}
	if len(m.invocations) > 0 && (!outcomes["ok"] || !outcomes["error"] || !outcomes["panic"]) {
		panic("missing explicit outcome baselines")
	}
	if observed != expected {
		panic(fmt.Sprintf("active: observed %d expected %d", observed, expected))
	}
}
func TestHegelOperationAccountingStateSwarm(t *testing.T) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	otel.SetMeterProvider(provider)
	defer provider.Shutdown(context.Background())
	hegel.Test(t, func(tc *hegel.T) {
		m := &operationMachine{reader: reader}
		defer func() {
			for _, entry := range m.invocations {
				entry.end(nil, false)
			}
		}()
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantActive"))
	}, options(100)...)
}
