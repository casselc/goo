package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

var rpcKeys362 = []string{"error.type", "rpc.method", "rpc.status_code", "rpc.system.name"}

func rpcPresence362(labels map[string]string, key string) string {
	v, ok := labels[key]
	if !ok {
		return "missing"
	}
	if v == "" {
		return "empty"
	}
	return "present"
}
func rpcObservationKey362(resource, attributes map[string]string, schema string) string {
	return continuationJSON352([]any{schema, resource, attributes})
}
func rpcCoverageCheck362(i measurements.Instrument, observations []measurements.DiscoveryObservation, p measurements.InstrumentCapabilityPlan) error {
	type group struct {
		resource, attributes map[string]string
		schema               string
		records              int
	}
	groups := map[string]*group{}
	for _, o := range observations {
		if o.Resource == nil || o.Attributes == nil {
			if p.Coverage.State != "unavailable" || p.Coverage.TotalStreams != 0 || len(p.Coverage.Streams) != 0 {
				return fmt.Errorf("malformed observations became partial coverage")
			}
			return nil
		}
		key := rpcObservationKey362(o.Resource, o.Attributes, o.ResourceSchemaURL)
		if groups[key] == nil {
			groups[key] = &group{o.Resource, o.Attributes, o.ResourceSchemaURL, 0}
		}
		groups[key].records++
	}
	n := len(groups)
	if n == 0 {
		if p.Coverage.State != "unavailable" || p.Coverage.TotalStreams != 0 {
			return fmt.Errorf("no observations became known presence")
		}
		return nil
	}
	c := p.Coverage
	if c.State != "observed" || c.TotalStreams != n || len(c.Streams) != min(n, 25) || c.OmittedStreams != max(0, n-25) || len(c.Dimensions) != 4 {
		return fmt.Errorf("display cap changed all-stream cardinality")
	}
	for index, key := range rpcKeys362 {
		d := c.Dimensions[index]
		want := map[string]int{}
		for _, g := range groups {
			want[rpcPresence362(g.attributes, key)]++
		}
		if d.Key != key || d.Present != want["present"] || d.Empty != want["empty"] || d.Missing != want["missing"] {
			return fmt.Errorf("independent literal key coverage %s", key)
		}
	}
	seen := map[string]bool{}
	last := ""
	for _, s := range c.Streams {
		key := rpcObservationKey362(s.Stream.Resource, s.Stream.Attributes, s.Stream.ResourceSchemaURL)
		g := groups[key]
		if g == nil || seen[key] || s.Records != g.records || s.Instance != rpcPresence362(g.resource, "service.instance.id") || !serviceContinuationHash352(s.ID) || s.ID <= last {
			return fmt.Errorf("reported exact stream identity/count/instance")
		}
		last = s.ID
		seen[key] = true
		if s.Stream.Scope != i.Scope || s.Stream.Version != i.Version || s.Stream.SchemaURL != i.SchemaURL || s.Stream.Name != i.Metric || s.Stream.Unit != i.Unit || len(s.Dimensions) != 4 {
			return fmt.Errorf("stream metadata converted")
		}
		for j, key := range rpcKeys362 {
			if s.Dimensions[j].Key != key || s.Dimensions[j].State != rpcPresence362(g.attributes, key) {
				return fmt.Errorf("per-stream conditional presence")
			}
		}
	}
	return nil
}

// Expected execution eligibility uses all raw resource identities independently
// of the instrument plan and its bounded coverage preview. Unknown resources do
// not establish an absent instance; a known missing or empty value does.
func rpcExpectedEligibility362(i measurements.Instrument, observations []measurements.DiscoveryObservation) bool {
	if !i.Supported || i.Definition == nil {
		return false
	}
	if i.Definition.Source.Resource != nil {
		return true
	}
	for _, observation := range observations {
		if observation.Resource != nil && observation.Resource["service.instance.id"] == "" {
			return false
		}
	}
	return true
}

