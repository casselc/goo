package verification

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

func TestHegelPoolSameRowWitnessAndDetachedDomain(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 100000))
		scope := exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}
		i := hdInstrument("pool", "db.client.connection.count", "stock", "{connection}", scope, "api")
		i.Temporality = 2
		i.SchemaURL = measurements.PoolConnectionsSchema
		i.Definition.Source.SchemaURL = i.SchemaURL
		stream := metrics.Stream{Resource: map[string]string{"deployment.environment.name": "prod", "service.name": "api", "service.instance.id": fmt.Sprintf("worker-%d", n)}, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Name: i.Metric, Unit: i.Unit}
		for _, pool := range []string{"", fmt.Sprintf("pool-%d", n)} {
			for _, state := range []string{"", "used", "custom-state"} {
				for _, split := range []bool{false, true} {
					s := stream
					s.Attributes = map[string]string{"db.client.connection.pool.name": pool, "db.client.connection.state": state}
					streams := []measurements.ObservedStream{{Stream: s}}
					if split {
						a, b := s, s
						a.Attributes = map[string]string{"db.client.connection.pool.name": pool}
						b.Attributes = map[string]string{"db.client.connection.state": state}
						streams = []measurements.ObservedStream{{Stream: a}, {Stream: b}}
					}
					candidate := i
					candidate.CapabilityPlan = &measurements.InstrumentCapabilityPlan{Coverage: measurements.InstrumentCoverage{Streams: streams}}
					want := pool != "" && state != "" && !split
					if measurements.PoolConnectionsCandidate(candidate) != want {
						tc.Errorf("same-row witness mismatch")
						return
					}
					if !want {
						continue
					}
					before, _ := json.Marshal(candidate)
					plan, e := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "auto", Scope: scope, Service: "api", Signals: []string{"metrics"}, Instruments: []measurements.Instrument{candidate}}, nil)
					after, _ := json.Marshal(candidate)
					if e != nil || string(before) != string(after) || len(plan.NativeOptions) != 1 || plan.NativeOptions[0].Instrument.Definition.StockDomain != "nonnegative-integer" {
						tc.Errorf("detached domain %v", e)
						return
					}
					derived := *candidate.Definition
					derived.StockDomain = "nonnegative-integer"
					if measurements.Hash(derived) != plan.NativeOptions[0].Key {
						tc.Errorf("selection hash differs from executed definition")
						return
					}
				}
			}
		}
	}, options(12)...)
}
func TestPoolDomainRawRecordedValuesAndReplay(t *testing.T) {
	scope := exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}
	i := hdInstrument("pool", "db.client.connection.count", "stock", "{connection}", scope, "api")
	i.Definition.StockDomain = "nonnegative-integer"
	q := measurements.Question{Definition: *i.Definition, FromMS: 100000, ToMS: 160000}
	for _, c := range []struct {
		value string
		flags int
		valid bool
	}{{"0", 0, true}, {"9007199254740991", 0, true}, {"-1", 0, false}, {"0.5", 0, false}, {"9007199254740992", 0, false}, {"9007199254740991.1", 0, false}, {"1e1", 0, true}, {"0x1p0", 0, false}, {"-1", 1, true}} {
		row := map[string]any{"resource": map[string]string{"service.name": "api", "service.instance.id": "one", "deployment.environment.name": "prod"}, "scope": i.Scope, "version": i.Version, "schema_url": i.SchemaURL, "name": i.Metric, "unit": i.Unit, "attributes": map[string]string{}, "start_ns": "90000000000", "end_ns": "120000000000", "value": c.value, "flags": c.flags, "temporality": 2, "monotonic": false}
		raw, _ := json.Marshal(row)
		r, e := measurements.Evaluate(q, []json.RawMessage{raw}, time.Unix(200, 0), "", nil)
		if (e == nil) != c.valid {
			t.Fatal("recorded domain", c, e)
		}
		if e == nil && measurements.ValidateResult(r) != nil {
			t.Fatal("valid domain replay")
		}
	}
}
