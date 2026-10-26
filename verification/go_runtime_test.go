package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
)

func grInstrument(metric, instance string) measurements.Instrument {
	unit := "By"
	if metric == "go.goroutine.count" {
		unit = "{goroutine}"
	}
	i := hdInstrument("runtime-"+metric, metric, "stock", unit, exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}, "runtime-api")
	i.Temporality = 2
	i.SchemaURL = measurements.GoRuntimeSchema
	i.Definition.Source.SchemaURL = i.SchemaURL
	s := metrics.Stream{Resource: map[string]string{"deployment.environment.name": "prod", "service.name": "runtime-api", "service.instance.id": instance, "process.pid": "123"}, ResourceSchemaURL: "resource-schema", Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Name: metric, Unit: unit, Attributes: map[string]string{}}
	if metric == "go.memory.used" {
		s.Attributes["go.memory.type"] = "stack"
	}
	i.CapabilityPlan = &measurements.InstrumentCapabilityPlan{Coverage: measurements.InstrumentCoverage{Streams: []measurements.ObservedStream{{Stream: s}}}}
	return i
}
func grCopy(i measurements.Instrument) measurements.Instrument {
	b, _ := json.Marshal(i)
	var out measurements.Instrument
	_ = json.Unmarshal(b, &out)
	return out
}
func TestHegelGoRuntimeExactPopulationAndDefaults(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		instance := fmt.Sprintf("worker-%d", hegel.Draw(tc, hegel.Integers(1, 100000)))
		inputs := []measurements.Instrument{grInstrument("go.memory.used", instance), grInstrument("go.memory.gc.goal", instance), grInstrument("go.goroutine.count", instance), grInstrument("go.memory.used", instance+"-other")}
		inputs[0].CapabilityPlan.Coverage.Streams = append(inputs[0].CapabilityPlan.Coverage.Streams, grCopy(inputs[0]).CapabilityPlan.Coverage.Streams[0])
		extra := grCopy(inputs[0]).CapabilityPlan.Coverage.Streams[0]
		extra.Stream.Attributes["go.memory.type"] = "custom"
		inputs[0].CapabilityPlan.Coverage.Streams = append(inputs[0].CapabilityPlan.Coverage.Streams, extra)
		before, _ := json.Marshal(inputs)
		expanded := []measurements.Instrument{}
		for _, i := range inputs {
			out := measurements.GoRuntimeCandidates(i)
			want := 1
			if i.Metric == "go.memory.used" && i.CapabilityPlan.Coverage.Streams[0].Stream.Resource["service.instance.id"] == instance {
				want = 2
			}
			if len(out) != want {
				tc.Errorf("exact population dedup %d want%d", len(out), want)
				return
			}
			for _, candidate := range out {
				s := candidate.Definition.Source.ExactStream
				if s == nil || candidate.Definition.StockDomain != "nonnegative-integer" {
					tc.Errorf("unpinned/domain")
					return
				}
				witness := false
				for _, p := range i.CapabilityPlan.Coverage.Streams {
					if reflect.DeepEqual(p.Stream, *s) {
						witness = true
					}
				}
				if !witness {
					tc.Errorf("identity fabricated")
					return
				}
				expanded = append(expanded, candidate)
			}
		}
		defaults := measurements.GoRuntimeDefaults(expanded)
		if len(defaults) != 2 {
			tc.Errorf("default count")
			return
		}
		selected := []measurements.Instrument{}
		for _, i := range expanded {
			if defaults[measurements.Hash(*i.Definition)] {
				if i.Metric == "go.goroutine.count" {
					tc.Errorf("goroutine autoread")
					return
				}
				selected = append(selected, i)
			}
		}
		if len(selected) != 2 || selected[0].Metric == selected[1].Metric || !reflect.DeepEqual(selected[0].Definition.Source.ExactStream.Resource, selected[1].Definition.Source.ExactStream.Resource) {
			tc.Errorf("default pair joined process identities")
			return
		}
		for a, b := 0, len(expanded)-1; a < b; a, b = a+1, b-1 {
			expanded[a], expanded[b] = expanded[b], expanded[a]
		}
		if !reflect.DeepEqual(defaults, measurements.GoRuntimeDefaults(expanded)) {
			tc.Errorf("arrival order changed defaults")
			return
		}
		expanded[0].Definition.Source.ExactStream.Resource["service.instance.id"] = "changed"
		after, _ := json.Marshal(inputs)
		if string(before) != string(after) {
			tc.Errorf("definition output aliases input")
			return
		}
	}, options(12)...)
}
func TestGoRuntimeAdmissionAndPreviewBoundaries(t *testing.T) {
	edits := []func(*measurements.Instrument){func(i *measurements.Instrument) { i.Kind = "gauge" }, func(i *measurements.Instrument) { i.Unit = "1" }, func(i *measurements.Instrument) { i.SchemaURL = "https://opentelemetry.io/schemas/1.36.0" }, func(i *measurements.Instrument) { i.Temporality = 1 }, func(i *measurements.Instrument) { i.Monotonic = true }, func(i *measurements.Instrument) {
		i.CapabilityPlan.Coverage.Streams[0].Stream.Resource["service.instance.id"] = ""
	}, func(i *measurements.Instrument) {
		i.CapabilityPlan.Coverage.Streams[0].Stream.Resource["service.name"] = "other"
	}, func(i *measurements.Instrument) { i.CapabilityPlan.Coverage.Streams[0].Stream.Version = "other" }, func(i *measurements.Instrument) {
		i.Definition.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "one"}
	}}
	for n, edit := range edits {
		i := grInstrument("go.memory.used", "one")
		edit(&i)
		if len(measurements.GoRuntimeCandidates(i)) != 0 {
			t.Fatal("ineligible shape/identity", n)
		}
	}
	for _, at := range []int{24, 25} {
		i := grInstrument("go.memory.used", "one")
		match := i.CapabilityPlan.Coverage.Streams[0]
		bad := grCopy(i).CapabilityPlan.Coverage.Streams[0]
		bad.Stream.Resource["service.instance.id"] = ""
		i.CapabilityPlan.Coverage.Streams = make([]measurements.ObservedStream, 26)
		for n := range i.CapabilityPlan.Coverage.Streams {
			i.CapabilityPlan.Coverage.Streams[n] = bad
		}
		i.CapabilityPlan.Coverage.Streams[at] = match
		want := 0
		if at == 24 {
			want = 1
		}
		n, e := measurements.GoRuntimePreviewOmissions(i)
		if e != nil || n != 1 || len(measurements.GoRuntimeCandidates(i)) != want {
			t.Fatal("bounded preview", at, e)
		}
	}
	for _, v := range []int{-1, 10001, int(^uint(0) >> 1)} {
		i := grInstrument("go.memory.used", "one")
		i.CapabilityPlan.Coverage.OmittedStreams = v
		if _, e := measurements.GoRuntimePreviewOmissions(i); e == nil {
			t.Fatal("omission overflow", v)
		}
	}
}

