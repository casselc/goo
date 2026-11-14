package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
	"github.com/casselc/goo/workbench"
)

func celAuthoringHTTPView(t *testing.T) workbench.ExplorationView {
	t.Helper()
	base := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &exploration.Predicate{Any: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "eq", Value: "edge"},
		{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"},
	}}}
	request, err := querycel.NewPlanRequest(base, "// preserve my explanation\nlog.severity_number >= 17u", "condition", "mount-1", "9007199254740993")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := querycel.PlanRecord(request)
	if err != nil {
		t.Fatal(err)
	}
	authoring, err := workbench.NewRecordAuthoring(request, "parent-7")
	if err != nil {
		t.Fatal(err)
	}
	return workbench.ExplorationView{Version: 1, Title: "Checked CEL logs", Queries: []exploration.Query{plan.Question}, Layout: "records", Columns: []string{"utc", "summary"}, Authoring: map[string]workbench.RecordAuthoring{"logs": authoring}}
}

func TestCELAuthoringHTTPPreserveAndRejectBeforePublication(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	call := func(method, path string, value any) *httptest.ResponseRecorder {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		return w
	}
	view := celAuthoringHTTPView(t)
	artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "platform", Name: "checked-cel", Revision: "v1", View: view}
	encoded := call("POST", "/view-artifacts/encode", workbench.ViewArtifactEncodeRequest{Artifact: artifact})
	if encoded.Code != 200 {
		t.Fatal(encoded.Code, encoded.Body.String())
	}
	var portable workbench.ViewArtifactResult
	if err := json.Unmarshal(encoded.Body.Bytes(), &portable); err != nil {
		t.Fatal(err)
	}
	decoded := call("POST", "/view-artifacts/decode", workbench.ViewArtifactDecodeRequest{YAML: portable.YAML})
	if decoded.Code != 200 {
		t.Fatal(decoded.Code, decoded.Body.String())
	}
	var reopened workbench.ViewArtifactResult
	if err := json.Unmarshal(decoded.Body.Bytes(), &reopened); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reopened.Artifact.View, view) {
		t.Fatal("portable codec changed checked source or typed meaning")
	}
	if _, err := os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(err) {
		t.Fatal("codec wrote to store", err)
	}
	// Every mutation is rejected by both the ordinary save and portable codecs,
	// before the first explicit publication creates any workbench directory.
	for _, tc := range []struct {
		name   string
		mutate func(*workbench.ExplorationView)
	}{
		{"source", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Source = "log.body == \"other\""
			v.Authoring["logs"] = a
		}},
		{"old-library", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Binding.Library = "goo.cel.otel/old"
			v.Authoring["logs"] = a
		}},
		{"old-compiler", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Binding.Compiler = "goo.record-cel/old"
			v.Authoring["logs"] = a
		}},
		{"digest", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.LoweringDigest = strings.Repeat("0", 64)
			v.Authoring["logs"] = a
		}},
		{"base", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Base.Environment = "other"
			v.Authoring["logs"] = a
		}},
		{"typed-question", func(v *workbench.ExplorationView) { v.Queries[0].Limit = 20 }},
		{"wrong-signal-key", func(v *workbench.ExplorationView) {
			v.Authoring = map[string]workbench.RecordAuthoring{"spans": v.Authoring["logs"]}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(view)
			var bad workbench.ExplorationView
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&bad)
			w := call("POST", "/exploration-views", bad)
			if w.Code != 400 {
				t.Fatal("tampered save accepted", w.Code, w.Body.String())
			}
			a := artifact
			a.View = bad
			w = call("POST", "/view-artifacts/encode", workbench.ViewArtifactEncodeRequest{Artifact: a})
			if w.Code != 400 {
				t.Fatal("tampered export accepted", w.Code, w.Body.String())
			}
			if _, err := os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(err) {
				t.Fatal("refusal wrote to store", err)
			}
		})
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { v["authoring"] = nil },
		func(v map[string]any) { v["authoring"] = map[string]any{} },
		func(v map[string]any) { v["authoring"].(map[string]any)["logs"] = nil },
		func(v map[string]any) { v["authoring"].(map[string]any)["logs"].(map[string]any)["extra"] = true },
		func(v map[string]any) {
			v["authoring"].(map[string]any)["logs"].(map[string]any)["binding"].(map[string]any)["Library"] = querycel.LibraryVersion
		},
	} {
		raw, _ := json.Marshal(view)
		var wire map[string]any
		if err := json.Unmarshal(raw, &wire); err != nil {
			t.Fatal(err)
		}
		mutate(wire)
		if w := call("POST", "/exploration-views", wire); w.Code != 400 {
			t.Fatal("malformed provenance carrier accepted", w.Code, w.Body.String())
		}
		if _, err := os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(err) {
			t.Fatal("malformed carrier wrote to store", err)
		}
	}
	oldYAML := strings.Replace(portable.YAML, querycel.LibraryVersion, "goo.cel.otel/old", 1)
	if oldYAML == portable.YAML {
		t.Fatal("portable artifact omitted compiler library")
	}
	if w := call("POST", "/view-artifacts/decode", workbench.ViewArtifactDecodeRequest{YAML: oldYAML}); w.Code != 400 {
		t.Fatal("old-library import silently migrated", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(err) {
		t.Fatal("refused import wrote to store", err)
	}
	saved := call("POST", "/exploration-views", reopened.Artifact.View)
	if saved.Code != 201 {
		t.Fatal(saved.Code, saved.Body.String())
	}
	var stored workbench.ExplorationView
	if err := json.Unmarshal(saved.Body.Bytes(), &stored); err != nil {
		t.Fatal(err)
	}
	loaded := call("GET", "/exploration-views/"+stored.ID, nil)
	if loaded.Code != 200 {
		t.Fatal(loaded.Code, loaded.Body.String())
	}
	var actual workbench.ExplorationView
	if err := json.Unmarshal(loaded.Body.Bytes(), &actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, stored) {
		t.Fatal("saved authored definition changed on reopen")
	}
	// A second isolated store imports the same typed portable artifact; source text
	// and revision tokens survive unchanged, without acquiring telemetry.
	destination := http.NewServeMux()
	registerWorkbench(destination, t.TempDir(), retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	raw, _ := json.Marshal(reopened.Artifact.View)
	w := httptest.NewRecorder()
	destination.ServeHTTP(w, httptest.NewRequest("POST", "/exploration-views", strings.NewReader(string(raw))))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var imported workbench.ExplorationView
	if err := json.Unmarshal(w.Body.Bytes(), &imported); err != nil || !reflect.DeepEqual(imported, stored) {
		t.Fatal("import changed source or content identity", err)
	}
	if reads != 0 {
		t.Fatal("configuration routes acquired telemetry", reads)
	}
}
