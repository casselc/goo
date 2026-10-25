package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// Independent, deliberately small interpreter for the service-lens predicates.
// Unsupported operators fail rather than borrowing the production evaluator.
func hdPredicate(p *exploration.Predicate, values map[exploration.Field]string) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for _, c := range p.All {
			if !hdPredicate(&c, values) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for _, c := range p.Any {
			if hdPredicate(&c, values) {
				return true
			}
		}
		return false
	}
	if p.Field == nil {
		return false
	}
	v, present := values[*p.Field]
	switch p.Op {
	case "exists":
		return present
	case "missing":
		return !present
	case "eq":
		return present && v == p.Value
	case "ne":
		return present && v != p.Value
	}
	return false
}
func TestHegelHTTPDashboardPopulationAndIsolation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		from := hegel.Draw(tc, hegel.Integers(1, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7200000))
		service := fmt.Sprintf("api'\\n-%d", seed)
		namespace := fmt.Sprintf("payments-π-%d", seed)
		for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: namespace}} {
			in := exploration.DashboardRequest{Profile: "auto", HTTPServerObserved: true, Scope: exploration.InventoryQuestion{Namespace: ns, Environment: "test", FromMS: int64(from), ToMS: int64(from + width)}, Service: service, Signals: []string{"spans", "logs"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "rpc.method"}, {Kind: "attribute", Key: "db.system.name"}}}
			before, _ := json.Marshal(in)
			plan, e := exploration.PlanServiceDashboard(in, nil)
			after, _ := json.Marshal(in)
			if e != nil || string(before) != string(after) || plan.Profile != "http-server" {
				tc.Errorf("HTTP candidate plan/input %v", e)
				return
			}
			for _, widget := range plan.Widgets {
				if widget.Query == nil {
					continue
				}
				q := widget.Query
				if q.Environment != in.Scope.Environment || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS || q.Limit != 100 || q.Offset != 0 {
					tc.Errorf("scope drift")
					return
				}
				if q.Signal == "spans" && widget.Grain != "server-span-record" || q.Signal == "logs" && widget.Grain != "log-record" {
					tc.Errorf("mixed grain")
					return
				}
				// All combinations include non-HTTP roles, exact nonempty custom method
				// strings, missing/empty maps, legacy-only keys and namespace collisions.
				for _, candidateService := range []string{service, "other"} {
					for _, reportedNS := range []*string{nil, lpcPointer(""), &namespace, lpcPointer("other")} {
						for _, kind := range []string{"Server", "Client", "Internal", "server", ""} {
							for _, method := range []*string{nil, lpcPointer(""), lpcPointer("GET"), lpcPointer(" ")} {
								values := map[exploration.Field]string{{Kind: "column", Key: "service"}: candidateService, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "http.method"}: "GET"}
								if reportedNS != nil {
									values[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *reportedNS
								}
								if method != nil {
									values[exploration.Field{Kind: "attribute", Key: "http.request.method"}] = *method
								}
								scopeOK := candidateService == service
								if ns != nil {
									if ns.Mode == "missing" {
										scopeOK = scopeOK && reportedNS == nil
									} else {
										scopeOK = scopeOK && reportedNS != nil && *reportedNS == ns.Value
									}
								}
								want := scopeOK
								if q.Signal == "spans" {
									want = want && kind == "Server" && method != nil && *method != ""
								}
								if hdPredicate(q.Where, values) != want {
									tc.Errorf("predicate/grain escape kind=%s method=%v namespace=%v", kind, method, reportedNS)
									return
								}
							}
						}
					}
				}
			}
			in.Fields[0], in.Fields[2] = in.Fields[2], in.Fields[0]
			in.Signals[0], in.Signals[1] = in.Signals[1], in.Signals[0]
			again, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || !reflect.DeepEqual(plan, again) {
				tc.Errorf("metadata permutation changes intent")
				return
			}
		}
	}, options(24)...)
}
func hdInstrument(name, metric, kind, unit string, scope exploration.InventoryQuestion, service string) measurements.Instrument {
	d := measurements.Definition{APIVersion: measurements.Version, Name: name, Title: name, Kind: kind, Environment: scope.Environment, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Namespace: scope.Namespace.Clone(), Services: []string{service}, Scope: "ordinary-otel", Version: "custom-version", SchemaURL: "https://example.test/unverified-schema", Metric: metric, Unit: unit}}
	temporality := 0
	if kind != "gauge" {
		temporality = 1
	}
	return measurements.Instrument{Kind: kind, Metric: metric, Unit: unit, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Temporality: temporality, Supported: true, Definition: &d}
}
func TestHTTPDashboardNativePriorityAndCandidateLimits(t *testing.T) {
	scope := exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000, Namespace: &identity.Namespace{Mode: "value", Value: "payments"}}
	in := exploration.DashboardRequest{Profile: "auto", Scope: scope, Service: "api", Signals: []string{"metrics"}}
	gauge := hdInstrument("aaa-unrelated", "custom.pool.capacity", "gauge", "1", scope, in.Service)
	seconds := hdInstrument("zzz-http-seconds", "http.server.request.duration", "histogram", "s", scope, in.Service)
	millis := hdInstrument("yyy-http-millis", "http.server.request.duration", "histogram", "ms", scope, in.Service)
	// Caller-supplied capability hints are not authority for dashboard priority.
	gauge.CapabilityPlan = &measurements.InstrumentCapabilityPlan{Profile: "http-server-duration", Eligible: true, ShapeMatches: true}
	in.Instruments = []measurements.Instrument{gauge, seconds, millis}
	plan, e := exploration.PlanServiceDashboard(in, nil)
	if e != nil || plan.Profile != "http-server" {
		t.Fatal(plan, e)
	}
	selected := []string{}
	for _, option := range plan.NativeOptions {
		if option.Selected {
			selected = append(selected, option.Instrument.Metric)
			if option.Instrument.Metric != "http.server.request.duration" {
				t.Fatal("unrelated metric displaced HTTP")
			}
		}
	}
	if len(selected) != 2 {
		t.Fatal("wrong bounded defaults")
	}
	for _, w := range plan.Widgets {
		if w.Query != nil && w.Query.Signal == "spans" {
			t.Fatal("invented span read without span inventory")
		}
		if w.Measurement != nil {
			if w.Grain != "per-stream" || !w.Measurement.Definition.Source.Namespace.Equal(scope.Namespace) || w.Measurement.FromMS != scope.FromMS || w.Measurement.ToMS != scope.ToMS {
				t.Fatal("native/span scope or grain merged")
			}
			unit := w.Measurement.Definition.Source.Unit
			if unit != "s" && unit != "ms" {
				t.Fatal("unit conversion/inference")
			}
		}
	}
	in.Instruments[0], in.Instruments[2] = in.Instruments[2], in.Instruments[0]
	for i := range in.Instruments {
		in.Instruments[i].Records += 1000
		in.Instruments[i].CapabilityPlan = &measurements.InstrumentCapabilityPlan{Profile: "generic", Coverage: measurements.InstrumentCoverage{TotalStreams: 99}}
	}
	again, e := exploration.PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	for i, option := range plan.NativeOptions {
		if option.Key != again.NativeOptions[i].Key || option.Selected != again.NativeOptions[i].Selected {
			t.Fatal("counts/hints changed stable native defaults")
		}
	}
	empty := []string{}
	in.NativeSelection = &empty
	cleared, e := exploration.PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, w := range cleared.Widgets {
		if w.Measurement != nil {
			t.Fatal("explicit empty selection read native data")
		}
	}
}
func TestHTTPDashboardFallbackAndUnknowns(t *testing.T) {
	base := exploration.DashboardRequest{Scope: exploration.InventoryQuestion{Environment: "test", FromMS: 1000, ToMS: 61000}, Service: "api", Profile: "auto"}
	for _, test := range []struct {
		name    string
		signals []string
		fields  []exploration.Field
		profile string
		Hint    bool
	}{
		{"missing all telemetry", nil, nil, "all-spans", false},
		{"legacy key does not activate", []string{"spans"}, []exploration.Field{{Kind: "attribute", Key: "http.method"}}, "all-spans", false},
		{"resource key is not span method", []string{"spans"}, []exploration.Field{{Kind: "resource", Key: "http.request.method"}}, "all-spans", false},
		{"current key is only candidate", []string{"spans"}, []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}, "all-spans", false},
		{"observed hint selects candidate", []string{"spans"}, []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}, "http-server", true},
		{"hint without current key does not activate", []string{"spans"}, nil, "all-spans", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			in := base
			in.Signals = test.signals
			in.Fields = test.fields
			in.HTTPServerObserved = test.Hint
			p, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || p.Profile != test.profile {
				t.Fatal(p, e)
			}
			for _, w := range p.Widgets {
				if len(test.signals) == 0 && (w.Query != nil || w.Measurement != nil || w.State != "not_observed") {
					t.Fatal("absence became executable/healthy")
				}
			}
			if test.profile == "http-server" {
				client := map[exploration.Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "column", Key: "kind"}: "Client", {Kind: "attribute", Key: "http.request.method"}: "GET"}
				for _, w := range p.Widgets {
					if w.Query != nil && w.Query.Signal == "spans" && hdPredicate(w.Query.Where, client) {
						t.Fatal("client-only evidence counted as server")
					}
				}
			}
		})
	}
	base.Profile = "http-server"
	if _, e := exploration.PlanServiceDashboard(base, nil); e == nil {
		t.Fatal("explicit unsupported profile accepted")
	}
	base.Profile = "rpc-guessed"
	if _, e := exploration.PlanServiceDashboard(base, nil); e == nil {
		t.Fatal("unknown profile accepted")
	}
}
