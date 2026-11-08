package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/workbench"
)

func TestTraceSearchCapture383HTTPQuestionEvidenceOffline(t *testing.T) {
	var reads atomic.Int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reads.Add(1) }))
	defer backend.Close()
	mux := http.NewServeMux()
	dir := t.TempDir()
	bank := &exploration.ResultSnapshots{}
	capturehttp.Register(mux, exploration.Source{Endpoint: backend.URL}, dir, bank)
	registerWorkbench(mux, dir, retentionConnection{}, bank, nil)
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		return w
	}
	request := exploration.TraceSearchRequest{Version: exploration.TraceSearchVersion, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 7, Conditions: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}}}
	w := call("POST", "/captures", map[string]any{"title": "Search evidence", "question": exploration.CaptureQuestion{Kind: "trace-search", TraceSearch: &request}})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var page exploration.CapturePage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || page.TraceSearch == nil || page.TraceSearch.State != "no_data" || reads.Load() != 1 {
		t.Fatal(page, err, reads.Load())
	}
	first := w.Body.String()
	backend.Close()
	w = call("GET", "/captures/"+page.Receipt.ID, nil)
	if w.Code != 200 || w.Body.String() != first || reads.Load() != 1 {
		t.Fatal("offline evidence changed", w.Code, w.Body.String())
	}
	step := &workbench.Step{Version: 1, Kind: "trace-search", View: "Traces", Scope: workbench.StepScope{Environment: "prod", FromMS: 1000, ToMS: 61000}, TraceSearch: &request}
	card := workbench.Card{Kind: "question", Key: strings.Repeat("a", 64), Step: step, Evidence: []string{page.Receipt.ID}}
	document := workbench.DocumentInput{Title: "Search investigation", Cards: []workbench.Card{card}}
	w = call("POST", "/documents", map[string]any{"document": document})
	if w.Code != 201 || reads.Load() != 1 {
		t.Fatal("compatible offline question evidence refused", w.Code, w.Body.String())
	}
	document.Cards[0].Step.TraceSearch.Limit = 8
	w = call("POST", "/documents", map[string]any{"document": document})
	if w.Code != 400 || reads.Load() != 1 {
		t.Fatal("foreign search evidence accepted", w.Code, w.Body.String())
	}
	w = call("POST", "/captures", map[string]any{"parent": page.Receipt.ID})
	if w.Code != 502 {
		t.Fatal("offline refresh disguised success", w.Code, w.Body.String())
	}
	w = call("GET", "/captures/"+page.Receipt.ID, nil)
	if w.Code != 200 || w.Body.String() != first {
		t.Fatal("failed refresh changed old evidence", w.Code, w.Body.String())
	}
}
