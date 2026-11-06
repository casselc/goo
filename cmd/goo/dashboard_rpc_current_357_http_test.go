package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestDashboardCurrentRPCProfileHTTPWireFixtures357(t *testing.T) {
	current := []exploration.Field{{Kind: "attribute", Key: "rpc.system.name"}}
	legacy := []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}
	mixed := append(append([]exploration.Field{}, current...), legacy...)
	httpMixed := append(append([]exploration.Field{}, mixed...), exploration.Field{Kind: "attribute", Key: "http.request.method"})
	body := func(profile any, fields []exploration.Field, hint, oldHint, httpHint any, selector any) string {
		var value map[string]any
		if err := json.Unmarshal([]byte(dashboardProfileBody(profile, []string{"spans", "logs"}, fields)), &value); err != nil {
			t.Fatal(err)
		}
		value["rpcServerCurrentObserved"], value["rpcServerObserved"], value["httpServerObserved"] = hint, oldHint, httpHint
		if selector != nil {
			scope := value["scope"].(map[string]any)
			delete(scope, "environment")
			scope["environmentSelector"] = selector
		}
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	selector := func(mode string) any {
		scope := map[string]any{"mode": mode}
		if mode == "named" {
			scope["value"] = "prod"
		}
		return map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": scope}
	}
	cases := []planningWireCase{
		{Name: "current-flat-key-candidate", Body: body("auto", current, false, false, false, nil), Status: 200},
		{Name: "current-positive-same-row-hint", Body: body("auto", current, true, false, false, nil), Status: 200},
		{Name: "current-explicit-candidate", Body: body("rpc-server-current", current, false, false, false, nil), Status: 200},
		{Name: "explicit-all-spans", Body: body("all-spans", current, true, false, false, nil), Status: 200},
		{Name: "legacy-distinct-population", Body: body("auto", legacy, false, true, false, nil), Status: 200},
		{Name: "current-before-legacy", Body: body("auto", mixed, true, true, false, nil), Status: 200},
		{Name: "HTTP-before-current", Body: body("auto", httpMixed, true, true, true, nil), Status: 200},
		{Name: "method-only-not-current", Body: body("auto", []exploration.Field{{Kind: "attribute", Key: "rpc.method"}}, true, false, false, nil), Status: 200},
		{Name: "explicit-current-legacy-only-refused", Body: body("rpc-server-current", legacy, true, true, false, nil), Status: 400},
		{Name: "current-hint-wrong-type", Body: body("auto", current, "yes", false, false, nil), Status: 400},
		{Name: "current-hint-null-is-unknown", Body: body("auto", current, nil, false, false, nil), Status: 200},
		{Name: "current-effective-named", Body: body("rpc-server-current", current, true, false, false, selector("named")), Status: 200},
		{Name: "current-effective-empty", Body: body("rpc-server-current", current, true, false, false, selector("empty")), Status: 200},
		{Name: "current-effective-unreported", Body: body("rpc-server-current", current, true, false, false, selector("unreported")), Status: 200},
	}
	for i := range cases {
		cases[i].Method, cases[i].Path = "POST", "/dashboards/plan"
	}
	const filename = "../../testdata/http/dashboard-current-rpc-profiles-357.json"
	if os.Getenv("GOO_UPDATE_DASHBOARD_CURRENT_RPC_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := planningWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[key] = w.Header().Get(key)
			}
		}
		raw, err := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filename, append(raw, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected := map[string]string{
		"current-flat-key-candidate": "all-spans", "current-positive-same-row-hint": "rpc-server-current",
		"current-explicit-candidate": "rpc-server-current", "explicit-all-spans": "all-spans",
		"legacy-distinct-population": "rpc-server", "current-before-legacy": "rpc-server-current",
		"HTTP-before-current": "http-server", "method-only-not-current": "all-spans",
		"current-hint-null-is-unknown": "all-spans", "current-effective-named": "rpc-server-current",
		"current-effective-empty": "rpc-server-current", "current-effective-unreported": "rpc-server-current",
	}
	for _, c := range dashboardProfileFixture(t, filename, len(cases)).Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := planningWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}
			for key, value := range c.Headers {
				if w.Header().Get(key) != value {
					t.Fatal("header changed", key)
				}
			}
			if c.Status == 200 {
				var plan exploration.DashboardPlan
				if err := json.Unmarshal(w.Body.Bytes(), &plan); err != nil || plan.Profile != expected[c.Name] {
					t.Fatal("profile population changed", plan.Profile, err)
				}
			}
		})
	}
}
