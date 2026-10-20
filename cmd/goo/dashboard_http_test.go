package main

import (
	"github.com/casselc/goo/annotations"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardHTTPPlansWithoutSourceAndMatchesRegistryDeclarations(t *testing.T) {
	mux := http.NewServeMux()
	p := &annotations.OperationPlan{Revision: strings.Repeat("a", 64)}
	p.Definition.Name = "declared"
	p.Definition.Title = "Declared service"
	p.Definition.Resource.Service = "api"
	p.Definition.Resource.Environment = "prod"
	registerDashboards(mux, []*annotations.OperationPlan{p})
	for _, body := range []string{`{"scope":{"environment":"prod","fromMs":1000,"toMs":2000},"service":"api","signals":[],"fields":[],"instruments":[]}`, `{"scope":{"environment":"prod","fromMs":1000,"toMs":2000},"service":"other","signals":["logs"]}`} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/dashboards/plan", strings.NewReader(body)))
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if strings.Contains(body, `"service":"api"`) != strings.Contains(w.Body.String(), `"name":"declared"`) {
			t.Fatal(w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/dashboards/plan", strings.NewReader(`{"unexpected":true}`)))
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
