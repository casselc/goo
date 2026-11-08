package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
)

func TestTraceSearchStep382HTTPDocumentIntent(t *testing.T) {
	mux := http.NewServeMux()
	// There is deliberately no telemetry source registration or retention backend.
	// Saving and reopening authored intent must work without either capability.
	registerWorkbench(mux, t.TempDir(), retentionConnection{}, &exploration.ResultSnapshots{}, nil)
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		return w
	}
	selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}
	search := &exploration.TraceSearchRequest{Version: exploration.TraceSearchVersion, EnvironmentSelector: selector, FromMS: 1000, ToMS: 61000, Limit: 12, Conditions: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}, {Field: &exploration.Field{Kind: "column", Key: "duration_ns"}, Op: "gte", Value: "9007199254740993"}}}
	step := workbench.Step{Version: 1, Kind: "trace-search", View: "Traces", Scope: workbench.StepScope{EnvironmentSelector: selector, FromMS: search.FromMS, ToMS: search.ToMS}, TraceSearch: search}
	document := workbench.DocumentInput{Title: "Cross-service latency", Cards: []workbench.Card{{Kind: "question", Key: strings.Repeat("a", 64), Step: &step}}}
	response := call("POST", "/documents", map[string]any{"document": document})
	if response.Code != 201 || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(response.Code, response.Body.String())
	}
	var first workbench.Document
	if err := json.Unmarshal(response.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	before := response.Body.String()
	response = call("GET", "/documents/"+first.ID, nil)
	if response.Code != 200 || response.Body.String() != before {
		t.Fatal("saved intent changed on reopen", response.Code, response.Body.String())
	}
	if first.Cards[0].Step.TraceSearch.Conditions[1].Value != "9007199254740993" || first.Cards[0].Step.TraceSearch.Limit != 12 {
		t.Fatal("search shape changed")
	}
	document.Cards[0].Step.TraceSearch.Conditions[1].Value = "18446744073709551615"
	response = call("POST", "/documents", map[string]any{"parent": first.ID, "document": document})
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	response = call("GET", "/documents/"+first.ID, nil)
	if response.Code != 200 || response.Body.String() != before {
		t.Fatal("new revision rewrote original intent")
	}
	document.Cards[0].Evidence = []string{strings.Repeat("b", 64)}
	response = call("POST", "/documents", map[string]any{"document": document})
	if response.Code != 400 || !strings.Contains(response.Body.String(), "unavailable") {
		t.Fatal("unavailable search captured evidence accepted", response.Code, response.Body.String())
	}
	document.Cards[0].Evidence = nil
	raw, _ := json.Marshal(map[string]any{"document": document})
	for _, bad := range []string{strings.Replace(string(raw), `"traceSearch":`, `"record":null,"traceSearch":`, 1), strings.Replace(string(raw), `"conditions":`, `"result":{},"conditions":`, 1), strings.ReplaceAll(string(raw), `"fromMs":1000`, `"fromMs":null`)} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/documents", strings.NewReader(bad)))
		if w.Code != 400 {
			t.Fatal("invalid search wire admitted", w.Code, w.Body.String())
		}
	}
}
