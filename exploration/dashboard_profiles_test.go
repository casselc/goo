package exploration

import (
	"encoding/json"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"reflect"
	"strings"
	"testing"
)

func dashboardHTTPInput() DashboardRequest {
	return DashboardRequest{HTTPServerObserved: true, Profile: "auto", Scope: InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans", "logs"}, Fields: []Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "db.system.name"}}}
}
func dashboardTestInstrument(name, metric, kind, unit string) measurements.Instrument {
	d := measurements.Definition{APIVersion: measurements.Version, Name: name, Title: name, Kind: kind, Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "ordinary-sdk", Metric: metric, Unit: unit}}
	tempo := 0
	if kind != "gauge" {
		tempo = 2
	}
	return measurements.Instrument{Kind: kind, Metric: metric, Unit: unit, Scope: d.Source.Scope, Supported: true, Temporality: tempo, Definition: &d}
}
func TestDashboardProfileHTTPExactPopulationAndFallback(t *testing.T) {
	for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "payments"}} {
		in := dashboardHTTPInput()
		in.Scope.Namespace = ns
		p, e := PlanServiceDashboard(in, nil)
		if e != nil || p.Profile != "http-server" || p.Families[0].State != "candidate" {
			t.Fatal(p, e)
		}
		for _, widget := range p.Widgets {
			if widget.Query == nil || widget.Query.Signal != "spans" {
				continue
			}
			if widget.Family != "http-server" || widget.Grain != "server-span-record" {
				t.Fatal(widget)
			}
			if _, e := Compile(*widget.Query); e != nil {
				t.Fatal(e)
			}
			for _, kind := range []string{"Server", "Client", "Internal", "SPAN_KIND_SERVER", ""} {
				for _, method := range []string{"GET", "", " "} {
					values := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "http.request.method"}: method}
					if ns != nil && ns.Mode == "value" {
						values[Field{Kind: "resource", Key: "service.namespace"}] = ns.Value
					}
					expected := kind == "Server" && method != ""
					if Match(widget.Query.Where, values) != expected {
						t.Fatalf("population kind=%q method=%q namespace=%v", kind, method, ns)
					}
					delete(values, Field{Kind: "attribute", Key: "http.request.method"})
					if Match(widget.Query.Where, values) {
						t.Fatal("missing method matched")
					}
					values[Field{Kind: "attribute", Key: "http.request.method"}] = "GET"
					values[Field{Kind: "column", Key: "service"}] = "other"
					if Match(widget.Query.Where, values) {
						t.Fatal("service scope lost")
					}
				}
			}
			if widget.Key == "dimension" && widget.Query.GroupBy[0] != (Field{Kind: "attribute", Key: "http.route"}) {
				t.Fatal("HTTP route bucket unavailable")
			}
		}
		in.Profile = "all-spans"
		generic, e := PlanServiceDashboard(in, nil)
		if e != nil || generic.Profile != "all-spans" {
			t.Fatal(e)
		}
		values := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: "Client"}
		if ns != nil && ns.Mode == "value" {
			values[Field{Kind: "resource", Key: "service.namespace"}] = ns.Value
		}
		if !Match(generic.Widgets[0].Query.Where, values) {
			t.Fatal("generic role fallback filtered")
		}
	}
}
func TestDashboardProfileCandidatesAreMetadataNotServerProof(t *testing.T) {
	in := dashboardHTTPInput()
	in.HTTPServerObserved = false
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "all-spans" || p.Families[0].State != "candidate" {
		t.Fatal(p, e)
	}
	in.Profile = "http-server"
	p, e = PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	client := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: "Client", {Kind: "attribute", Key: "http.request.method"}: "GET"}
	if Match(p.Widgets[0].Query.Where, client) {
		t.Fatal("client-only metadata became server evidence")
	}
	in.Profile = "auto"
	for _, field := range []Field{{Kind: "attribute", Key: "http.method"}, {Kind: "resource", Key: "http.request.method"}, {Kind: "attribute", Key: "http.route"}} {
		in.Fields = []Field{field}
		p, e := PlanServiceDashboard(in, nil)
		if e != nil || p.Profile != "all-spans" || p.Families[0].State != "not_observed" {
			t.Fatal("unsupported current profile guessed", field, p, e)
		}
	}
	in.Profile = "http-server"
	if _, e := PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("explicit unobserved profile accepted")
	}
	in.Profile = "rpc-server"
	if _, e := PlanServiceDashboard(in, nil); e == nil {
		t.Fatal("future profile accepted")
	}
	in = dashboardHTTPInput()
	in.Profile = ""
	legacy, e := PlanServiceDashboard(in, nil)
	if e != nil || legacy.Profile != "" || legacy.Families != nil {
		t.Fatal("legacy new metadata emitted")
	}
	raw, _ := json.Marshal(legacy)
	if strings.Contains(string(raw), `"family"`) || strings.Contains(string(raw), `"profile"`) {
		t.Fatal("legacy optional bytes added")
	}
}
func TestDashboardProfileNativePurposeRankAndStableIdentity(t *testing.T) {
	in := dashboardHTTPInput()
	in.Signals = append(in.Signals, "metrics")
	a := dashboardTestInstrument("a", "other-a", "gauge", "1")
	b := dashboardTestInstrument("b", "other-b", "gauge", "1")
	http := dashboardTestInstrument("z-http", "http.server.request.duration", "histogram", "s")
	in.Instruments = []measurements.Instrument{a, b, http}
	p, e := PlanServiceDashboard(in, nil)
	if e != nil || p.NativeOptions[0].Instrument.Metric != "http.server.request.duration" || !p.NativeOptions[0].Default || p.NativeLimit != 2 {
		t.Fatal(p, e)
	}
	selected := 0
	for _, w := range p.Widgets {
		if w.Measurement != nil {
			selected++
		}
	}
	if selected != 2 {
		t.Fatal("native budget changed")
	}
	keys := []string{}
	for _, o := range p.NativeOptions {
		keys = append(keys, o.Key)
	}
	in.Instruments[0].Records = 999
	in.Instruments[0].CapabilityPlan = &measurements.InstrumentCapabilityPlan{Profile: "spoof-http", Eligible: true}
	in.Instruments[0], in.Instruments[2] = in.Instruments[2], in.Instruments[0]
	again, e := PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	againKeys := []string{}
	for _, o := range again.NativeOptions {
		againKeys = append(againKeys, o.Key)
	}
	if !reflect.DeepEqual(keys, againKeys) {
		t.Fatal("counts/coverage/order changed default definition identity")
	}
	empty := []string{}
	in.NativeSelection = &empty
	p, e = PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, w := range p.Widgets {
		if w.Measurement != nil {
			t.Fatal("explicit no native selection ignored")
		}
	}
	in.Fields = nil
	in.NativeSelection = nil
	p, e = PlanServiceDashboard(in, nil)
	if e != nil || p.Profile != "http-server" {
		t.Fatal("eligible native cannot suggest family", p, e)
	}
	for _, schema := range []string{"", "https://opentelemetry.io/schemas/1.30.0", "https://opentelemetry.io/schemas/1.37.0"} {
		candidate := http
		d := *http.Definition
		d.Source.SchemaURL = schema
		candidate.SchemaURL = schema
		candidate.Definition = &d
		in.Instruments = []measurements.Instrument{candidate}
		p, e = PlanServiceDashboard(in, nil)
		if e != nil || p.Profile != "http-server" || p.NativeOptions[0].Instrument.SchemaURL != schema {
			t.Fatal("schema metadata converted/refused", schema, p, e)
		}
	}
	bad := http
	bad.Unit = "By"
	d := *http.Definition
	d.Source.Unit = "By"
	bad.Definition = &d
	if dashboardNativeHTTP(in.Scope, in.Service, bad) {
		t.Fatal("non duration native activated HTTP")
	}
	bad = http
	d = *http.Definition
	d.Environment = "other"
	bad.Definition = &d
	if dashboardNativeHTTP(in.Scope, in.Service, bad) {
		t.Fatal("cross environment candidate")
	}
}
