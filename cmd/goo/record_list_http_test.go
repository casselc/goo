package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func TestRecordListHTTPConfigurationRoundTripHasNoSourceRead(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(body)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		return response
	}
	plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "http-server", HTTPServerObserved: true, Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var input workbench.ViewTemplateExpansionRequest
	for _, widget := range plan.Widgets {
		if widget.Key == "dimension" {
			input = workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: widget}
		}
	}
	response := call("POST", "/view-templates/expand", input)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var expansion workbench.ViewTemplateExpansion
	if err = json.Unmarshal(response.Body.Bytes(), &expansion); err != nil {
		t.Fatal(err)
	}
	for i := range expansion.View.Composition.Children {
		if records := expansion.View.Composition.Children[i].Records; records != nil {
			records.Kind.Name = "record-list"
		}
	}
	response = call("POST", "/exploration-views", expansion.View)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var saved, loaded workbench.ExplorationView
	if err = json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	response = call("GET", "/exploration-views/"+saved.ID, nil)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	if err = json.Unmarshal(response.Body.Bytes(), &loaded); err != nil || !reflect.DeepEqual(saved, loaded) {
		t.Fatal("list publication changed on reload", err)
	}
	query := loaded.Queries[0]
	query.GroupBy = []exploration.Field{{Kind: "resource", Key: "region"}}
	query.Offset = 100
	response = call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &loaded, Query: &query})
	if response.Code != 400 || !strings.Contains(response.Body.String(), "page-zero root") || reads != 0 {
		t.Fatal("continuation was rewritten or accessed a source", response.Code, response.Body.String(), reads)
	}
	// Reusing a view requires an explicit page-zero destination. Configuration
	// expansion cannot silently turn a continuation into a different question.
	query.Offset = 0
	response = call("POST", "/view-templates/expand", workbench.ViewTemplateExpansionRequest{Source: &loaded, Query: &query})
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	if err = json.Unmarshal(response.Body.Bytes(), &expansion); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expansion.View.Queries[0], query) {
		t.Fatal("list expansion changed target scope")
	}
	for _, child := range expansion.View.Composition.Children {
		if child.Records != nil && child.Records.Kind.Name != "record-list" {
			t.Fatal("list kind lost")
		}
	}
	if reads != 0 {
		t.Fatal("configuration accessed a telemetry source", reads)
	}
}
