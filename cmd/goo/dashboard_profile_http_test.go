package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func dashboardProfileBody(profile any, signals []string, fields []exploration.Field) string {
	body, _ := json.Marshal(map[string]any{"profile": profile, "scope": map[string]any{"environment": "prod", "fromMs": 1000, "toMs": 61000, "namespace": map[string]string{"mode": "value", "value": "commerce"}}, "service": "checkout", "signals": signals, "fields": fields, "instruments": []any{}})
	return string(body)
}
func dashboardProfileBodyHint(profile any, signals []string, fields []exploration.Field, hint any) string {
	var body map[string]any
	_ = json.Unmarshal([]byte(dashboardProfileBody(profile, signals, fields)), &body)
	body["httpServerObserved"] = hint
	raw, _ := json.Marshal(body)
	return string(raw)
}
func dashboardProfileFixture(t *testing.T, filename string, count int) planningWireFixtures {
	t.Helper()
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var fixture planningWireFixtures
	if err = json.Unmarshal(raw, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) != count {
		t.Fatal("invalid dashboard fixture", err, len(fixture.Cases))
	}
	return fixture
}
func TestDashboardOmittedProfileLegacyHTTPBytes(t *testing.T) {
	// Frozen from e6ca91a before the profile fields/implementation were added.
	fixture := dashboardProfileFixture(t, "../../testdata/http/dashboard-legacy-profile.json", 1)
	for _, c := range fixture.Cases {
		w := planningWireResponse(t, c)
		if w.Code != c.Status || w.Body.String() != c.Response {
			t.Fatal("omitted profile changed legacy response", w.Code, w.Body.String())
		}
	}
}
func TestDashboardProfileHTTPWireFixtures(t *testing.T) {
	current := []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "http.route"}}
	legacy := []exploration.Field{{Kind: "attribute", Key: "http.method"}, {Kind: "attribute", Key: "http.route"}}
	cases := []planningWireCase{
		{Name: "auto-current-HTTP-candidate", Body: dashboardProfileBody("auto", []string{"spans", "logs"}, current), Status: 200},
		{Name: "auto-positive-server-hint", Body: dashboardProfileBodyHint("auto", []string{"spans", "logs"}, current, true), Status: 200},
		{Name: "explicit-all-spans", Body: dashboardProfileBody("all-spans", []string{"spans", "logs"}, current), Status: 200},
		{Name: "explicit-HTTP-server", Body: dashboardProfileBody("http-server", []string{"spans", "logs"}, current), Status: 200},
		{Name: "auto-legacy-keys-no-alias", Body: dashboardProfileBody("auto", []string{"spans"}, legacy), Status: 200},
		{Name: "auto-no-span-signal", Body: dashboardProfileBody("auto", []string{"logs"}, current), Status: 200},
		{Name: "explicit-HTTP-without-current-candidate", Body: dashboardProfileBody("http-server", []string{"spans"}, legacy), Status: 400},
		{Name: "unsupported-profile", Body: dashboardProfileBody("rpc-server", []string{"spans"}, current), Status: 400},
		{Name: "wrong-profile-type", Body: dashboardProfileBody(true, []string{"spans"}, current), Status: 400},
		{Name: "wrong-server-hint-type", Body: dashboardProfileBodyHint("auto", []string{"spans"}, current, "yes"), Status: 400},
	}
	const filename = "../../testdata/http/dashboard-profiles.json"
	for i := range cases {
		cases[i].Method = "POST"
		cases[i].Path = "/dashboards/plan"
	}
	if os.Getenv("GOO_UPDATE_DASHBOARD_PROFILE_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := planningWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal("fixture status", cases[i].Name, w.Code, w.Body.String())
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
	fixture := dashboardProfileFixture(t, filename, 10)
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := planningWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal("dashboard profile wire changed", w.Code, w.Body.String())
			}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				if w.Header().Get(k) != c.Headers[k] {
					t.Fatal("dashboard profile header changed", k)
				}
			}
		})
	}
}
func TestDashboardHTTPProfileReturnsExactScopedServerPredicate(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "http.route"}}
	c := planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardProfileBodyHint("auto", []string{"spans", "logs"}, fields, true)}
	response := planningWireResponse(t, c)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var plan exploration.DashboardPlan
	if err := json.Unmarshal(response.Body.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Profile != "http-server" {
		t.Fatal("auto didn't select HTTP candidate", plan.Profile)
	}
	candidate := false
	for _, family := range plan.Families {
		if family.Key == "http-server" {
			candidate = family.State == "candidate" && family.Selected
		}
	}
	if !candidate {
		t.Fatal("flat metadata advertised observed/conformant HTTP", plan.Families)
	}
	queries := 0
	for _, widget := range plan.Widgets {
		if widget.Query == nil || widget.Query.Signal != "spans" {
			continue
		}
		queries++
		q := widget.Query
		if q.Environment != "prod" || q.FromMS != 1000 || q.ToMS != 61000 {
			t.Fatal("HTTP widget scope changed", q)
		}
		for _, sample := range []struct {
			kind, method, namespace string
			present, want           bool
		}{
			{"Server", "GET", "commerce", true, true}, {"Client", "GET", "commerce", true, false},
			{"Internal", "GET", "commerce", true, false}, {"Server", "", "commerce", true, false},
			{"Server", "", "commerce", false, false}, {"Server", "GET", "billing", true, false},
		} {
			values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "checkout", {Kind: "column", Key: "kind"}: sample.kind, {Kind: "resource", Key: "service.namespace"}: sample.namespace}
			if sample.present {
				values[exploration.Field{Kind: "attribute", Key: "http.request.method"}] = sample.method
			}
			if got := exploration.Match(q.Where, values); got != sample.want {
				t.Fatal("HTTP question population mismatch", sample, got)
			}
		}
		compiled, err := exploration.Compile(*q)
		if err != nil || !strings.Contains(compiled.Rows, "SpanKind") {
			t.Fatal("HTTP question not supported by existing source adapter", err)
		}
	}
	if queries == 0 {
		t.Fatal("HTTP candidate had no executable span questions")
	}
}

func TestDashboardHTTPFlatKeysOfferCandidateWithoutAutomaticAdmission(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "http.route"}}
	response := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardProfileBody("auto", []string{"spans"}, fields)})
	var plan exploration.DashboardPlan
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &plan) != nil || plan.Profile != "all-spans" {
		t.Fatal("flat keys invented positive server evidence", response.Code, plan.Profile)
	}
	offered := false
	for _, family := range plan.Families {
		if family.Key == "http-server" {
			offered = family.State == "candidate" && !family.Selected
		}
	}
	if !offered {
		t.Fatal("flat metadata candidate hidden or marked observed", plan.Families)
	}
}

func TestDashboardEmptyOrNullProfileRetainsLegacyHTTPBytes(t *testing.T) {
	fixture := dashboardProfileFixture(t, "../../testdata/http/dashboard-legacy-profile.json", 1)
	base := fixture.Cases[0]
	for _, profile := range []any{"", nil} {
		var request map[string]any
		if err := json.Unmarshal([]byte(base.Body), &request); err != nil {
			t.Fatal(err)
		}
		request["profile"] = profile
		request["httpServerObserved"] = true
		body, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		c := base
		c.Body = string(body)
		response := planningWireResponse(t, c)
		if response.Code != 200 || response.Body.String() != base.Response {
			t.Fatal("empty/null profile changed legacy result", profile, response.Code)
		}
	}
}
