package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func containerCPUHTTPParent() measurements.PodMetricDiscoveryQuestion {
	schema := ""
	return measurements.PodMetricDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000, Resource: map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "pod-one", "k8s.container.name": "worker", "container.id": "containerd://literal-one"}, ResourceSchemaURL: &schema}
}
func containerCPUHTTPResult(t *testing.T, parent measurements.PodMetricDiscoveryQuestion, role, value string) *measurements.Result {
	t.Helper()
	scope, metric := measurements.KubeletStatsScope, "container.cpu.usage"
	if role != "usage" {
		scope = measurements.K8sClusterScope
		metric = "k8s.container.cpu_" + role
	}
	stream := metrics.Stream{Resource: parent.Resource, ResourceSchemaURL: *parent.ResourceSchemaURL, Scope: scope, Version: "0.137.0", Name: metric, Unit: "{cpu}", Attributes: map[string]string{}}
	def := measurements.Definition{APIVersion: measurements.Version, Name: "container-" + role, Title: role, Kind: "gauge", Environment: parent.Environment, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "kubernetes-container", ID: parent.Resource["container.id"]}, Services: []string{}, Scope: scope, Version: "0.137.0", Metric: metric, Unit: "{cpu}", ExactStream: &stream}}
	q := measurements.Question{Definition: def, FromMS: parent.FromMS, ToMS: parent.ToMS}
	raw, _ := json.Marshal(map[string]any{"resource": stream.Resource, "resource_schema_url": stream.ResourceSchemaURL, "scope": scope, "version": "0.137.0", "schema_url": "", "name": metric, "unit": "{cpu}", "attributes": map[string]string{}, "end_ns": "120000000000", "flags": 0, "value": value})
	r, e := measurements.Evaluate(q, []json.RawMessage{raw}, time.Unix(200, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func containerCPUHTTPRequest(t *testing.T) measurements.ContainerCPUViewRequest {
	p := containerCPUHTTPParent()
	return measurements.ContainerCPUViewRequest{Parent: p, Usage: containerCPUHTTPResult(t, p, "usage", "0.125"), Request: containerCPUHTTPResult(t, p, "request", "0.25"), Limit: containerCPUHTTPResult(t, p, "limit", "0.5")}
}
func containerCPUHTTPBody(t *testing.T, edit func(*measurements.ContainerCPUViewRequest)) string {
	r := containerCPUHTTPRequest(t)
	if edit != nil {
		edit(&r)
	}
	b, _ := json.Marshal(r)
	return string(b)
}
func containerCPUWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	registerContainerCPUView(mux)
	body := c.Body + c.Suffix
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	return w
}
func TestContainerCPUHTTPFrozenContracts(t *testing.T) {
	full := containerCPUHTTPBody(t, nil)
	cases := []planningWireCase{
		{Name: "aligned-three-signals", Body: full, Status: 200},
		{Name: "all-inputs-missing", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) { r.Usage = nil; r.Request = nil; r.Limit = nil }), Status: 200},
		{Name: "explicit-zero-denominator", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) {
			r.Limit = containerCPUHTTPResult(t, r.Parent, "limit", "0")
		}), Status: 200},
		{Name: "independent-resource-schema", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) {
			p := r.Parent
			s := "independent"
			p.ResourceSchemaURL = &s
			r.Request = containerCPUHTTPResult(t, p, "request", "0.25")
		}), Status: 200},
		{Name: "reported-decimal-retained-before-float-rounding", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) {
			r.Usage = containerCPUHTTPResult(t, r.Parent, "usage", "1.00000000000000001")
		}), Status: 200},
		{Name: "wrong-valid-container-context", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) {
			p := containerCPUHTTPParent()
			p.Resource["container.id"] = "replacement"
			r.Limit = containerCPUHTTPResult(t, p, "limit", "0.5")
		}), Status: 400},
		{Name: "forged-derived-value", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) { r.Limit.Gauges.Series[0].Samples[0].Value = nil }), Status: 400},
		{Name: "parent-context-too-long", Body: containerCPUHTTPBody(t, func(r *measurements.ContainerCPUViewRequest) {
			r.Parent.Resource["container.id"] = strings.Repeat("x", 513)
		}), Status: 400},
		{Name: "unknown-field", Body: `{"unexpected":true}`, Status: 400},
		{Name: "second-document", Body: full, Suffix: " {}", Status: 400},
		{Name: "oversize", Body: full, PadTo: (8 << 20) + 1, Status: 400},
		{Name: "method", Method: "GET", Status: 405},
	}
	for n := range cases {
		cases[n].Path = "/container-cpu/view"
		if cases[n].Method == "" {
			cases[n].Method = "POST"
		}
	}
	const file = "../../testdata/http/container-cpu.json"
	if os.Getenv("GOO_UPDATE_CONTAINER_CPU_HTTP_FIXTURES") == "1" {
		for n := range cases {
			w := containerCPUWireResponse(t, cases[n])
			if w.Code != cases[n].Status {
				t.Fatal(cases[n].Name, w.Code, w.Body.String())
			}
			cases[n].Response = w.Body.String()
			cases[n].Headers = map[string]string{}
			for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[n].Headers[key] = w.Header().Get(key)
			}
		}
		b, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e := os.WriteFile(file, append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	fixtures := dashboardProfileFixture(t, file, len(cases))
	for _, c := range fixtures.Cases {
		w := containerCPUWireResponse(t, c)
		if w.Code != c.Status || w.Body.String() != c.Response {
			t.Fatal(c.Name, w.Code, w.Body.String())
		}
		for key, value := range c.Headers {
			if w.Header().Get(key) != value {
				t.Fatal("wireheader changed", key)
			}
		}
	}
}
func TestContainerCPUPureHTTPAndOfflineNativeCapture(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	dir := t.TempDir()
	calls := 0
	held := containerCPUHTTPRequest(t)
	raw := held.Usage.Rows[0]
	source := measurements.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(raw) + "\n"))}, nil
	})}}
	mux := http.NewServeMux()
	registerMeasurements(mux, source, dir)
	body := containerCPUHTTPBody(t, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/container-cpu/view", strings.NewReader(body)))
	if w.Code != 200 || calls != 0 {
		t.Fatal(w.Code, w.Body.String(), calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/container-cpu/view", strings.NewReader(body)).WithContext(ctx))
	if w.Code != 400 || calls != 0 {
		t.Fatal("pure cancellation", w.Code, calls)
	}
	// Capture retains one native signal through the existing fresh-read path;
	// the composer itself neither publishes nor rereads evidence.
	q := held.Usage.Question
	request, _ := json.Marshal(map[string]any{"title": "Container CPU usage", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}})
	captures := http.NewServeMux()
	registerCaptures(captures, exploration.Source{Endpoint: source.Endpoint, Client: source.Client}, dir, resultSnapshots)
	w = httptest.NewRecorder()
	captures.ServeHTTP(w, httptest.NewRequest("POST", "/captures", strings.NewReader(string(request))))
	if w.Code != 201 || calls != 1 {
		t.Fatal("fresh individual capture failed", w.Code, w.Body.String(), calls)
	}
	var page exploration.CapturePage
	if e := json.Unmarshal(w.Body.Bytes(), &page); e != nil || page.Measurement == nil || measurements.ValidateResult(page.Measurement) != nil {
		t.Fatal("invalid captured container measurement", e)
	}
	original, _ := json.Marshal(page.Measurement)
	offline := http.NewServeMux()
	registerCaptures(offline, exploration.Source{Endpoint: "http://offline", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		t.Fatal("offline reopen contacted source")
		return nil, nil
	})}}, dir, resultSnapshots)
	w = httptest.NewRecorder()
	offline.ServeHTTP(w, httptest.NewRequest("GET", "/captures/"+page.Receipt.ID, nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var reopened exploration.CapturePage
	if e := json.Unmarshal(w.Body.Bytes(), &reopened); e != nil || reopened.Measurement == nil || measurements.ValidateResult(reopened.Measurement) != nil {
		t.Fatal("offline replay failed", e)
	}
	again, _ := json.Marshal(reopened.Measurement)
	if string(original) != string(again) || calls != 1 {
		t.Fatal("offline evidence bytes changed or source reread")
	}
}