func TestGoRuntimeDashboardExplicitCandidatesAndInputCustody(t *testing.T) {
	inputs := []measurements.Instrument{grInstrument("go.memory.used", "one"), grInstrument("go.memory.gc.goal", "one"), grInstrument("go.goroutine.count", "one")}
	for _, name := range []string{"other", "custom", ""} {
		p := grCopy(inputs[0]).CapabilityPlan.Coverage.Streams[0]
		if name == "" {
			delete(p.Stream.Attributes, "go.memory.type")
		} else {
			p.Stream.Attributes["go.memory.type"] = name
		}
		inputs[0].CapabilityPlan.Coverage.Streams = append(inputs[0].CapabilityPlan.Coverage.Streams, p)
	}
	request := exploration.DashboardRequest{Profile: "auto", Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}, Service: "runtime-api", Signals: []string{"metrics"}, Instruments: inputs}
	before, _ := json.Marshal(request)
	plan, e := exploration.PlanServiceDashboard(request, nil)
	if e != nil {
		t.Fatal(e)
	}
	defaults := 0
	for _, o := range plan.NativeOptions {
		if o.Key != measurements.Hash(*o.Instrument.Definition) || o.Instrument.Definition.Source.ExactStream == nil {
			t.Fatal("executed selection not exact definition")
		}
		if o.Default {
			defaults++
			if o.Instrument.Metric == "go.goroutine.count" {
				t.Fatal("goroutine default")
			}
		}
	}
	if defaults != 2 || len(plan.NativeOptions) != 6 {
		t.Fatal("candidate loss or automatic budget", defaults, len(plan.NativeOptions))
	}
	for _, w := range plan.Widgets {
		if w.Measurement != nil && (w.Measurement.FromMS != request.Scope.FromMS || w.Measurement.ToMS != request.Scope.ToMS) {
			t.Fatal("evaluation scope changed")
		}
	}
	after, _ := json.Marshal(request)
	if string(before) != string(after) {
		t.Fatal("planner mutated supplied discovery")
	}
}
