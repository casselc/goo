package measurements

import (
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestCurrentRPCDiscoveryCapability(t *testing.T) {
	q := DiscoveryQuestion{Environment: "test", Service: "api", FromMS: 1000, ToMS: 2000}
	for _, unit := range []string{"s", "ms"} {
		for _, temporality := range []int{1, 2} {
			rows := []json.RawMessage{}
			for _, attrs := range []map[string]string{
				{"rpc.system.name": "grpc", "rpc.method": "example.Service/Call", "rpc.status_code": "OK"},
				{"rpc.system.name": "custom", "rpc.method": "", "error.type": "custom.failure"},
				{"rpc.system.name": "grpc", "rpc.method": "_OTHER"},
				{},
			} {
				b, err := json.Marshal(map[string]any{"kind": "histogram", "scope": "rpc", "version": "test", "schemaURL": "custom-schema", "metric": CurrentRPCDurationMetric, "unit": unit, "temporality": temporality, "resource": map[string]string{"service.name": "api"}, "attributes": attrs})
				if err != nil {
					t.Fatal(err)
				}
				rows = append(rows, b)
			}
			result, err := discoverRows(q, rows)
			if err != nil || len(result.Instruments) != 1 {
				t.Fatal(result, err)
			}
			i := result.Instruments[0]
			p := i.CapabilityPlan
			if p.Profile != "rpc-server-call-duration" || p.Title != "Current RPC duration candidate" || p.RuleSet != CurrentRPCRuleSet || p.ShapeMatches != (unit == "s") || !p.Eligible || p.Verification != "not-run" {
				t.Fatalf("%+v", p)
			}
			if i.Definition.Source.Unit != unit || i.Definition.Source.SchemaURL != "custom-schema" || i.Definition.Source.ExactStream != nil {
				t.Fatal("metadata changed binding", i.Definition)
			}
			if p.Coverage.TotalStreams != 4 || len(p.Coverage.Dimensions) != 4 {
				t.Fatal(p.Coverage)
			}
			for _, d := range p.Coverage.Dimensions {
				if d.Present+d.Empty+d.Missing != 4 {
					t.Fatal(d)
				}
				if d.Key == "rpc.method" && (d.Present != 2 || d.Empty != 1 || d.Missing != 1) {
					t.Fatal(d)
				}
			}
			if !slices.ContainsFunc(p.Hints, func(h string) bool {
				return strings.Contains(h, "conditional") && strings.Contains(h, "missing method does not disable Run")
			}) {
				t.Fatal(p.Hints)
			}
			// Observations describe a plan; they cannot modify the executable source.
			without := []json.RawMessage{json.RawMessage(`{"kind":"histogram","scope":"rpc","version":"test","schemaURL":"custom-schema","metric":"rpc.server.call.duration","unit":"` + unit + `","temporality":` + strconv.Itoa(temporality) + `}`)}
			old, err := discoverRows(q, without)
			if err != nil || !reflect.DeepEqual(i.Definition, old.Instruments[0].Definition) || old.Instruments[0].CapabilityPlan.Coverage.State != "unavailable" {
				t.Fatal("preview changed binding", err)
			}
		}
	}
}

func TestCurrentRPCShapeDoesNotGrantEligibility(t *testing.T) {
	base := Instrument{Metric: CurrentRPCDurationMetric, Kind: "histogram", Unit: "s", Scope: "rpc", Temporality: 1}
	if p := PlanInstrument(base, nil); !p.ShapeMatches || p.Eligible || p.Verification != "not-run" {
		t.Fatal(p)
	}
	for _, change := range []func(*Instrument){func(i *Instrument) { i.Metric = "rpc.server.duration" }, func(i *Instrument) { i.Metric = "rpc.client.call.duration" }, func(i *Instrument) { i.Metric = "custom.rpc.server.call.duration" }, func(i *Instrument) { i.Kind = "gauge" }, func(i *Instrument) { i.Unit = "ms" }} {
		i := base
		change(&i)
		if CurrentRPCDurationShape(i) {
			t.Fatal("wrong shape recognized", i)
		}
		p := PlanInstrument(i, nil)
		if p.ShapeMatches || i.Metric != CurrentRPCDurationMetric && p.Profile != "generic" {
			t.Fatal(p)
		}
	}
}
