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
	"github.com/casselc/goo/workbench"
)

func artifactHTTPInput() workbench.ViewArtifactEncodeRequest {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
	v := workbench.ExplorationView{Version: 1, Title: "HTTP logs", Queries: []exploration.Query{q}, Layout: "records", Columns: []string{"utc", "summary"}}
	return workbench.ViewArtifactEncodeRequest{Artifact: workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "payments", Name: "http-logs", Revision: "v1", View: v}}
}
func TestViewArtifactHTTPWireFixtures(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))
		return w
	}
	encode, _ := json.Marshal(artifactHTTPInput())
	out, err := workbench.EncodeViewArtifact(artifactHTTPInput().Artifact)
	if err != nil {
		t.Fatal(err)
	}
	decode, _ := json.Marshal(workbench.ViewArtifactDecodeRequest{YAML: out.YAML})
	alias, _ := json.Marshal(workbench.ViewArtifactDecodeRequest{YAML: "a: &a hello\nb: *a\n"})
	duplicate, _ := json.Marshal(workbench.ViewArtifactDecodeRequest{YAML: out.YAML + "name: duplicate\n"})
	oversize, _ := json.Marshal(workbench.ViewArtifactDecodeRequest{YAML: strings.Repeat(" ", workbench.ViewArtifactBytes+1)})
	cases := []planningWireCase{
		{Name: "encode", Method: "POST", Path: "/view-artifacts/encode", Body: string(encode), Status: 200},
		{Name: "decode", Method: "POST", Path: "/view-artifacts/decode", Body: string(decode), Status: 200},
		{Name: "unknown", Method: "POST", Path: "/view-artifacts/encode", Body: `{"artifact":{},"unknown":1}`, Status: 400},
		{Name: "null", Method: "POST", Path: "/view-artifacts/encode", Body: `{"artifact":null}`, Status: 400},
		{Name: "aliases", Method: "POST", Path: "/view-artifacts/decode", Body: string(alias), Status: 400},
		{Name: "duplicate", Method: "POST", Path: "/view-artifacts/decode", Body: string(duplicate), Status: 400},
		{Name: "trailing", Method: "POST", Path: "/view-artifacts/decode", Body: string(decode) + ` {}`, Status: 400},
	}
	for _, body := range []string{string(oversize), string(decode) + strings.Repeat(" ", workbench.MaxBytes)} {
		if w := post("/view-artifacts/decode", body); w.Code != 400 {
			t.Fatal("artifact size budget bypassed", w.Code)
		}
	}
	for i := range cases {
		w := post(cases[i].Path, cases[i].Body)
		if w.Code != cases[i].Status {
			t.Fatal(cases[i].Name, w.Code, w.Body.String())
		}
		cases[i].Response = w.Body.String()
		cases[i].Headers = map[string]string{"Content-Type": w.Header().Get("Content-Type"), "Cache-Control": w.Header().Get("Cache-Control")}
	}
	const filename = "../../testdata/http/view-artifacts.json"
	if os.Getenv("GOO_UPDATE_VIEW_ARTIFACT_HTTP_FIXTURES") == "1" {
		raw, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if err := os.WriteFile(filename, append(raw, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var frozen planningWireFixtures
	if err = json.Unmarshal(raw, &frozen); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(frozen.Cases, cases) {
		t.Fatal("artifact HTTP wire fixtures changed")
	}
	if reads != 0 {
		t.Fatal("codec acquired telemetry")
	}
	if _, err = os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(err) {
		t.Fatal("codec published configuration", err)
	}
	// Import is an explicit ordinary save, and works in a fresh destination store.
	var decoded workbench.ViewArtifactResult
	if err = json.Unmarshal([]byte(cases[1].Response), &decoded); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(decoded.Artifact.View)
	w := post("/exploration-views", string(body))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var saved workbench.ExplorationView
	json.Unmarshal(w.Body.Bytes(), &saved)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/exploration-views/"+saved.ID, nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var loaded workbench.ExplorationView
	json.Unmarshal(w.Body.Bytes(), &loaded)
	loaded.ID = ""
	if !reflect.DeepEqual(loaded, decoded.Artifact.View) {
		t.Fatal("import changed fixed view")
	}
}
