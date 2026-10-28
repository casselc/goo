package capturehttp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/internal/explanationhttp"
	"github.com/casselc/goo/internal/queryhttp"
)

func TestAdmissionPreservesHeldPublicationAndPurePlanBoundaries(t *testing.T) {
	mux := http.NewServeMux()
	bank := &exploration.ResultSnapshots{}
	capturehttp.Register(mux, exploration.Source{}, t.TempDir(), bank)
	explanationhttp.Register(mux, exploration.Source{}, bank)
	budget := queryhttp.NewBudget(mux, 0, func(r *http.Request) bool {
		for _, lookup := range []func(string, string) (queryhttp.Policy, bool){capturehttp.PolicyFor, explanationhttp.PolicyFor} {
			if policy, ok := lookup(r.Method, r.URL.Path); ok {
				return policy.Effect != queryhttp.Pure
			}
		}
		return false
	})
	call := func(method, path string, input any) *httptest.ResponseRecorder {
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		budget.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(raw)))
		return w
	}
	for _, tc := range []struct {
		path string
		body any
	}{
		{"/captures", map[string]any{"title": "Held observation", "resultToken": strings.Repeat("a", 64)}},
		{"/captures/live", map[string]any{"freeze": strings.Repeat("b", 64)}},
		{"/captures/live", map[string]any{"baseline": strings.Repeat("c", 64)}},
		{"/population/explain", map[string]any{}},
	} {
		if w := call("POST", tc.path, tc.body); w.Code != 429 {
			t.Fatalf("mixed/source route bypassed existing admission: %s %d %s", tc.path, w.Code, w.Body.String())
		}
	}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
	selected := q
	selected.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "worker"}
	request := exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: q, Selection: selected, Fields: []exploration.Field{{Kind: "attribute", Key: "region"}}}
	for _, tc := range []struct {
		path string
		body any
	}{
		{"/population/explain/plan", request},
		{"/population/explain/filter", map[string]any{"request": request, "field": request.Fields[0], "value": nil, "side": "selection", "mode": "exclude"}},
	} {
		if w := call("POST", tc.path, tc.body); w.Code != 200 {
			t.Fatalf("pure plan inherited source admission: %s %d %s", tc.path, w.Code, w.Body.String())
		}
	}
	if w := call("GET", "/captures/not-an-id", nil); w.Code != 404 {
		t.Fatal("saved-capture reads inherited POST admission", w.Code)
	}
	for _, lookup := range []func(string, string) (queryhttp.Policy, bool){capturehttp.PolicyFor, explanationhttp.PolicyFor} {
		for _, route := range []struct{ method, path string }{{"GET", "/captures"}, {"POST", "/captures/unknown"}, {"POST", "/population/explain/unknown"}, {"POST", "/tail"}, {"POST", "/retained-evidence"}} {
			if _, ok := lookup(route.method, route.path); ok {
				t.Fatal("family claimed an unrelated route", route)
			}
		}
	}
}
