package exploration

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func dashboardCurrentRPCInput357() DashboardRequest {
	in := dashboardHTTPInput()
	in.HTTPServerObserved = false
	in.RPCServerCurrentObserved = true
	in.Fields = []Field{{Kind: "attribute", Key: "rpc.system.name"}}
	return in
}

func TestDashboardCurrentRPCPopulation357(t *testing.T) {
	selectors := []*identity.EnvironmentSelector{nil,
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
	for _, selection := range selectors {
		for _, namespace := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "team'雪"}} {
			in := dashboardCurrentRPCInput357()
			in.Service = "api' OR 1=1 -- 雪"
			in.Scope.Namespace = namespace
			in.Scope.EnvironmentSelector = selection
			if selection != nil {
				in.Scope.Environment = ""
			}
			p, err := PlanServiceDashboard(in, nil)
			if err != nil || p.Profile != "rpc-server-current" {
				t.Fatal(p, err)
			}
			if !reflect.DeepEqual(p.Scope, in.Scope) {
				t.Fatal("scope changed", p.Scope, in.Scope)
			}
			count := 0
			for _, w := range p.Widgets {
				if w.Query == nil || w.Query.Signal != "spans" {
					continue
				}
				count++
				q := w.Query
				if w.Family != "rpc-server-current" || w.Grain != "server-span-record" || q.Summary != "span-duration" || !q.SpanStatus || !reflect.DeepEqual(q.EnvironmentSelector, selection) || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS {
					t.Fatal(w)
				}
				plan, err := Compile(*q)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(plan.Rows, in.Service) || namespace != nil && namespace.Value != "" && strings.Contains(plan.Rows, namespace.Value) {
					t.Fatal("literal interpolated", plan.Rows)
				}
				for _, kind := range []string{"Server", "Client", "Internal", "SPAN_KIND_SERVER", ""} {
					for _, system := range []string{"grpc", "", " "} {
						for _, method := range []string{"example.Service/Get", "_OTHER", "", " "} {
							values := map[Field]string{{Kind: "column", Key: "service"}: in.Service, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "rpc.system.name"}: system, {Kind: "attribute", Key: "rpc.method"}: method}
							if namespace != nil && namespace.Mode == "value" {
								values[Field{Kind: "resource", Key: "service.namespace"}] = namespace.Value
							}
							want := kind == "Server" && system != ""
							if got := Match(q.Where, values); got != want {
								t.Fatalf("kind=%q system=%q method=%q got=%v want=%v", kind, system, method, got, want)
							}
							delete(values, Field{Kind: "attribute", Key: "rpc.method"})
							if got := Match(q.Where, values); got != want {
								t.Fatal("conditional absent method excluded", kind, system, got, want)
							}
							delete(values, Field{Kind: "attribute", Key: "rpc.system.name"})
							values[Field{Kind: "attribute", Key: "rpc.system"}] = "grpc"
							if Match(q.Where, values) {
								t.Fatal("legacy system substituted for current")
							}
							values[Field{Kind: "attribute", Key: "rpc.system.name"}] = system
							values[Field{Kind: "column", Key: "service"}] = "other"
							if Match(q.Where, values) {
								t.Fatal("service selector lost")
							}
						}
					}
				}
				if w.Key == "dimension" && !reflect.DeepEqual(q.GroupBy, []Field{{Kind: "attribute", Key: "rpc.method"}}) {
					t.Fatal("method grouping changed", w)
				}
			}
			if count != 2 {
				t.Fatal("current method dimension must exist even without method metadata", count)
			}
			family := slices.IndexFunc(p.Families, func(f DashboardFamily) bool { return f.Key == "rpc-server-current" })
			if family < 0 || !p.Families[family].Selected || p.Families[family].RuleSet != "goo/rpc-server-current-span-observations/v1" {
				t.Fatal(p.Families)
			}
		}
	}
}