func rpcObservations362(n, seed int) []measurements.DiscoveryObservation {
	out := []measurements.DiscoveryObservation{}
	for k := range n {
		resource := map[string]string{"service.name": "api"}
		switch k % 3 {
		case 1:
			resource["service.instance.id"] = ""
		case 2:
			resource["service.instance.id"] = fmt.Sprintf("node%d", k)
		}
		attributes := map[string]string{"custom.extra": strconv.Itoa(k), "rpc.system": "legacy"}
		for j, key := range rpcKeys362 {
			switch (k + j + seed) % 4 {
			case 0:
				attributes[key] = ""
			case 1:
				attributes[key] = "_OTHER"
			case 2:
				attributes[key] = "literal.雪/value"
			}
		}
		schema := ""
		if k%2 == 1 {
			schema = "https://example.invalid/resource"
		}
		o := measurements.DiscoveryObservation{Resource: resource, Attributes: attributes, ResourceSchemaURL: schema}
		for repeat := 0; repeat < 1+k%3; repeat++ {
			out = append(out, o)
		}
	}
	return out
}
func TestHegelRPCCapability362(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		n := hegel.Draw(tc, hegel.Integers(0, 48))
		observations := rpcObservations362(n, seed)
		scope := exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000}
		i := hdInstrument("rpc-metadata", "rpc.server.call.duration", "histogram", "s", scope, "api")
		i.SchemaURL = []string{"", "https://opentelemetry.io/schemas/1.45.0", "https://example.invalid/custom-schema"}[seed%3]
		i.Definition.Source.SchemaURL = i.SchemaURL
		before := savedJSON360([]any{i, observations})
		p := measurements.PlanInstrument(i, observations)
		if p.Profile != "rpc-server-call-duration" || p.RuleSet != "OpenTelemetry semantic conventions v1.45.0" || !p.ShapeMatches || p.Eligible != rpcExpectedEligibility362(i, observations) || p.Verification != "not-run" {
			tc.Fatal("literal current RPC metadata candidate", p)
		}
		if err := rpcCoverageCheck362(i, observations, p); err != nil {
			tc.Fatal(err)
		}
		reverse := slices.Clone(observations)
		slices.Reverse(reverse)
		if savedJSON360(p) != savedJSON360(measurements.PlanInstrument(i, reverse)) {
			tc.Fatal("raw observation order changed metadata")
		}
		if savedJSON360([]any{i, observations}) != before {
			tc.Fatal("planning mutated input labels or binding")
		}
		// Shape is only literal name/type/unit. Generic binding eligibility remains a
		// separate capability: optional dimensions/schema do not establish instance
		// identity, and known invalid instances remain an execution-policy refusal.
		for variant := range 8 {
			item := i
			definition := *i.Definition
			item.Definition = &definition
			wantProfile := "rpc-server-call-duration"
			switch variant {
			case 1:
				item.Unit = "ms"
				item.Definition.Source.Unit = "ms"
			case 2:
				item.Kind = "gauge"
				item.Temporality = 0
				item.Definition.Kind = "gauge"
			case 3:
				item.Metric = "rpc.server.duration"
				item.Unit = "ms"
				item.Definition.Source.Metric = item.Metric
				item.Definition.Source.Unit = "ms"
				wantProfile = "generic"
			case 4:
				item.Metric = "custom.rpc.server.call.duration"
				item.Definition.Source.Metric = item.Metric
				wantProfile = "generic"
			case 5:
				item.Metric = "http.server.request.duration"
				item.Definition.Source.Metric = item.Metric
				wantProfile = "http-server-duration"
			case 6:
				item.Metric = "messaging.process.duration"
				item.Definition.Source.Metric = item.Metric
				wantProfile = "messaging-process-duration"
			case 7:
				item.Supported = false
			}
			shape := item.Metric == "rpc.server.call.duration" && item.Kind == "histogram" && item.Unit == "s"
			if measurements.CurrentRPCDurationShape(item) != shape {
				tc.Fatal("shared shape fact broadened name/type/unit")
			}
			plan := measurements.PlanInstrument(item, observations)
			if plan.Profile != wantProfile || plan.Eligible != rpcExpectedEligibility362(item, observations) || plan.Verification != "not-run" {
				tc.Fatal("generic binding/legacy/custom planning changed", variant)
			}
			if item.Metric == "rpc.server.call.duration" && plan.ShapeMatches != shape {
				tc.Fatal("candidate shape mismatch")
			}
			if variant <= 4 || variant == 7 {
				f := rpcFixture357{Scope: scope, Service: "api"}
				service, e := rpcPlan357(f, "auto", false, false, []measurements.Instrument{item})
				if e != nil {
					tc.Fatal("valid existing binding rejected", e)
				}
				if (service.Profile == "rpc-server-current") != (shape && item.Supported) {
					tc.Fatal("Services/Measurements literal metadata capability disagreement", variant, service.Profile)
				}
				if variant == 3 && service.Profile != "rpc-server" {
					tc.Fatal("legacy milliseconds silently aliased/converted")
				}
			}
		}
		nilBinding := i
		nilBinding.Definition = nil
		if measurements.PlanInstrument(nilBinding, observations).Eligible {
			tc.Fatal("metadata invented definition")
		}
		emptyKeys := []measurements.DiscoveryObservation{{Resource: map[string]string{"service.name": "api", "service.instance.id": "one"}, Attributes: map[string]string{}}}
		optional := measurements.PlanInstrument(i, emptyKeys)
		if !optional.Eligible || !optional.ShapeMatches || optional.Verification != "not-run" {
			tc.Fatal("missing conditional keys blocked generic Run or verified conformance")
		}
		if err := rpcCoverageCheck362(i, emptyKeys, optional); err != nil {
			tc.Fatal(err)
		}
		bad := slices.Clone(observations)
		if len(bad) > 0 {
			bad[0].Resource = nil
			if err := rpcCoverageCheck362(i, bad, measurements.PlanInstrument(i, bad)); err != nil {
				tc.Fatal(err)
			}
		}
		// A metadata candidate cannot turn a no-value retained point into valid time
		// or rewrite the original generic binding/verification after evaluation.
		row, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": "api", "service.instance.id": "one", "deployment.environment.name": "test"}, "scope": i.Scope, "version": i.Version, "schema_url": i.SchemaURL, "name": i.Metric, "unit": i.Unit, "attributes": map[string]string{}, "temporality": 1, "start_ns": "1000000000", "end_ns": "3000000000", "count": "4", "bounds": []float64{.01, .05}, "bucket_counts": []string{"1", "2", "1"}, "flags": 1})
		result, e := measurements.Evaluate(measurements.Question{Definition: *i.Definition, FromMS: 1000, ToMS: 3000}, []json.RawMessage{row}, time.UnixMilli(4000))
		if e != nil || result.State != "unknown" || savedJSON360(p) != savedJSON360(measurements.PlanInstrument(i, observations)) {
			tc.Fatal("metadata promoted missing evidence/conformance", e)
		}
		encoded, _ := json.Marshal(map[string]any{"instrument": i, "plan": p})
		cmd := exec.CommandContext(t.Context(), "node", "rpc_capability_362_fixture.mjs")
		cmd.Stdin = bytes.NewReader(encoded)
		raw, e := cmd.Output()
		if e != nil {
			if fail, ok := e.(*exec.ExitError); ok {
				tc.Fatal("actual metadata presentation", string(fail.Stderr))
			} else {
				tc.Fatal(e)
			}
		}
		var verdict struct{ Complete bool }
		if json.Unmarshal(raw, &verdict) != nil || !verdict.Complete {
			tc.Fatal("incomplete actual capability display")
		}
		if len(p.Coverage.Streams) > 0 {
			p.Coverage.Streams[0].Stream.Resource["edited"] = "caller"
			p.Coverage.Streams[0].Stream.Attributes["rpc.method"] = "caller"
		}
		if savedJSON360([]any{i, observations}) != before {
			tc.Fatal("coverage output retained mutable input aliases")
		}
	}, options(40)...)
}
