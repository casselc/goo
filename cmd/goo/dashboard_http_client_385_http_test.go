package main

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestHTTPClientDashboardPlanHTTP385(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}
	call := func(profile string, fields []exploration.Field, status int) exploration.DashboardPlan {
		t.Helper()
		response := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardProfileBody(profile, []string{"spans", "logs"}, fields)})
		if response.Code != status {
			t.Fatalf("%s: %d %s", profile, response.Code, response.Body.String())
		}
		var plan exploration.DashboardPlan
		if status == 200 {
			if err := json.Unmarshal(response.Body.Bytes(), &plan); err != nil {
				t.Fatal(err)
			}
		}
		return plan
	}
	plan := call("http-client", fields, 200)
	if plan.Profile != "http-client" || !slices.ContainsFunc(plan.Families, func(f exploration.DashboardFamily) bool {
		return f.Key == "http-client" && f.Selected && f.RuleSet == "goo/http-client-span-observations/v1"
	}) {
		t.Fatal(plan)
	}
	count := 0
	for _, widget := range plan.Widgets {
		if widget.Query == nil || widget.Query.Signal != "spans" {
			continue
		}
		count++
		values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "checkout", {Kind: "resource", Key: "service.namespace"}: "commerce", {Kind: "column", Key: "kind"}: "Client", {Kind: "attribute", Key: "http.request.method"}: "GET"}
		if widget.Family != "http-client" || widget.Grain != "client-span-record" || !exploration.Match(widget.Query.Where, values) {
			t.Fatal(widget)
		}
		for _, field := range []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "resource", Key: "service.namespace"}, {Kind: "column", Key: "kind"}, {Kind: "attribute", Key: "http.request.method"}} {
			bad := map[exploration.Field]string{}
			for k, v := range values {
				bad[k] = v
			}
			delete(bad, field)
			if exploration.Match(widget.Query.Where, bad) {
				t.Fatal("HTTP returned weakened predicate", field)
			}
		}
	}
	if count != 5 {
		t.Fatal("expected activity plus four independent typed dimensions", count)
	}
	if auto := call("auto", fields, 200); auto.Profile != "all-spans" {
		t.Fatal("outgoing metadata changed auto profile", auto.Profile)
	}
	call("http-client", nil, 400)
	call("http-client", []exploration.Field{{Kind: "attribute", Key: "http.method"}}, 400)
	call("http-client", []exploration.Field{{Kind: "resource", Key: "http.request.method"}}, 400)
}
