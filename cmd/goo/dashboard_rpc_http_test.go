package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/exploration"
)

func dashboardRPCBody(profile any, fields []exploration.Field, hint any) string {
	var body map[string]any
	_ = json.Unmarshal([]byte(dashboardProfileBody(profile, []string{"spans", "logs"}, fields)), &body)
	body["rpcServerObserved"] = hint
	raw, _ := json.Marshal(body)
	return string(raw)
}

func TestDashboardRPCProfileHTTPWireFixtures(t *testing.T) {
	rpc := []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}, {Kind: "attribute", Key: "rpc.service"}}
	partial := []exploration.Field{{Kind: "attribute", Key: "rpc.method"}}
	mixed := append(append([]exploration.Field{}, rpc...), exploration.Field{Kind: "attribute", Key: "http.request.method"})
	both := dashboardRPCBody("auto", mixed, true)
	var body map[string]any
	_ = json.Unmarshal([]byte(both), &body)
	body["httpServerObserved"] = true
	raw, _ := json.Marshal(body)
	cases := []planningWireCase{
		{Name: "RPC-flat-keys-candidate-only", Body: dashboardRPCBody("auto", rpc, false), Status: 200},
		{Name: "RPC-positive-server-hint", Body: dashboardRPCBody("auto", rpc, true), Status: 200},
		{Name: "explicit-RPC-without-positive-hint", Body: dashboardRPCBody("rpc-server", rpc, false), Status: 200},
		{Name: "explicit-all-spans-RPC-candidate", Body: dashboardRPCBody("all-spans", rpc, true), Status: 200},
		{Name: "RPC-partial-keys-not-admitted", Body: dashboardRPCBody("auto", partial, true), Status: 200},
		{Name: "explicit-RPC-partial-keys-refused", Body: dashboardRPCBody("rpc-server", partial, true), Status: 400},
		{Name: "both-positive-hints-HTTP-precedence", Body: string(raw), Status: 200},
		{Name: "RPC-wrong-hint-type", Body: dashboardRPCBody("auto", rpc, "yes"), Status: 400},
	}
	for i := range cases {
		cases[i].Method = "POST"
		cases[i].Path = "/dashboards/plan"
	}
	const filename = "../../testdata/http/dashboard-rpc-profiles.json"
	if os.Getenv("GOO_UPDATE_DASHBOARD_RPC_HTTP_FIXTURES") == "1" {
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
		out, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		out = append(out, '\n')
		if err := os.WriteFile(filename, out, 0644); err != nil {
			t.Fatal(err)
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
					t.Fatal("header changed", k)
				}
			}
		})
	}
}

func TestDashboardRPCProfileExactScopedServerPredicate(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}
	w := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardRPCBody("rpc-server", fields, true)})
	var plan exploration.DashboardPlan
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || plan.Profile != "rpc-server" {
		t.Fatal(w.Code, w.Body.String())
	}
	queries := 0
	for _, widget := range plan.Widgets {
		q := widget.Query
		if q == nil || q.Signal != "spans" {
			continue
		}
		queries++
		if q.Environment != "prod" || q.FromMS != 1000 || q.ToMS != 61000 {
			t.Fatal("scope changed", q)
		}
		for _, sample := range []struct {
			kind, system, method, namespace string
			want                            bool
		}{
			{"Server", "grpc", "Get", "commerce", true}, {"Client", "grpc", "Get", "commerce", false},
			{"Internal", "grpc", "Get", "commerce", false}, {"Server", "", "Get", "commerce", false},
			{"Server", "grpc", "", "commerce", false}, {"Server", "grpc", "Get", "billing", false},
		} {
			values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "checkout", {Kind: "column", Key: "kind"}: sample.kind, {Kind: "resource", Key: "service.namespace"}: sample.namespace}
			if sample.system != "" {
				values[exploration.Field{Kind: "attribute", Key: "rpc.system"}] = sample.system
			}
			if sample.method != "" {
				values[exploration.Field{Kind: "attribute", Key: "rpc.method"}] = sample.method
			}
			if got := exploration.Match(q.Where, values); got != sample.want {
				t.Fatal("RPC population mismatch", widget.Key, sample, got)
			}
		}
		if _, err := exploration.Compile(*q); err != nil {
			t.Fatal(err)
		}
	}
	if queries == 0 {
		t.Fatal("no executable RPC span question")
	}
}

func TestDashboardRPCHintCannotActivateOmittedProfile(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "rpc.system"}, {Kind: "attribute", Key: "rpc.method"}}
	var baseline string
	for _, profile := range []any{"", nil} {
		for _, hint := range []any{false, true} {
			w := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardRPCBody(profile, fields, hint)})
			var plan exploration.DashboardPlan
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || plan.Profile != "" || len(plan.Families) != 0 {
				t.Fatal("hint activated legacy", w.Code, w.Body.String())
			}
			if baseline == "" {
				baseline = w.Body.String()
			} else if baseline != w.Body.String() {
				t.Fatal("hint changed legacy bytes")
			}
		}
	}
}
