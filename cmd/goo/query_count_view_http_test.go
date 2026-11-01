package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func TestFixedQueryCountViewHTTPArtifactAndCLI308(t *testing.T) {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "selected"}}
	baseline := q
	baseline.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "baseline"}
	scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: "logs"}
	view := workbench.ExplorationView{Version: 1, Title: "Fixed query cohorts", Queries: []exploration.Query{q}, Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
		{Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.ViewKind{Name: "bar", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count"}}},
		{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: "record-table", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}},
		{Comparison: &workbench.CountComparisonViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.CountComparisonViewKind{Name: "query-count-comparison", Version: 1}, BaselineQuery: &baseline, Display: workbench.CountComparisonViewDisplay{Value: "count", View: "groups"}}},
	}}}
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerWorkbench(mux, t.TempDir(), retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	observations := []apiObservation{}
	call := func(method, path string, input any, want int) json.RawMessage {
		t.Helper()
		raw, _ := json.Marshal(input)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(raw)))
		if w.Code != want {
			t.Fatal(path, w.Code, w.Body.String())
		}
		if method == "POST" {
			observations = append(observations, apiObservation{Method: method, Path: path, Status: want, Request: raw, Response: append(json.RawMessage(nil), w.Body.Bytes()...)})
		}
		return w.Body.Bytes()
	}
	savedRaw := call("POST", "/exploration-views", view, 201)
	var saved workbench.ExplorationView
	if err := json.Unmarshal(savedRaw, &saved); err != nil {
		t.Fatal(err)
	}
	loadedRaw := call("GET", "/exploration-views/"+saved.ID, nil, 200)
	var loaded workbench.ExplorationView
	if err := json.Unmarshal(loadedRaw, &loaded); err != nil {
		t.Fatal(err)
	}
	loaded.ID = ""
	if !reflect.DeepEqual(loaded, view) {
		t.Fatal("fixed operands changed on ordinary save/open")
	}
	artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "platform", Name: "fixed-cohorts", Revision: "v1", View: loaded}
	encodedRaw := call("POST", "/view-artifacts/encode", workbench.ViewArtifactEncodeRequest{Artifact: artifact}, 200)
	var encoded workbench.ViewArtifactResult
	if err := json.Unmarshal(encodedRaw, &encoded); err != nil {
		t.Fatal(err)
	}
	decodedRaw := call("POST", "/view-artifacts/decode", workbench.ViewArtifactDecodeRequest{YAML: encoded.YAML}, 200)
	var decoded workbench.ViewArtifactResult
	if err := json.Unmarshal(decodedRaw, &decoded); err != nil || !reflect.DeepEqual(decoded.Artifact, artifact) {
		t.Fatal("fixed query artifact changed", err)
	}
	if reads != 0 {
		t.Fatal("save/codec acquired source", reads)
	}
	http.DefaultTransport = previous
	file := filepath.Join(t.TempDir(), "cohorts.yaml")
	if err := os.WriteFile(file, []byte(encoded.YAML), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"validate-view", "-file", file}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	destinationDir := t.TempDir()
	destination := http.NewServeMux()
	registerWorkbench(destination, destinationDir, retentionConnection{}, &exploration.ResultSnapshots{}, nil)
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		if r.Method != "POST" || r.URL.Path != "/api/exploration-views" {
			t.Error("CLI did more than explicit import")
		}
		http.StripPrefix("/api", destination).ServeHTTP(w, r)
	}))
	defer server.Close()
	var receipt bytes.Buffer
	if err := run([]string{"import-view", "-file", file, "-url", server.URL}, &receipt); err != nil {
		t.Fatal(err)
	}
	var imported viewArtifactCLIReceipt
	if err := json.Unmarshal(receipt.Bytes(), &imported); err != nil || posts != 1 || imported.ID != saved.ID {
		t.Fatal("fixed query CLI import identity changed", err, receipt.String(), posts)
	}
	importedView, err := (workbench.Store{Dir: filepath.Join(destinationDir, "workbench")}).LoadExplorationView(imported.ID)
	if err != nil {
		t.Fatal(err)
	}
	importedView.ID = ""
	if !reflect.DeepEqual(*importedView, view) {
		t.Fatal("fresh destination lost fixed query pair")
	}
	if file := os.Getenv("GOO_QUERY_COUNT_VIEW_OBSERVATIONS"); file != "" {
		raw, _ := json.MarshalIndent(observations, "", "  ")
		if err := os.WriteFile(file, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
