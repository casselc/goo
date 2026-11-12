package exploration

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
)

func TestHTTPServerResponseContext410(t *testing.T) {
	for _, selection := range []*identity.EnvironmentSelector{nil,
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	} {
		for _, ns := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "team'雪"}} {
			in := dashboardHTTPInput()
			in.Service = "api'雪"
			in.Scope.Namespace = ns
			in.Scope.EnvironmentSelector = selection
			if selection != nil {
				in.Scope.Environment = ""
			}
			// Response metadata is deliberately absent: absence is a population, not a
			// reason to hide the context or silently add a present-key predicate.
			in.Fields = []Field{{Kind: "attribute", Key: "http.request.method"}}
			plan, err := PlanServiceDashboard(in, nil)
			if err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(plan.Widgets, func(w DashboardWidget) bool { return w.Key == "http-server-responses" })
			if index < 0 {
				t.Fatal("response context absent without field metadata")
			}
			widget := plan.Widgets[index]
			if widget.Query == nil || widget.Family != "http-server" || widget.Grain != "server-span-record" || widget.State != "supported" {
				t.Fatal(widget)
			}
			q := *widget.Query
			if q.Signal != "spans" || q.Summary != "span-duration" || !q.SpanStatus || !reflect.DeepEqual(q.GroupBy, []Field{{Kind: "attribute", Key: "http.response.status_code"}}) || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS || q.Environment != in.Scope.Environment || !reflect.DeepEqual(q.EnvironmentSelector, selection) || q.Offset != 0 {
				t.Fatal("response scope or independent measures changed", q)
			}
			compiled, err := Compile(q)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(compiled.Rows, in.Service) {
				t.Fatal("service interpolated into SQL")
			}
			for _, kind := range []string{"Server", "Client", "Internal", "SPAN_KIND_SERVER", ""} {
				for _, method := range []string{"GET", "_OTHER", " ", ""} {
					values := map[Field]string{{Kind: "column", Key: "service"}: in.Service, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "http.request.method"}: method}
					if ns != nil && ns.Mode == "value" {
						values[Field{Kind: "resource", Key: "service.namespace"}] = ns.Value
					}
					want := kind == "Server" && method != ""
					for _, status := range []string{"Error", "Ok", "Unset", "unknown", ""} {
						values[Field{Kind: "column", Key: "status"}] = status
						for _, response := range []*string{nil, viewText(""), viewText("200"), viewText("500"), viewText("0500"), viewText("5xx"), viewText("unknown'雪")} {
							key := Field{Kind: "attribute", Key: "http.response.status_code"}
							if response == nil {
								delete(values, key)
							} else {
								values[key] = *response
							}
							if Match(q.Where, values) != want {
								t.Fatalf("literal response or span status changed membership: kind=%q method=%q status=%q response=%v", kind, method, status, response)
							}
						}
					}
					delete(values, Field{Kind: "attribute", Key: "http.request.method"})
					values[Field{Kind: "attribute", Key: "http.method"}] = "GET"
					if Match(q.Where, values) {
						t.Fatal("legacy method alias admitted")
					}
					values[Field{Kind: "attribute", Key: "http.request.method"}] = "GET"
					values[Field{Kind: "column", Key: "kind"}] = "Server"
					values[Field{Kind: "column", Key: "service"}] = "foreign"
					if Match(q.Where, values) {
						t.Fatal("foreign service admitted")
					}
					values[Field{Kind: "column", Key: "service"}] = in.Service
					if ns != nil {
						values[Field{Kind: "resource", Key: "service.namespace"}] = "foreign"
						if Match(q.Where, values) {
							t.Fatal("foreign namespace admitted")
						}
					}
				}
			}
		}
	}
}

func TestHTTPServerResponseContextProfileAdmission410(t *testing.T) {
	for _, profile := range []string{"auto", "all-spans", "http-client"} {
		in := dashboardHTTPInput()
		in.Profile = profile
		in.HTTPServerObserved = false
		plan, err := PlanServiceDashboard(in, nil)
		if err != nil {
			t.Fatal(err)
		}
		if slices.ContainsFunc(plan.Widgets, func(w DashboardWidget) bool { return w.Key == "http-server-responses" }) {
			t.Fatal("inactive server profile acquired response context", profile)
		}
	}
}

func TestHTTPServerResponseContextDoesNotInventSpanSignal410(t *testing.T) {
	for _, profile := range []string{"auto", "http-server"} {
		in := dashboardHTTPInput()
		in.Profile = profile
		in.Signals = []string{"metrics"}
		in.Fields = nil
		in.HTTPServerObserved = false
		in.Instruments = append(in.Instruments, dashboardTestInstrument("native-http", "http.server.request.duration", "histogram", "s"))
		plan, err := PlanServiceDashboard(in, nil)
		if err != nil || plan.Profile != "http-server" {
			t.Fatal("native HTTP profile unavailable", plan, err)
		}
		if slices.ContainsFunc(plan.Widgets, func(w DashboardWidget) bool { return w.Key == "http-server-responses" }) {
			t.Fatal("native instrument invented response-span read")
		}
	}
}
