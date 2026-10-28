package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// A discovered binding is a candidate. Actual observations determine whether
// evaluation succeeds, stays unknown, or refuses the entire bounded input.
func TestDiscoveryCandidatesDoNotVerifyEvaluation(t *testing.T) {
	for _, metric := range []string{"http.server.request.duration", "messaging.process.duration", "custom.duration"} {
		for _, unit := range []string{"s", "ms"} {
			t.Run(metric+"/"+unit, func(t *testing.T) {
				dq := DiscoveryQuestion{Environment: "test", Service: "ordinary", FromMS: 1000, ToMS: 3000}
				resource := map[string]string{"service.name": "ordinary", "service.instance.id": "one", "deployment.environment.name": "test"}
				obs := map[string]any{"kind": "histogram", "scope": "ordinary", "metric": metric, "unit": unit, "temporality": 1, "resource": resource, "attributes": map[string]string{}}
				inventory, err := discoverRows(dq, []json.RawMessage{raw(obs)})
				if err != nil {
					t.Fatal(err)
				}
				i := inventory.Instruments[0]
				before := Hash(i)
				if !i.CapabilityPlan.Eligible || i.CapabilityPlan.Verification != "not-run" {
					t.Fatal(i)
				}
				q := Question{Definition: *i.Definition, FromMS: 1000, ToMS: 3000}
				row := metricRow{Resource: resource, Scope: "ordinary", Name: metric, Unit: unit, Attributes: map[string]string{}, StartNS: ns(1000), EndNS: ns(3000), Temporality: 1, Count: "4", Bounds: []float64{1}, BucketCounts: []string{"1", "3"}}
				check := func(q Question, rows []json.RawMessage, state, refusal string) {
					t.Helper()
					result, e := Evaluate(q, rows, time.UnixMilli(4000))
					if refusal != "" {
						if e == nil || result != nil || !strings.Contains(e.Error(), refusal) {
							t.Fatalf("expected %s refusal, got %+v %v", refusal, result, e)
						}
					} else {
						if e != nil || result.State != state {
							t.Fatalf("want %s: %+v %v", state, result, e)
						}
						if e = ValidateResult(result); e != nil {
							t.Fatal(e)
						}
					}
					if Hash(i) != before {
						t.Fatal("evaluation mutated discovery candidate")
					}
				}
				check(q, []json.RawMessage{raw(row)}, "observed", "")
				partial := row
				partial.EndNS = ns(2000)
				check(q, []json.RawMessage{raw(partial)}, "unknown", "")
				flagged := row
				flagged.Flags = 1
				check(q, []json.RawMessage{raw(flagged)}, "unknown", "")
				malformed := row
				malformed.Count = "5"
				check(q, []json.RawMessage{raw(malformed)}, "", "bucket counts")
				missing := row
				missing.Resource = cloneLabels(resource)
				delete(missing.Resource, "service.instance.id")
				check(q, []json.RawMessage{raw(missing)}, "", "identity/unit/flags")
				check(q, []json.RawMessage{raw(row), raw(missing)}, "", "identity/unit/flags")
				missing.Resource["service.instance.id"] = ""
				check(q, []json.RawMessage{raw(missing)}, "", "identity/unit/flags")
				mixed := row
				mixed.StartNS = ns(500)
				mixed.EndNS = ns(2500)
				mixed.Temporality = 2
				check(q, []json.RawMessage{raw(mixed), raw(row)}, "", "mixed histogram temporality")
				check(q, nil, "no_data", "")
				q.FromMS = 1500
				check(q, []json.RawMessage{raw(row)}, "unknown", "")
			})
		}
	}
}
