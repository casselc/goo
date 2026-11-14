package celhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
)

func planningRequest(t *testing.T) []byte {
	t.Helper()
	request, err := querycel.NewPlanRequest(exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20}, `log.body.contains("x' OR 1=1 --")`, "filter", "mount-1", "9007199254740993")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPlanRouteBoundedSourceFree(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux)
	data := planningRequest(t)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/cel/plan", bytes.NewReader(data)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var out querycel.RecordPlan
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Binding.Revision != "9007199254740993" || strings.Contains(out.Plan.Rows, "x' OR 1=1 --") {
		t.Fatal(w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Header())
	}
	if policy, ok := PolicyFor("POST", "/cel/plan"); !ok || policy.BodyLimit != BodyLimit {
		t.Fatal(policy, ok)
	}
	if _, ok := PolicyFor("GET", "/cel/plan"); ok {
		t.Fatal("GET policy admitted")
	}
}

func TestPlanRouteRefusesWithoutPartialPlan(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux)
	data := planningRequest(t)
	var invalid map[string]any
	if err := json.Unmarshal(data, &invalid); err != nil {
		t.Fatal(err)
	}
	invalid["expression"] = `false && log.body.matches(".*")`
	badExpression, _ := json.Marshal(invalid)
	cases := []struct {
		name, path string
		body       []byte
	}{
		{"unsupported", "/cel/plan", badExpression},
		{"missing", "/cel/plan", []byte(`{}`)},
		{"null", "/cel/plan", []byte(`null`)},
		{"unknown", "/cel/plan", append(append([]byte{}, data[:len(data)-1]...), []byte(`,"extra":true}`)...)},
		{"trailing", "/cel/plan", append(append([]byte{}, data...), []byte(` {}`)...)},
		{"malformed", "/cel/plan", []byte(`{"base":`)},
		{"oversize", "/cel/plan", []byte(strings.Repeat(" ", int(BodyLimit)+1))},
		{"query", "/cel/plan?scope=other", data},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", tc.path, bytes.NewReader(tc.body)))
			if w.Code != 400 {
				t.Fatal(w.Code, w.Body.String())
			}
			var result map[string]json.RawMessage
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"question", "plan", "predicate", "binding"} {
				if _, ok := result[key]; ok {
					t.Fatal("partial plan", w.Body.String())
				}
			}
			if tc.name == "unsupported" {
				var out ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || out.Diagnostic == nil || out.Diagnostic.Code != querycel.CodeUnsupported {
					t.Fatal(err, w.Body.String())
				}
			}
		})
	}
	req := httptest.NewRequest("POST", "/cel/plan", bytes.NewReader(data))
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req.WithContext(ctx))
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
