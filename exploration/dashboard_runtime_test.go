package exploration

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/measurements"
)

func dashboardRuntimeInstrument(metric, instance, memoryType string) measurements.Instrument {
	unit := "By"
	if metric == "go.goroutine.count" {
		unit = "{goroutine}"
	}
	i := dashboardTestInstrument("source-runtime", metric, "stock", unit)
	i.Temporality = 2
	i.SchemaURL = measurements.GoRuntimeSchema
	i.Definition.Source.SchemaURL = i.SchemaURL
	attrs := map[string]string{}
	if memoryType != "" {
		attrs["go.memory.type"] = memoryType
	}
	p := measurements.PlanInstrument(i, []measurements.DiscoveryObservation{{Resource: map[string]string{"service.name": "api", "service.instance.id": instance, "deployment.environment.name": "prod"}, Attributes: attrs}})
	i.CapabilityPlan = &p
	return i
}
func TestDashboardRuntimeExactDefaultsDedupAndOptionalGoroutines(t *testing.T) {
	in := dashboardHTTPInput()
	in.Signals = []string{"metrics"}
	in.Profile = "auto"
	memory := dashboardRuntimeInstrument("go.memory.used", "one", "stack")
	goal := dashboardRuntimeInstrument("go.memory.gc.goal", "one", "")
	goroutines := dashboardRuntimeInstrument("go.goroutine.count", "one", "")
	duplicate := memory
	d := *memory.Definition
	d.Name = "duplicate-metadata"
	d.Title = "Another title"
	duplicate.Definition = &d
	in.Instruments = []measurements.Instrument{goroutines, duplicate, goal, memory, dashboardRuntimeInstrument("go.memory.used", "one", "other")}
	plan, e := PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(plan.NativeOptions) != 4 {
		t.Fatal("duplicate metadata multiplied stream", len(plan.NativeOptions))
	}
	selected := 0
	for _, option := range plan.NativeOptions {
		if option.Selected {
			selected++
			if option.Instrument.Metric == "go.goroutine.count" {
				t.Fatal("goroutines default")
			}
			if option.Instrument.Definition.Source.ExactStream == nil || option.Instrument.Definition.StockDomain != "nonnegative-integer" {
				t.Fatal(option)
			}
		}
	}
	if selected != 2 {
		t.Fatal("default pair missing")
	}
	count := 0
	for _, w := range plan.Widgets {
		if w.Measurement != nil {
			count++
			if w.Family != "go-runtime" {
				t.Fatal(w)
			}
		}
	}
	if count != 2 {
		t.Fatal(count)
	}
	reversed := in
	reversed.Instruments = append([]measurements.Instrument{}, in.Instruments...)
	for a, b := 0, len(reversed.Instruments)-1; a < b; a, b = a+1, b-1 {
		reversed.Instruments[a], reversed.Instruments[b] = reversed.Instruments[b], reversed.Instruments[a]
	}
	again, e := PlanServiceDashboard(reversed, nil)
	if e != nil || !reflect.DeepEqual(plan, again) {
		t.Fatal("metadata order changed plan", e)
	}
	in.Instruments = []measurements.Instrument{goroutines}
	plan, e = PlanServiceDashboard(in, nil)
	if e != nil || plan.NativeOptions[0].Default || plan.NativeOptions[0].Selected {
		t.Fatal("optional runtime automaticallyread", plan, e)
	}
	memory.CapabilityPlan = nil
	in.Instruments = []measurements.Instrument{memory}
	plan, e = PlanServiceDashboard(in, nil)
	if e != nil || plan.NativeOptions[0].Default {
		t.Fatal("unconfirmed runtime broad read", plan, e)
	}
}
func TestDashboardRuntimeGlobalOptionBoundAndPreviewOmissions(t *testing.T) {
	in := dashboardHTTPInput()
	in.Profile = ""
	in.Signals = []string{"metrics"}
	in.Instruments = nil
	for n := 0; n < 6; n++ {
		i := dashboardRuntimeInstrument("go.memory.used", "one", "stack")
		i.Definition.Name = fmt.Sprintf("input-%d", n)
		i.CapabilityPlan.Coverage.Streams = nil
		for j := 0; j < 25; j++ {
			single := dashboardRuntimeInstrument("go.memory.used", fmt.Sprintf("instance-%d-%d", n, j), "stack")
			i.CapabilityPlan.Coverage.Streams = append(i.CapabilityPlan.Coverage.Streams, single.CapabilityPlan.Coverage.Streams[0])
		}
		i.CapabilityPlan.Coverage.OmittedStreams = 2
		in.Instruments = append(in.Instruments, i)
	}
	plan, e := PlanServiceDashboard(in, nil)
	if e != nil || len(plan.NativeOptions) != 128 || plan.NativeOptionsOmitted != 22 || plan.RuntimeStreamsOmitted != 12 {
		t.Fatal("unbounded or undisclosed options", plan, e)
	}
	selected := 0
	for _, o := range plan.NativeOptions {
		if o.Selected {
			selected++
		}
	}
	if selected != 1 {
		t.Fatal("multiple memorycomponents default", selected)
	}
}
