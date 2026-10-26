package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func runtimeHTTPInstrument(metric string) measurements.Instrument {
	unit := "By"
	if metric == "go.goroutine.count" {
		unit = "{goroutine}"
	}
	d := measurements.Definition{APIVersion: measurements.Version, Name: "runtime-" + strings.ReplaceAll(metric, ".", "-"), Title: metric, Kind: "stock", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Namespace: &identity.Namespace{Mode: "value", Value: "commerce"}, Services: []string{"checkout"}, Scope: "go-sdk", Version: "1.0.0", SchemaURL: measurements.GoRuntimeSchema, Metric: metric, Unit: unit}}
	attrs := map[string]string{}
	if metric == "go.memory.used" {
		attrs["go.memory.type"] = "heap"
	}
	stream := metrics.Stream{Resource: map[string]string{"service.name": "checkout", "service.instance.id": "runtime-1", "service.namespace": "commerce", "deployment.environment.name": "prod"}, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: metric, Unit: unit, Attributes: attrs}
	return measurements.Instrument{Kind: "stock", Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Metric: metric, Unit: unit, Temporality: 2, Supported: true, Definition: &d, CapabilityPlan: &measurements.InstrumentCapabilityPlan{Coverage: measurements.InstrumentCoverage{Streams: []measurements.ObservedStream{{Stream: stream}}}}}
}
func runtimeHTTPBody(t *testing.T, change func(*exploration.DashboardRequest)) string {
	t.Helper()
	var req exploration.DashboardRequest
	if err := json.Unmarshal([]byte(dashboardProfileBody("auto", []string{"metrics"}, []exploration.Field{})), &req); err != nil {
		t.Fatal(err)
	}
	for _, metric := range []string{"go.memory.used", "go.memory.gc.goal", "go.goroutine.count"} {
		req.Instruments = append(req.Instruments, runtimeHTTPInstrument(metric))
	}
	if change != nil {
		change(&req)
	}
	raw, _ := json.Marshal(req)
	return string(raw)
}
func runtimeHTTPPlan(t *testing.T, body string) (*exploration.DashboardPlan, *httptest.ResponseRecorder) {
	t.Helper()
	mux := http.NewServeMux()
	registerDashboards(mux, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/dashboards/plan", strings.NewReader(body)))
	var plan exploration.DashboardPlan
	if w.Code == 200 {
		if e := json.Unmarshal(w.Body.Bytes(), &plan); e != nil {
			t.Fatal(e)
		}
	}
	return &plan, w
}
func TestRuntimeDashboardHTTPFrozenContracts(t *testing.T) {
	cases := []planningWireCase{
		{Name: "runtime-exact-memory-goal-defaults", Body: runtimeHTTPBody(t, nil), Status: 200},
		{Name: "runtime-missing-instance-generic-explicit-only", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
			for i := range r.Instruments {
				delete(r.Instruments[i].CapabilityPlan.Coverage.Streams[0].Stream.Resource, "service.instance.id")
			}
		}), Status: 200},
		{Name: "runtime-empty-selection", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) { selection := []string{}; r.NativeSelection = &selection }), Status: 200},
		{Name: "runtime-explicit-goroutines", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
			candidates := measurements.GoRuntimeCandidates(r.Instruments[2])
			if len(candidates) != 1 {
				t.Fatal("goroutine fixture")
			}
			selection := []string{measurements.Hash(*candidates[0].Definition)}
			r.NativeSelection = &selection
		}), Status: 200},
		{Name: "runtime-preview-omissions", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) { r.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = 3 }), Status: 200},
		{Name: "runtime-invalid-preview-omissions", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) { r.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = -1 }), Status: 400},
		{Name: "runtime-wrong-definition-scope", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
			r.Instruments[0].Definition.Source.Services = []string{"other-service"}
		}), Status: 400},
		{Name: "runtime-delta-nonmonotonic-refused", Body: runtimeHTTPBody(t, func(r *exploration.DashboardRequest) { r.Instruments[0].Temporality = 1 }), Status: 400},
	}
	for i := range cases {
		cases[i].Method = "POST"
		cases[i].Path = "/dashboards/plan"
	}
	const filename = "../../testdata/http/runtime-dashboard.json"
	if os.Getenv("GOO_UPDATE_RUNTIME_DASHBOARD_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := planningWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		raw, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e := os.WriteFile(filename, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	fixture := dashboardProfileFixture(t, filename, len(cases))
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := planningWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("runtime header changed", k)
				}
			}
		})
	}
}
func TestRuntimeDashboardHTTPExactDefaultsAndStableSelection(t *testing.T) {
	body := runtimeHTTPBody(t, nil)
	plan, w := runtimeHTTPPlan(t, body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	chosen := map[string]bool{}
	resource := ""
	for _, widget := range plan.Widgets {
		if widget.Measurement == nil {
			continue
		}
		q := widget.Measurement
		d := q.Definition
		stream := d.Source.ExactStream
		if widget.Family != "go-runtime" || d.StockDomain != "nonnegative-integer" || stream == nil || stream.Resource["service.instance.id"] == "" {
			t.Fatal("runtime not pinned", widget)
		}
		encoded, _ := json.Marshal(stream.Resource)
		if resource == "" {
			resource = string(encoded)
		} else if resource != string(encoded) {
			t.Fatal("default process identities differ")
		}
		if q.FromMS != 1000 || q.ToMS != 61000 || d.Environment != "prod" || d.Source.Namespace == nil || d.Source.Namespace.Value != "commerce" {
			t.Fatal("runtime question changed scope", q)
		}
		chosen[d.Source.Metric] = true
	}
	if len(chosen) != 2 || !chosen["go.memory.used"] || !chosen["go.memory.gc.goal"] || chosen["go.goroutine.count"] {
		t.Fatal("runtime automatic set", chosen)
	}
	reordered := runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
		for a, b := 0, len(r.Instruments)-1; a < b; a, b = a+1, b-1 {
			r.Instruments[a], r.Instruments[b] = r.Instruments[b], r.Instruments[a]
		}
	})
	_, other := runtimeHTTPPlan(t, reordered)
	if other.Code != 200 || other.Body.String() != w.Body.String() {
		t.Fatal("runtime input arrival order changed plan")
	}
	missing, _ := runtimeHTTPPlan(t, runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
		for i := range r.Instruments {
			delete(r.Instruments[i].CapabilityPlan.Coverage.Streams[0].Stream.Resource, "service.instance.id")
		}
	}))
	for _, widget := range missing.Widgets {
		if widget.Measurement != nil {
			t.Fatal("unconfirmed runtime broadly auto-executed", widget)
		}
	}
	for _, option := range missing.NativeOptions {
		if option.Selected || option.Default {
			t.Fatal("generic runtime fallback selected")
		}
	}
}
func TestRuntimeDashboardHTTPOptionAndPreviewBounds(t *testing.T) {
	body := runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
		r.Instruments = nil
		for index := 0; index < 6; index++ {
			i := runtimeHTTPInstrument("go.memory.used")
			i.Definition.Name = fmt.Sprintf("runtime-%d", index)
			i.CapabilityPlan.Coverage.Streams = nil
			for n := 0; n < 25; n++ {
				stream := runtimeHTTPInstrument("go.memory.used").CapabilityPlan.Coverage.Streams[0].Stream
				stream.Resource["service.instance.id"] = fmt.Sprintf("runtime-%d-%d", index, n)
				i.CapabilityPlan.Coverage.Streams = append(i.CapabilityPlan.Coverage.Streams, measurements.ObservedStream{Stream: stream})
			}
			r.Instruments = append(r.Instruments, i)
		}
	})
	if len(body) > 256<<10 {
		t.Fatal("option bound fixture exceeds Go transport")
	}
	plan, w := runtimeHTTPPlan(t, body)
	if w.Code != 200 || len(plan.NativeOptions) != 128 || plan.NativeOptionsOmitted != 22 {
		t.Fatal("native option cap failed", w.Code, len(plan.NativeOptions), plan.NativeOptionsOmitted)
	}
}

func TestRuntimeDashboardHTTPPreviewClipDisclosesUnknownStreams(t *testing.T) {
	body := runtimeHTTPBody(t, func(r *exploration.DashboardRequest) {
		r.Instruments = r.Instruments[:1]
		i := &r.Instruments[0]
		for n := 1; n < 26; n++ {
			stream := runtimeHTTPInstrument("go.memory.used").CapabilityPlan.Coverage.Streams[0].Stream
			stream.Resource["service.instance.id"] = fmt.Sprintf("runtime-%d", n+1)
			i.CapabilityPlan.Coverage.Streams = append(i.CapabilityPlan.Coverage.Streams, measurements.ObservedStream{Stream: stream})
		}
		i.CapabilityPlan.Coverage.OmittedStreams = 2
	})
	plan, w := runtimeHTTPPlan(t, body)
	if w.Code != 200 || len(plan.NativeOptions) != 25 || plan.RuntimeStreamsOmitted != 3 {
		t.Fatal("runtime clipped/omitted streams undisclosed", w.Code, len(plan.NativeOptions), plan.RuntimeStreamsOmitted)
	}
}
