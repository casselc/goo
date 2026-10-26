package measurements

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"math"
	"reflect"
	"testing"
)

func runtimeInstrument(t *testing.T, metric string, attributes map[string]string) Instrument {
	d := stockDefinition(t)
	d.Source.Metric = metric
	d.Source.Unit = "By"
	if metric == "go.goroutine.count" {
		d.Source.Unit = "{goroutine}"
	}
	d.Source.SchemaURL = GoRuntimeSchema
	i := Instrument{Kind: "stock", Metric: metric, Unit: d.Source.Unit, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: GoRuntimeSchema, Temporality: 2, Supported: true, Definition: &d}
	row := stockRow(d, 120000, "4")
	observedAttributes := attributes
	if observedAttributes == nil {
		observedAttributes = map[string]string{}
	}
	p := PlanInstrument(i, []DiscoveryObservation{{Resource: row.Resource, Attributes: observedAttributes}})
	if attributes == nil && len(p.Coverage.Streams) > 0 {
		p.Coverage.Streams[0].Stream.Attributes = nil
	}
	i.CapabilityPlan = &p
	return i
}
func TestGoRuntimePinnedCandidatesAndAuthoring(t *testing.T) {
	for _, metric := range []string{"go.memory.used", "go.memory.gc.goal", "go.goroutine.count"} {
		i := runtimeInstrument(t, metric, nil)
		before, _ := json.Marshal(i)
		candidates := GoRuntimeCandidates(i)
		if len(candidates) != 1 || !GoRuntimeExactCandidate(candidates[0]) || i.CapabilityPlan.Profile != "go-runtime" {
			t.Fatal(metric, candidates)
		}
		d := candidates[0].Definition
		if d.Source.ExactStream == nil || d.Source.ExactStream.Resource["service.instance.id"] == "" || d.StockDomain != "nonnegative-integer" {
			t.Fatal(d)
		}
		authored, e := yaml.Marshal(d)
		if e != nil {
			t.Fatal(e)
		}
		reopened, e := Read(string(authored))
		if e != nil || Hash(*d) != Hash(reopened) {
			t.Fatal("generated runtime YAML drift", e)
		}
		candidates[0].Definition.Source.ExactStream.Resource["service.instance.id"] = "changed"
		after, _ := json.Marshal(i)
		if string(before) != string(after) {
			t.Fatal("candidate mutated original metadata")
		}
	}
}
func TestGoRuntimeCandidateRefusalAndLiteralTypes(t *testing.T) {
	for _, edit := range []func(*Instrument){func(i *Instrument) { i.Kind = "gauge" }, func(i *Instrument) { i.Temporality = 1 }, func(i *Instrument) { i.Monotonic = true }, func(i *Instrument) { i.Unit = "1" }, func(i *Instrument) { i.SchemaURL = "https://opentelemetry.io/schemas/1.30.0" }, func(i *Instrument) { i.CapabilityPlan.Coverage.Streams[0].Stream.Resource["service.instance.id"] = "" }, func(i *Instrument) {
		i.CapabilityPlan.Coverage.Streams[0].Stream.Resource["deployment.environment.name"] = "other"
	}, func(i *Instrument) { i.CapabilityPlan.Coverage.Streams[0].Stream.Scope = "other" }} {
		i := runtimeInstrument(t, "go.memory.used", map[string]string{})
		edit(&i)
		if len(GoRuntimeCandidates(i)) != 0 {
			t.Fatal("wrong runtime metadata admitted", i)
		}
	}
	for _, attributes := range []map[string]string{nil, {}, {"go.memory.type": ""}, {"go.memory.type": "stack"}, {"go.memory.type": "custom"}} {
		i := runtimeInstrument(t, "go.memory.used", attributes)
		c := GoRuntimeCandidates(i)
		if len(c) != 1 {
			t.Fatal(attributes)
		}
		if !reflect.DeepEqual(c[0].Definition.Source.ExactStream.Attributes, func() map[string]string {
			if attributes == nil {
				return map[string]string{}
			}
			return attributes
		}()) {
			t.Fatal("literal dimensions altered")
		}
	}
	i := runtimeInstrument(t, "go.memory.used", map[string]string{})
	i.CapabilityPlan.Coverage.OmittedStreams = math.MaxInt
	if _, e := GoRuntimePreviewOmissions(i); e == nil {
		t.Fatal("overflow metadata")
	}
}
func TestGoRuntimeDefaultPairAndGoroutinesExplicit(t *testing.T) {
	memory := GoRuntimeCandidates(runtimeInstrument(t, "go.memory.used", map[string]string{"go.memory.type": "stack"}))[0]
	goal := GoRuntimeCandidates(runtimeInstrument(t, "go.memory.gc.goal", map[string]string{}))[0]
	goroutines := GoRuntimeCandidates(runtimeInstrument(t, "go.goroutine.count", map[string]string{}))[0]
	chosen := GoRuntimeDefaults([]Instrument{goroutines, goal, memory})
	if len(chosen) != 2 || !chosen[Hash(*memory.Definition)] || !chosen[Hash(*goal.Definition)] || chosen[Hash(*goroutines.Definition)] {
		t.Fatal(chosen)
	}
	if len(GoRuntimeDefaults([]Instrument{goroutines})) != 0 {
		t.Fatal("goroutines auto selected")
	}
}
