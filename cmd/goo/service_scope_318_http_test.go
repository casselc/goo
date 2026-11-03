package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func TestTypedServiceScopeHTTP318(t *testing.T) {
	reads := 0
	var sqls []string
	var bindings []map[string][]string
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(r.Body)
		sqls = append(sqls, string(sql))
		bindings = append(bindings, r.URL.Query())
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, &exploration.ResultSnapshots{}, nil)
	registerDashboards(mux, nil)
	selectors := []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
	for _, selector := range selectors {
		for _, path := range []string{"/inventory", "/discovery"} {
			t.Run(path+selector.Policy+selector.Scope.Mode, func(t *testing.T) {
				q := exploration.InventoryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "value"}, FromMS: 100000, ToMS: 160000}
				raw, _ := json.Marshal(q)
				before := reads
				response := httptest.NewRecorder()
				mux.ServeHTTP(response, httptest.NewRequest("POST", path, strings.NewReader(string(raw))))
				expected := 1
				if path == "/discovery" {
					expected = 2
				}
				if response.Code != 200 || reads-before != expected || response.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("typed service source admission failed", response.Code, response.Body.String(), reads-before)
				}
				var result struct {
					Question exploration.InventoryQuestion `json:"question"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(result.Question, q) {
					t.Fatal("returned service question lost its typed identity")
				}
				for i, sql := range sqls[before:] {
					if strings.Contains(sql, "'deployment.environment'") != (selector.Policy == "effective") {
						t.Fatal("executed SQL lost environment policy")
					}
					_, present := bindings[before+i]["param_environment"]
					if present != (selector.Scope.Mode == "named") {
						t.Fatal("nameless scope acquired a named parameter")
					}
				}
			})
		}
	}
	good := `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported"}}`
	invalid := []string{
		`{"Environment":null,"environmentSelector":` + good + `,"fromMs":1,"toMs":60001}`,
		`{"environment":"prod","environmentSelector":` + good + `,"fromMs":1,"toMs":60001}`,
		`{"environment":null,"environmentSelector":` + good + `,"fromMs":1,"toMs":60001}`,
		`{"environment":"prod","environmentSelector":null,"fromMs":1,"toMs":60001}`,
		`{"environmentSelector":{"version":"goo/environment-selector/v1","policy":"current-only","scope":{"mode":"empty"}},"fromMs":1,"toMs":60001}`,
		`{"environmentSelector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported","unknown":1}},"fromMs":1,"toMs":60001}`,
		`{"environmentSelector":` + good + `,"fromMs":1,"toMs":60001,"namespace":{"mode":"missing","value":"not-missing"}}`,
	}
	for _, body := range invalid {
		for _, path := range []string{"/inventory", "/discovery", "/dashboards/plan"} {
			before := reads
			request := body
			if path == "/dashboards/plan" {
				request = `{"scope":` + body + `,"service":"api","signals":["spans"],"fields":[],"instruments":[]}`
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest("POST", path, strings.NewReader(request)))
			if response.Code != 400 || reads != before {
				t.Fatal("invalid service scope reached a source", path, response.Code)
			}
		}
	}
	for _, native := range []string{`"instruments":[{"supported":true}]`, `"instruments":[],"nativeSelection":["native"]`} {
		before := reads
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest("POST", "/dashboards/plan", strings.NewReader(`{"scope":{"environmentSelector":`+good+`,"fromMs":1,"toMs":60001},"service":"api","signals":["spans","metrics"],"fields":[],`+native+`}`)))
		if response.Code != 400 || reads != before {
			t.Fatal("unsupported native scope was silently projected", response.Code)
		}
	}
	before := reads
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest("POST", "/dashboards/plan", strings.NewReader(`{"scope":{"environmentSelector":`+good+`,"fromMs":1,"toMs":60001},"service":"api","signals":["spans","logs","metrics"],"fields":[],"instruments":[]}`)))
	var plan exploration.DashboardPlan
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &plan) != nil || plan.NativeCapability == nil || plan.NativeCapability.State != "gauge-only" || reads != before {
		t.Fatal("typed record plan/native capability missing", response.Code, response.Body.String())
	}
}
