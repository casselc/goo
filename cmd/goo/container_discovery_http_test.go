package main

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestContainerReadinessDiscoveryHTTPWorkflow(t *testing.T) {
	scope := "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	req := podMetricQuestion()
	req.Resource["container.id"] = "containerd://reported-literal"
	req.Resource["k8s.container.name"] = "worker"
	metadata, _ := json.Marshal(map[string]any{"kind": "gauge", "scope": scope, "version": "0.137.0", "schemaURL": "", "metric": "k8s.container.ready", "unit": "", "temporality": 0, "monotonic": false, "resource": req.Resource, "resourceSchemaURL": "", "attributes": map[string]string{}})
	row, _ := json.Marshal(map[string]any{"resource": req.Resource, "resource_schema_url": "", "scope": scope, "version": "0.137.0", "schema_url": "", "name": "k8s.container.ready", "unit": "", "attributes": map[string]string{}, "end_ns": "102000000000", "flags": 0, "value": "0"})
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(r.Body)
		payload := row
		if strings.Contains(string(sql), " UNION ALL ") {
			payload = metadata
		}
		if strings.Contains(string(sql), "ServiceName") || !strings.Contains(string(sql), "ResourceSchemaUrl=") {
			t.Fatal("readiness identity broadened", string(sql))
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(payload) + "\n")), Header: make(http.Header)}, nil
	})}
	mux := http.NewServeMux()
	source := measurements.Source{Endpoint: "http://source", Client: client}
	registerMeasurements(mux, source, t.TempDir())
	observed := []apiObservation{}
	post := func(path string, body any) []byte {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(string(b))))
		if w.Code != 200 {
			t.Fatal(path, w.Code, w.Body.String())
		}
		observed = append(observed, apiObservation{"POST", path, w.Code, b, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	var discovery measurements.PodMetricDiscovery
	if json.Unmarshal(post("/pod-metric-discovery", req), &discovery) != nil || reads != 1 {
		t.Fatal("readiness discovery reads", reads)
	}
	var plan measurements.PodDashboardPlan
	if json.Unmarshal(post("/pod-dashboard/plan", measurements.PodDashboardRequest{Discovery: discovery}), &plan) != nil || len(plan.Widgets) != 1 || plan.Widgets[0].Question.Definition.Kind != "state" || reads != 1 {
		t.Fatal("pure readiness planning", reads)
	}
	var result measurements.Result
	if json.Unmarshal(post("/measurements", plan.Widgets[0].Question), &result) != nil || measurements.ValidateResult(&result) != nil || result.States == nil || result.States.Series[0].Samples[0].Label != "Not ready" || result.State != "observed" || reads != 2 {
		t.Fatal("discovered readiness result", reads)
	}
	if path := os.Getenv("GOO_CONTAINER_DISCOVERY_API_OBSERVATIONS"); path != "" {
		b, _ := json.Marshal(observed)
		if err := os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
