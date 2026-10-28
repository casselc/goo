package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func TestContainerReadinessHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	scope := "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	s := metrics.Stream{Resource: map[string]string{"k8s.pod.uid": "pod-a", "container.id": "reported-container", "k8s.container.name": "worker", "deployment.environment.name": "prod"}, Attributes: map[string]string{}, Scope: scope, Version: "0.137.0", Name: "k8s.container.ready"}
	q := measurements.Question{FromMS: 100000, ToMS: 104000, Definition: measurements.Definition{APIVersion: measurements.Version, Name: "container-ready", Title: "Reported container readiness", Kind: "state", StateProfile: metrics.ContainerReadyProfile, Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{}, Resource: &measurements.ResourceBinding{Kind: "kubernetes-container", ID: "reported-container"}, ExactStream: &s, Scope: scope, Version: s.Version, Metric: s.Name}}}
	row, _ := json.Marshal(map[string]any{"resource": s.Resource, "resource_schema_url": "", "scope": scope, "version": s.Version, "schema_url": "", "name": s.Name, "unit": "", "attributes": s.Attributes, "start_ns": "0", "end_ns": "102000000000", "flags": 0, "value": "0"})
	reads := 0
	offline := false
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if offline {
			t.Fatal("offline reopening read telemetry")
		}
		reads++
		sql, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(sql), "FROM otel_metrics_gauge") || strings.Contains(string(sql), "ServiceName") {
			t.Fatal("wrong readiness source", string(sql))
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(row) + "\n")), Header: make(http.Header)}, nil
	})}
	dir := t.TempDir()
	makeMux := func() *http.ServeMux {
		mux := http.NewServeMux()
		registerMeasurements(mux, measurements.Source{Endpoint: "http://source", Client: client}, dir)
		registerCaptures(mux, exploration.Source{Endpoint: "http://source", Client: client}, dir, resultSnapshots)
		return mux
	}
	mux := makeMux()
	observed := []apiObservation{}
	call := func(method, path string, body any, want int) []byte {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(b))))
		if w.Code != want {
			t.Fatal(path, w.Code, w.Body.String())
		}
		pattern := path
		if strings.HasPrefix(path, "/captures/") {
			pattern = "/captures/{id}"
		}
		observed = append(observed, apiObservation{method, pattern, w.Code, b, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	var result measurements.Result
	if json.Unmarshal(call("POST", "/measurements", q, 200), &result) != nil || measurements.ValidateResult(&result) != nil || result.States == nil || result.States.Series[0].Samples[0].Label != "Not ready" || result.State != "observed" {
		t.Fatal("readiness execution failed")
	}
	var page exploration.CapturePage
	if json.Unmarshal(call("POST", "/captures", map[string]any{"title": "Readiness evidence", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}}, 201), &page) != nil || page.Measurement == nil {
		t.Fatal("readiness capture failed")
	}
	saved := &exploration.SavedCapture{CaptureReceipt: page.Receipt, Measurement: page.Measurement}
	comparison, err := exploration.CompareEvidence(saved, saved)
	if err != nil || comparison.State != "unsupported" || comparison.CountDifference != "" || len(comparison.Gauges) != 0 {
		t.Fatal("state arithmetic admitted", comparison, err)
	}
	capturePath := filepath.Join(dir, "exploration", page.Receipt.ID+".json")
	storedBefore, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatal(err)
	}
	frozen, _ := json.Marshal(page.Measurement)
	if reads != 2 {
		t.Fatal("unexpected source count", reads)
	}
	offline = true
	mux = makeMux()
	var reopened exploration.CapturePage
	if json.Unmarshal(call("GET", "/captures/"+page.Receipt.ID, nil, 200), &reopened) != nil {
		t.Fatal("reopen failed")
	}
	again, _ := json.Marshal(reopened.Measurement)
	if string(frozen) != string(again) || reads != 2 {
		t.Fatal("frozen readiness changed")
	}
	storedAfter, err := os.ReadFile(capturePath)
	if err != nil || string(storedBefore) != string(storedAfter) {
		t.Fatal("stored capture bytes changed", err)
	}
	q.Definition.StateProfile = "future"
	call("POST", "/measurements", q, 400)
	if reads != 2 {
		t.Fatal("invalid profile read source")
	}
	if path := os.Getenv("GOO_CONTAINER_STATE_API_OBSERVATIONS"); path != "" {
		b, _ := json.Marshal(observed)
		if err := os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
