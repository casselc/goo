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
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
)

func TestPreviousCountViewHTTPAdmission321(t *testing.T) {
	selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
	q := exploration.Query{Signal: "logs", EnvironmentSelector: selector, FromMS: 61000, ToMS: 121000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}
	baseline := q
	baseline.FromMS = 1000
	baseline.ToMS = 61000
	scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: "logs"}
	view := workbench.ExplorationView{Version: 1, Title: "Previous count window", Queries: []exploration.Query{q}, Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
		{Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.ViewKind{Name: "bar", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count"}}},
		{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: "record-table", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}},
		{Comparison: &workbench.CountComparisonViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.CountComparisonViewKind{Name: "query-count-comparison", Version: 1}, BaselineQuery: &baseline, Display: workbench.CountComparisonViewDisplay{Value: "count", View: "groups"}}},
	}}}
	reads := 0
	transport := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = transport })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	var observations []apiObservation
	call := func(path string, value any, want int) []byte {
		t.Helper()
		raw, _ := json.Marshal(value)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest("POST", path, bytes.NewReader(raw)))
		if response.Code != want {
			t.Fatal(path, response.Code, response.Body.String())
		}
		observations = append(observations, apiObservation{Method: "POST", Path: path, Status: want, Request: raw, Response: append(json.RawMessage(nil), response.Body.Bytes()...)})
		return response.Body.Bytes()
	}
	raw := call("/exploration-views", view, 201)
	var saved workbench.ExplorationView
	if json.Unmarshal(raw, &saved) != nil {
		t.Fatal("invalid saved response")
	}
	loaded, err := (workbench.Store{Dir: filepath.Join(dir, "workbench")}).LoadExplorationView(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	loaded.ID = ""
	if !reflect.DeepEqual(*loaded, view) {
		t.Fatal("adjacent selectors/windows changed on reopening")
	}
	artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "platform", Name: "previous-counts", Revision: "v1", View: *loaded}
	encodedRaw := call("/view-artifacts/encode", workbench.ViewArtifactEncodeRequest{Artifact: artifact}, 200)
	var encoded workbench.ViewArtifactResult
	if json.Unmarshal(encodedRaw, &encoded) != nil {
		t.Fatal("invalid codec response")
	}
	decodedRaw := call("/view-artifacts/decode", workbench.ViewArtifactDecodeRequest{YAML: encoded.YAML}, 200)
	var decoded workbench.ViewArtifactResult
	if json.Unmarshal(decodedRaw, &decoded) != nil || !reflect.DeepEqual(decoded.Artifact, artifact) {
		t.Fatal("portable previous pair lost context")
	}
	for _, change := range []func(*exploration.Query){func(b *exploration.Query) { b.FromMS = 0; b.ToMS = 60000 }, func(b *exploration.Query) { b.FromMS = 2000; b.ToMS = 62000 }, func(b *exploration.Query) { b.FromMS = 1000; b.ToMS = 60000 }, func(b *exploration.Query) { b.FromMS = 121000; b.ToMS = 181000 }, func(b *exploration.Query) { b.Offset = 1 }, func(b *exploration.Query) {
		b.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "other"}}
	}} {
		var copy workbench.ExplorationView
		wire, _ := json.Marshal(view)
		json.Unmarshal(wire, &copy)
		copy.Composition.Children[2].Hidden = true
		change(copy.Composition.Children[2].Comparison.BaselineQuery)
		call("/exploration-views", copy, 400)
		invalid := artifact
		invalid.View = copy
		call("/view-artifacts/encode", workbench.ViewArtifactEncodeRequest{Artifact: invalid}, 400)
	}
	if reads != 0 {
		t.Fatal("configuration validation/save/codec performed source IO", reads)
	}
	if path := os.Getenv("GOO_PREVIOUS_COUNT_VIEW_OBSERVATIONS"); path != "" {
		raw, _ := json.MarshalIndent(observations, "", "  ")
		if err := os.WriteFile(path, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
