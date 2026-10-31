package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestPopulationPlanTypedScopeHTTPWithoutSource(t *testing.T) {
	// No telemetry source is registered: planning is a pure configuration operation.
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, &exploration.ResultSnapshots{}, nil)
	call := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/population/plan", strings.NewReader(body)))
		return w
	}
	wires := []string{
		`{"version":"goo/environment-selector/v1","policy":"current-only","scope":{"mode":"named","value":"prod"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"empty"}}`,
		`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported"}}`,
	}
	question := func(selector string) string {
		return `{"signal":"spans","environment":"","environmentSelector":` + selector + `,"fromMs":1000,"toMs":61000,"limit":2,"offset":0,"groupBy":[{"kind":"column","key":"service"}],"where":{"any":[{"field":{"kind":"attribute","key":"zone"},"op":"missing"},{"field":{"kind":"attribute","key":"zone"},"op":"eq","value":""}]}}`
	}
	for _, wire := range wires {
		raw := question(wire)
		w := call(`{"question":` + raw + `}`)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var plan exploration.PopulationPlan
		var requested exploration.Query
		if json.Unmarshal(w.Body.Bytes(), &plan) != nil || json.Unmarshal([]byte(raw), &requested) != nil || len(plan.Queries) != 2 {
			t.Fatal("invalid plan", w.Body.String())
		}
		for _, q := range plan.Queries {
			if !q.SameEnvironment(requested) || !reflect.DeepEqual(q.EnvironmentSelector, requested.EnvironmentSelector) || !reflect.DeepEqual(q.Where, requested.Where) || q.FromMS != requested.FromMS || q.ToMS != requested.ToMS || q.Offset != 0 {
				t.Fatal("plan changed scope", q)
			}
		}
		var fields map[string]json.RawMessage
		json.Unmarshal(w.Body.Bytes(), &fields)
		if len(fields) != 3 || fields["version"] == nil || fields["queries"] == nil || fields["notices"] == nil {
			t.Fatal("planner exposed runtime data", w.Body.String())
		}
	}
	for _, bad := range []string{
		strings.Replace(question(wires[1]), `"environment":""`, `"environment":"prod"`, 1),
		question(strings.Replace(wires[1], `"effective"`, `"future"`, 1)),
		question(`null`),
	} {
		if w := call(`{"question":` + bad + `}`); w.Code != 400 {
			t.Fatal("invalid scope planned", w.Code, w.Body.String())
		}
	}
	companion := strings.Replace(question(wires[0]), `"signal":"spans"`, `"signal":"logs"`, 1)
	if w := call(`{"question":` + question(wires[1]) + `,"companion":` + companion + `}`); w.Code != 400 {
		t.Fatal("same literal name erased environment policy", w.Code, w.Body.String())
	}
	if w := call(`{"question":` + question(wires[1]) + `,"dropFilters":true}`); w.Code != 400 {
		t.Fatal("planner accepted selector weakening", w.Code, w.Body.String())
	}
}
