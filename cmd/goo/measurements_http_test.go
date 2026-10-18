package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMeasurementPreviewRegistryAndWorkflowRoutes(t *testing.T) {
	t.Chdir("../..")
	mux := http.NewServeMux()
	registerMeasurements(mux, http.DefaultClient, t.TempDir())
	request := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	for _, path := range []string{"/measurement-packages", "/investigation-workflows"} {
		r := request("GET", path, "")
		if r.Code != 200 {
			t.Fatal(path, r.Code, r.Body.String())
		}
	}
	b, e := os.ReadFile("config/measurements/queue-lag.yaml")
	if e != nil {
		t.Fatal(e)
	}
	payload, _ := json.Marshal(map[string]any{"yaml": string(b), "fromMs": 100000, "toMs": 220000})
	preview := request("POST", "/measurement-packages/plan", string(payload))
	if preview.Code != 200 || !strings.Contains(preview.Body.String(), "otel_metrics_gauge") {
		t.Fatal(preview.Code, preview.Body.String())
	}
	for _, bad := range []string{`{"yaml":"invalid","evidence":{}}`, string(payload) + `{}`, `{"yaml":"invalid"}`} {
		if r := request("POST", "/measurement-packages/plan", bad); r.Code != 400 {
			t.Fatal("bad preview accepted", r.Code)
		}
	}
	payload, _ = json.Marshal(map[string]any{"yaml": string(b)})
	first := request("POST", "/measurement-packages", string(payload))
	second := request("POST", "/measurement-packages", string(payload))
	if first.Code != 201 || first.Body.String() != second.Body.String() {
		t.Fatal("immutable package write", first.Code, second.Code)
	}
	if r := request("POST", "/measurements", `{"definition":{},"rows":[]}`); r.Code != 400 {
		t.Fatal("caller evidence accepted")
	}
}

func TestCustomMeasurementCannotShadowBuiltinWorkflow(t *testing.T) {
	t.Chdir("../..")
	mux := http.NewServeMux()
	registerMeasurements(mux, http.DefaultClient, t.TempDir())
	data, e := os.ReadFile("config/measurements/worker-utilization.yaml")
	if e != nil {
		t.Fatal(e)
	}
	custom := strings.Replace(string(data), "environment: local", "environment: staging", 1)
	payload, _ := json.Marshal(map[string]string{"yaml": custom})
	saved := httptest.NewRecorder()
	mux.ServeHTTP(saved, httptest.NewRequest("POST", "/measurement-packages", strings.NewReader(string(payload))))
	if saved.Code != 201 {
		t.Fatal(saved.Code, saved.Body.String())
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest("GET", "/measurement-packages", nil))
	var items []struct {
		Builtin    bool                               `json:"builtin"`
		Definition struct{ Name, Environment string } `json:"definition"`
	}
	if e = json.Unmarshal(response.Body.Bytes(), &items); e != nil {
		t.Fatal(e)
	}
	builtin, customFound := false, false
	for _, p := range items {
		if p.Definition.Name != "worker-utilization" {
			continue
		}
		if p.Definition.Environment == "local" {
			builtin = p.Builtin
		}
		if p.Definition.Environment == "staging" {
			customFound = !p.Builtin
		}
	}
	if !builtin || !customFound {
		t.Fatal("custom definition changed builtin resolution", items)
	}
}