func TestDashboardCurrentRPCSelection357(t *testing.T) {
	in := dashboardCurrentRPCInput357()
	in.RPCServerCurrentObserved = false
	p, err := PlanServiceDashboard(in, nil)
	if err != nil || p.Profile != "all-spans" {
		t.Fatal(p, err)
	}
	if !slices.ContainsFunc(p.Families, func(f DashboardFamily) bool {
		return f.Key == "rpc-server-current" && f.State == "candidate" && !f.Selected
	}) {
		t.Fatal(p.Families)
	}
	in.Profile = "rpc-server-current"
	if _, err := PlanServiceDashboard(in, nil); err != nil {
		t.Fatal("explicit metadata candidate refused", err)
	}
	for _, fields := range [][]Field{nil, {{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}, {{Kind: "resource", Key: "rpc.system.name"}}, {{Kind: "attribute", Key: "rpc.method"}}} {
		in.Fields = fields
		in.RPCServerCurrentObserved = true
		if _, err := PlanServiceDashboard(in, nil); err == nil {
			t.Fatal("explicit current profile fabricated from unsupported metadata", fields)
		}
		in.Profile = "auto"
		p, err := PlanServiceDashboard(in, nil)
		if err != nil || p.Profile != "all-spans" {
			t.Fatal(p, err)
		}
		in.Profile = "rpc-server-current"
	}
	in = dashboardCurrentRPCInput357()
	in.Fields = append(in.Fields, Field{Kind: "attribute", Key: "rpc.system"}, Field{Kind: "attribute", Key: "rpc.method"}, Field{Kind: "attribute", Key: "http.request.method"}, Field{Kind: "attribute", Key: "db.system.name"}, Field{Kind: "attribute", Key: "messaging.system"})
	in.HTTPServerObserved, in.RPCServerObserved, in.DBClientObserved, in.MessagingConsumerObserved, in.MessagingProducerObserved = true, true, true, true, true
	for _, want := range []string{"http-server", "rpc-server-current", "rpc-server", "db-client", "messaging-consumer", "messaging-producer", "all-spans"} {
		p, err := PlanServiceDashboard(in, nil)
		if err != nil || p.Profile != want {
			t.Fatal("automatic precedence", want, p, err)
		}
		switch want {
		case "http-server":
			in.HTTPServerObserved = false
		case "rpc-server-current":
			in.RPCServerCurrentObserved = false
		case "rpc-server":
			in.RPCServerObserved = false
		case "db-client":
			in.DBClientObserved = false
		case "messaging-consumer":
			in.MessagingConsumerObserved = false
		case "messaging-producer":
			in.MessagingProducerObserved = false
		}
	}
	// Deliberately reported dual-key rows are eligible under both distinct lenses.
	values := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: "Server", {Kind: "attribute", Key: "rpc.system.name"}: "grpc", {Kind: "attribute", Key: "rpc.system"}: "grpc", {Kind: "attribute", Key: "rpc.method"}: "_OTHER"}
	base := DashboardRecordQuestion(in.Scope, in.Service, "spans")
	if !Match(dashboardRPCQuestion(base).Where, values) || !Match(dashboardRPCCurrentQuestion(base).Where, values) {
		t.Fatal("dual-key record lost without alias normalization")
	}
	delete(values, Field{Kind: "attribute", Key: "rpc.method"})
	if Match(dashboardRPCQuestion(base).Where, values) || !Match(dashboardRPCCurrentQuestion(base).Where, values) {
		t.Fatal("conditional method distinction lost")
	}
	in.Profile = "all-spans"
	p, err = PlanServiceDashboard(in, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !Match(p.Widgets[0].Query.Where, map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: "Client"}) {
		t.Fatal("manual all-spans fallback constrained")
	}
}

func TestDashboardCurrentRPCNative357(t *testing.T) {
	for _, temporality := range []int{1, 2} {
		in := dashboardCurrentRPCInput357()
		in.Fields = nil
		in.RPCServerCurrentObserved = false
		in.Signals = []string{"metrics"}
		current := dashboardTestInstrument("z-current-rpc", "rpc.server.call.duration", "histogram", "s")
		current.Temporality = temporality
		legacy := dashboardTestInstrument("a-legacy-rpc", "rpc.server.duration", "histogram", "ms")
		in.Instruments = []measurements.Instrument{legacy, dashboardTestInstrument("b-load", "load", "gauge", "1"), current}
		p, err := PlanServiceDashboard(in, nil)
		if err != nil || p.Profile != "rpc-server-current" || p.NativeOptions[0].Instrument.Metric != current.Metric {
			t.Fatal(p, err)
		}
		if !slices.ContainsFunc(p.Widgets, func(w DashboardWidget) bool {
			return w.Measurement != nil && w.Family == "rpc-server-current" && w.Measurement.Definition.Source.Unit == "s"
		}) {
			t.Fatal("current native binding not qualified", p.Widgets)
		}
		if !slices.ContainsFunc(p.Widgets, func(w DashboardWidget) bool {
			return w.Measurement != nil && w.Family == "rpc-server" && w.Measurement.Definition.Source.Unit == "ms"
		}) {
			t.Fatal("separate legacy stream was converted or relabeled", p.Widgets)
		}
		for _, invalid := range []measurements.Instrument{legacy, dashboardTestInstrument("wrong-unit", current.Metric, "histogram", "ms"), dashboardTestInstrument("client", "rpc.client.call.duration", "histogram", "s"), dashboardTestInstrument("gauge", current.Metric, "gauge", "s")} {
			if dashboardNativeRPCCurrent(in.Scope, in.Service, invalid) {
				t.Fatal("unsupported duration binding admitted", invalid)
			}
		}
		clone := func() measurements.Instrument {
			raw, _ := json.Marshal(current)
			var out measurements.Instrument
			if err := json.Unmarshal(raw, &out); err != nil {
				t.Fatal(err)
			}
			return out
		}
		for _, corrupt := range []func(*measurements.Instrument){func(i *measurements.Instrument) { i.Supported = false }, func(i *measurements.Instrument) { i.Temporality = 0 }, func(i *measurements.Instrument) { i.Temporality = 3 }, func(i *measurements.Instrument) { i.Definition.Source.Metric = "other" }, func(i *measurements.Instrument) { i.Definition.Environment = "other" }, func(i *measurements.Instrument) { i.Definition.Source.Services = []string{"other"} }, func(i *measurements.Instrument) {
			i.Definition.Source.Namespace = &identity.Namespace{Mode: "value", Value: "other"}
		}, func(i *measurements.Instrument) { i.Definition.Source.SchemaURL = "other" }} {
			i := clone()
			corrupt(&i)
			if dashboardNativeRPCCurrent(in.Scope, in.Service, i) {
				t.Fatal("corrupt identity or temporality admitted", i)
			}
		}
	}
}
