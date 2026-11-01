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

func TestFacetViewSavedPortableAndCLIHTTP310(t *testing.T) {
	raw, err := os.ReadFile("../../config/views/service-log-facet.yaml")
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := workbench.DecodeViewArtifact(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	observations := []apiObservation{}
	post := func(path string, input any, want int) json.RawMessage {
		t.Helper()
		body, _ := json.Marshal(input)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, bytes.NewReader(body)))
		if w.Code != want {
			t.Fatal(path, w.Code, w.Body.String())
		}
		if want < 300 {
			observations = append(observations, apiObservation{Method: "POST", Path: path, Status: want, Request: body, Response: append(json.RawMessage(nil), w.Body.Bytes()...)})
		}
		return w.Body.Bytes()
	}
	savedRaw := post("/exploration-views", artifact.Artifact.View, 201)
	var saved workbench.ExplorationView
	if err = json.Unmarshal(savedRaw, &saved); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/exploration-views/"+saved.ID, nil))
	var loaded workbench.ExplorationView
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err = json.Unmarshal(w.Body.Bytes(), &loaded); err != nil {
		t.Fatal(err)
	}
	loaded.ID = ""
	if !reflect.DeepEqual(loaded, artifact.Artifact.View) {
		t.Fatal("saved facet changed configured root")
	}
	encodedRaw := post("/view-artifacts/encode", workbench.ViewArtifactEncodeRequest{Artifact: artifact.Artifact}, 200)
	var encoded workbench.ViewArtifactResult
	if err = json.Unmarshal(encodedRaw, &encoded); err != nil {
		t.Fatal(err)
	}
	post("/view-artifacts/decode", workbench.ViewArtifactDecodeRequest{YAML: encoded.YAML}, 200)
	var invalid map[string]any
	json.Unmarshal(savedRaw, &invalid)
	delete(invalid, "id")
	child := invalid["composition"].(map[string]any)["children"].([]any)[0].(map[string]any)
	child["facet"].(map[string]any)["values"] = []any{"api"}
	post("/exploration-views", invalid, 400)
	page, err := (workbench.Store{Dir: filepath.Join(dir, "workbench")}).ExplorationViews("")
	if err != nil || len(page.Items) != 1 {
		t.Fatal("invalid facet was published", err)
	}
	if reads != 0 {
		t.Fatal("facet save/codec attempted a source read", reads)
	}
	http.DefaultTransport = previous
	destinationDir := t.TempDir()
	destination := http.NewServeMux()
	registerWorkbench(destination, destinationDir, retentionConnection{}, &exploration.ResultSnapshots{}, nil)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != "POST" || r.URL.Path != "/api/exploration-views" {
			t.Error("facet CLI performed more than explicit import")
		}
		http.StripPrefix("/api", destination).ServeHTTP(w, r)
	}))
	defer server.Close()
	file := filepath.Join(t.TempDir(), "facet.yaml")
	if err = os.WriteFile(file, []byte(encoded.YAML), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err = run([]string{"validate-view", "-file", file}, &output); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err = run([]string{"import-view", "-file", file, "-url", server.URL}, &output); err != nil {
		t.Fatal(err)
	}
	var receipt viewArtifactCLIReceipt
	if err = json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.ID != saved.ID || requests != 1 {
		t.Fatal("facet CLI changed identity or performed extra requests", err, requests, output.String())
	}
	imported, err := (workbench.Store{Dir: filepath.Join(destinationDir, "workbench")}).LoadExplorationView(receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	imported.ID = ""
	if !reflect.DeepEqual(*imported, loaded) {
		t.Fatal("facet portable import changed root or declaration")
	}
	if file := os.Getenv("GOO_FACET_VIEW_OBSERVATIONS"); file != "" {
		raw, _ := json.MarshalIndent(observations, "", "  ")
		if err := os.WriteFile(file, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
