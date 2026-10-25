package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func dashboardDBBody(profile any, fields []exploration.Field, hint any) string {
	var body map[string]any
	_ = json.Unmarshal([]byte(dashboardProfileBody(profile, []string{"spans", "logs"}, fields)), &body)
	body["dbClientObserved"] = hint
	raw, _ := json.Marshal(body)
	return string(raw)
}

func dashboardDBNativeBody(unit string) string {
	var body map[string]any
	_ = json.Unmarshal([]byte(dashboardProfileBody("auto", []string{"metrics"}, []exploration.Field{})), &body)
	d := measurements.Definition{APIVersion: measurements.Version, Name: "db-client-duration", Title: "DB client duration", Kind: "histogram", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Namespace: &identity.Namespace{Mode: "value", Value: "commerce"}, Services: []string{"checkout"}, Scope: "ordinary-sdk", Metric: "db.client.operation.duration", Unit: unit}}
	body["instruments"] = []measurements.Instrument{{Kind: "histogram", Scope: "ordinary-sdk", Metric: d.Source.Metric, Unit: unit, Temporality: 2, Supported: true, Definition: &d}}
	raw, _ := json.Marshal(body)
	return string(raw)
}

func TestDashboardDBProfileHTTPWireFixtures(t *testing.T) {
	db := []exploration.Field{{Kind: "attribute", Key: "db.system.name"}, {Kind: "attribute", Key: "db.operation.name"}}
	partial := []exploration.Field{{Kind: "attribute", Key: "db.system"}}
	mixed := append(append([]exploration.Field{}, db...), exploration.Field{Kind: "attribute", Key: "http.request.method"})
	both := dashboardDBBody("auto", mixed, true)
	var body map[string]any
	_ = json.Unmarshal([]byte(both), &body)
	body["httpServerObserved"] = true
	raw, _ := json.Marshal(body)
	rpcBody := map[string]any{}
	_ = json.Unmarshal([]byte(dashboardDBBody("auto", append(append([]exploration.Field{}, db...), exploration.Field{Kind: "attribute", Key: "rpc.system"}, exploration.Field{Kind: "attribute", Key: "rpc.method"}), true)), &rpcBody)
	rpcBody["rpcServerObserved"] = true
	rpcRaw, _ := json.Marshal(rpcBody)
	cases := []planningWireCase{
		{Name: "DB-flat-keys-candidate-only", Body: dashboardDBBody("auto", db, false), Status: 200},
		{Name: "DB-positive-client-hint", Body: dashboardDBBody("auto", db, true), Status: 200},
		{Name: "explicit-DB-without-positive-hint", Body: dashboardDBBody("db-client", db, false), Status: 200},
		{Name: "explicit-all-spans-DB-candidate", Body: dashboardDBBody("all-spans", db, true), Status: 200},
		{Name: "DB-legacy-key-not-admitted", Body: dashboardDBBody("auto", partial, true), Status: 200},
		{Name: "explicit-DB-legacy-key-refused", Body: dashboardDBBody("db-client", partial, true), Status: 400},
		{Name: "both-positive-hints-HTTP-precedence", Body: string(raw), Status: 200},
		{Name: "RPC-positive-hint-before-DB", Body: string(rpcRaw), Status: 200},
		{Name: "DB-native-seconds-auto", Body: dashboardDBNativeBody("s"), Status: 200},
		{Name: "DB-native-wrong-unit-no-admission", Body: dashboardDBNativeBody("ms"), Status: 200},
		{Name: "DB-wrong-hint-type", Body: dashboardDBBody("auto", db, "yes"), Status: 400},
	}
	for i := range cases {
		cases[i].Method = "POST"
		cases[i].Path = "/dashboards/plan"
	}
	const filename = "../../testdata/http/dashboard-db-profiles.json"
	if os.Getenv("GOO_UPDATE_DASHBOARD_DB_HTTP_FIXTURES") == "1" {
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

			if c.Status == 200 {
				var plan exploration.DashboardPlan
				if err := json.Unmarshal(w.Body.Bytes(), &plan); err != nil {
					t.Fatal(err)
				}
				expected := map[string]string{"DB-flat-keys-candidate-only": "all-spans", "DB-positive-client-hint": "db-client", "explicit-DB-without-positive-hint": "db-client", "explicit-all-spans-DB-candidate": "all-spans", "DB-legacy-key-not-admitted": "all-spans", "both-positive-hints-HTTP-precedence": "http-server", "RPC-positive-hint-before-DB": "rpc-server", "DB-native-seconds-auto": "db-client", "DB-native-wrong-unit-no-admission": "all-spans"}[c.Name]
				if expected == "" || plan.Profile != expected {
					t.Fatal("profile admission changed", c.Name, plan.Profile, expected)
				}
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("header changed", k)
				}
			}
		})
	}
}

func TestDashboardDBProfileExactScopedClientPredicate(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "db.system.name"}}
	w := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardDBBody("db-client", fields, true)})
	var plan exploration.DashboardPlan
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || plan.Profile != "db-client" {
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
			kind, system, namespace string
			present, want           bool
		}{
			{"Client", "postgresql", "commerce", true, true}, {"Server", "postgresql", "commerce", true, false},
			{"Internal", "postgresql", "commerce", true, false}, {"Client", "", "commerce", true, false},
			{"Client", "", "commerce", false, false}, {"Client", "postgresql", "billing", true, false},
		} {
			values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "checkout", {Kind: "column", Key: "kind"}: sample.kind, {Kind: "resource", Key: "service.namespace"}: sample.namespace}
			if sample.present {
				values[exploration.Field{Kind: "attribute", Key: "db.system.name"}] = sample.system
			}
			if got := exploration.Match(q.Where, values); got != sample.want {
				t.Fatal("DB population mismatch", widget.Key, sample, got)
			}
			if !sample.want {
				values[exploration.Field{Kind: "attribute", Key: "db.system"}] = "postgresql"
				if exploration.Match(q.Where, values) {
					t.Fatal("legacy alias admitted", sample)
				}
			}
		}
		if _, err := exploration.Compile(*q); err != nil {
			t.Fatal(err)
		}
	}
	if queries == 0 {
		t.Fatal("no executable DB span question")
	}
}

func TestDashboardDBHintCannotActivateOmittedProfile(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "db.system.name"}}
	var baseline string
	for _, profile := range []any{"", nil} {
		for _, hint := range []any{false, true} {
			w := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardDBBody(profile, fields, hint)})
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
