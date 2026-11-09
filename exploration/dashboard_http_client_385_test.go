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

func dashboardHTTPClient385() DashboardRequest {
	in := dashboardHTTPInput()
	in.Profile = "http-client"
	in.HTTPServerObserved = false
	in.Fields = []Field{{Kind: "attribute", Key: "http.request.method"}}
	return in
}

func TestHTTPClientDashboardExactPopulationAndOptionalDimensions385(t *testing.T) {
	fields := map[string]string{"spans": "name", "dimension": "server.address", "http-client-methods": "http.request.method", "http-client-ports": "server.port", "http-client-responses": "http.response.status_code"}
	for _, selection := range []*identity.EnvironmentSelector{nil,
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	} {
		for _, namespace := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "team'雪"}} {
			in := dashboardHTTPClient385()
			in.Service = "api' OR 1=1 -- 雪"
			in.Scope.Namespace = namespace
			in.Scope.EnvironmentSelector = selection
			if selection != nil {
				in.Scope.Environment = ""
			}
			p, err := PlanServiceDashboard(in, nil)
			if err != nil || p.Profile != "http-client" || !reflect.DeepEqual(p.Scope, in.Scope) {
				t.Fatal(p, err)
			}
			selected := 0
			for _, family := range p.Families {
				if family.Selected {
					selected++
					if family.Key != "http-client" || family.RuleSet != "goo/http-client-span-observations/v1" || family.Grain != "client-span-record" {
						t.Fatal(family)
					}
				}
			}
			if selected != 1 {
				t.Fatal("profile acquired multiple selected families", p.Families)
			}
			seen := map[string]bool{}
			for _, widget := range p.Widgets {
				if widget.Query == nil || widget.Query.Signal != "spans" {
					continue
				}
				seen[widget.Key] = true
				q := *widget.Query
				if widget.Family != "http-client" || widget.Grain != "client-span-record" || q.Summary != "span-duration" || !q.SpanStatus || len(q.GroupBy) != 1 || q.GroupBy[0].Key != fields[widget.Key] {
					t.Fatal(widget)
				}
				if q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS || !reflect.DeepEqual(q.EnvironmentSelector, selection) || q.Environment != in.Scope.Environment {
					t.Fatal("scope widened", q)
				}
				plan, err := Compile(q)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(plan.Rows, in.Service) {
					t.Fatal("service interpolated into SQL")
				}
				for _, kind := range []string{"Client", "Server", "Internal", "CLIENT", "SPAN_KIND_CLIENT", ""} {
					for _, method := range []string{"GET", "_OTHER", "custom'雪", " ", ""} {
						values := map[Field]string{{Kind: "column", Key: "service"}: in.Service, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "http.request.method"}: method}
						if namespace != nil && namespace.Mode == "value" {
							values[Field{Kind: "resource", Key: "service.namespace"}] = namespace.Value
						}
						want := kind == "Client" && method != ""
						if Match(q.Where, values) != want {
							t.Fatalf("%s kind=%q method=%q", widget.Key, kind, method)
						}
						// Optional values cannot become admission conditions or protocol policy.
						for _, key := range []string{"server.address", "server.port", "http.response.status_code"} {
							for _, value := range []string{"", "0", "443", "5xx", "500", "unknown'雪"} {
								values[Field{Kind: "attribute", Key: key}] = value
								if Match(q.Where, values) != want {
									t.Fatalf("optional %s=%q restricted the population", key, value)
								}
							}
							delete(values, Field{Kind: "attribute", Key: key})
						}
						values[Field{Kind: "attribute", Key: "http.method"}] = "GET"
						delete(values, Field{Kind: "attribute", Key: "http.request.method"})
						if Match(q.Where, values) {
							t.Fatal("legacy method alias admitted")
						}
					}
				}
			}
			if len(seen) != len(fields) {
				t.Fatal("missing typed dimension", seen)
			}
		}
	}
}

func TestHTTPClientDashboardExplicitOnlyMetadataAndNativeSeparation385(t *testing.T) {
	in := dashboardHTTPClient385()
	in.Profile = "auto"
	p, err := PlanServiceDashboard(in, nil)
	if err != nil || p.Profile != "all-spans" {
		t.Fatal("method metadata changed auto priority", p, err)
	}
	if !slices.ContainsFunc(p.Families, func(f DashboardFamily) bool { return f.Key == "http-client" && f.State == "candidate" && !f.Selected }) {
		t.Fatal("manual candidate absent", p)
	}
	in.HTTPServerObserved = true
	p, err = PlanServiceDashboard(in, nil)
	if err != nil || p.Profile != "http-server" {
		t.Fatal("client displaced incoming profile", p, err)
	}
	for _, fields := range [][]Field{nil, {{Kind: "attribute", Key: "http.method"}}, {{Kind: "resource", Key: "http.request.method"}}, {{Kind: "attribute", Key: "server.address"}}} {
		bad := dashboardHTTPClient385()
		bad.Fields = fields
		if _, err := PlanServiceDashboard(bad, nil); err == nil {
			t.Fatal("non-current metadata admitted", fields)
		}
	}
	metric := dashboardTestInstrument("native-client", "http.client.request.duration", "histogram", "s")
	in = dashboardHTTPClient385()
	in.Signals = []string{"spans", "metrics"}
	in.Instruments = []measurements.Instrument{metric}
	p, err = PlanServiceDashboard(in, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, widget := range p.Widgets {
		if widget.Measurement != nil && widget.Family == "http-client" {
			t.Fatal("generic native metric acquired client profile authority")
		}
	}
	in.Fields = nil
	if _, err = PlanServiceDashboard(in, nil); err == nil {
		t.Fatal("native client histogram enabled record profile")
	}
	in = dashboardHTTPClient385()
	in.Signals = []string{"metrics"}
	if _, err = PlanServiceDashboard(in, nil); err == nil {
		t.Fatal("method metadata without spans enabled profile")
	}
}

func TestHTTPClientDashboardDetachesAuthoredScope385(t *testing.T) {
	in := dashboardHTTPClient385()
	in.Scope.Environment = ""
	in.Scope.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
	in.Scope.Namespace = &identity.Namespace{Mode: "value", Value: "team"}
	before, _ := json.Marshal(in)
	plan, err := PlanServiceDashboard(in, nil)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("planner mutated request")
	}
	plan.Scope.EnvironmentSelector.Scope.Value = "other"
	plan.Scope.Namespace.Value = "other"
	for _, widget := range plan.Widgets {
		if widget.Query != nil && widget.Query.EnvironmentSelector != nil {
			widget.Query.EnvironmentSelector.Scope.Value = "changed"
		}
	}
	after, _ = json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("plan retained mutable request scope")
	}
}
