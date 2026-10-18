package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/exploration"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTelemetryPackageAuthoringHTTPAndGeneratedGoManifest(t *testing.T) {
	mux := http.NewServeMux()
	dir := t.TempDir()
	registerTelemetryPackages(mux, dir)
	call := func(method, path string, data any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(data)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(b)))
		return w
	}
	input := map[string]any{"question": exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 10, Offset: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "name"}}}, "name": "python-jobs", "title": "Python jobs", "lookbackMinutes": 5}
	w := call("POST", "/telemetry-packages/draft", input)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var plan exploration.PackagePlan
	json.Unmarshal(w.Body.Bytes(), &plan)
	if plan.Question.Offset != 0 || plan.Question.Limit != 100 {
		t.Fatal("page retained")
	}
	w = call("POST", "/telemetry-packages/plan", map[string]any{"yaml": plan.YAML, "fromMs": 1000, "toMs": 61000, "annotation": map[string]string{"name": "index.refresh", "role": "consumer"}})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &plan)
	source := t.TempDir()
	os.WriteFile(filepath.Join(source, "main.go"), []byte(plan.Handoff.GoExample), 0600)
	manifest, e := annotations.ScanDir(source)
	if e != nil || len(manifest.Declarations) != 1 || manifest.Declarations[0].Name != "index.refresh" {
		t.Fatal("generated annotation is not accepted", e)
	}
	w = call("POST", "/telemetry-packages", map[string]string{"yaml": plan.YAML})
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var saved exploration.SavedPackage
	json.Unmarshal(w.Body.Bytes(), &saved)
	if w = call("GET", "/telemetry-packages/"+saved.Revision, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "Python jobs") {
		t.Fatal(w.Body.String())
	}
	if w = call("GET", "/telemetry-packages", nil); w.Code != 200 || !strings.Contains(w.Body.String(), saved.Revision) {
		t.Fatal(w.Body.String())
	}
	for _, data := range []any{map[string]any{"yaml": plan.YAML, "definition": plan.Definition}, map[string]any{"yaml": plan.YAML, "sql": "select 1"}, map[string]any{"yaml": plan.YAML + "unknown: true\n"}, map[string]any{"yaml": plan.YAML, "fromMs": 123}} {
		if w = call("POST", "/telemetry-packages/plan", data); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
