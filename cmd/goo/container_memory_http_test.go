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

func containerMemoryHTTPParent() measurements.PodMetricDiscoveryQuestion {
	schema := ""
	return measurements.PodMetricDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000, Resource: map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "pod-one", "k8s.container.name": "worker", "container.id": "containerd://literal-one"}, ResourceSchemaURL: &schema}
}
func containerMemoryHTTPResult(t *testing.T, parent measurements.PodMetricDiscoveryQuestion, role, value string) *measurements.Result {
	t.Helper()
	scope, metric := measurements.KubeletStatsScope, "container.memory.working_set"
	if role != "working-set" {
		scope = measurements.K8sClusterScope
		metric = "k8s.container.memory_" + role
	}
	stream := metrics.Stream{Resource: parent.Resource, ResourceSchemaURL: *parent.ResourceSchemaURL, Scope: scope, Version: "0.137.0", Name: metric, Unit: "By", Attributes: map[string]string{}}
	def := measurements.Definition{APIVersion: measurements.Version, Name: "container-" + role, Title: role, Kind: "gauge", Environment: parent.Environment, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "kubernetes-container", ID: parent.Resource["container.id"]}, Services: []string{}, Scope: scope, Version: "0.137.0", Metric: metric, Unit: "By", ExactStream: &stream}}
	q := measurements.Question{Definition: def, FromMS: parent.FromMS, ToMS: parent.ToMS}
	raw, _ := json.Marshal(map[string]any{"resource": stream.Resource, "resource_schema_url": stream.ResourceSchemaURL, "scope": scope, "version": "0.137.0", "schema_url": "", "name": metric, "unit": "By", "attributes": map[string]string{}, "end_ns": "120000000000", "flags": 0, "value": value})
	r, e := measurements.Evaluate(q, []json.RawMessage{raw}, time.Unix(200, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func containerMemoryHTTPRequest(t *testing.T) measurements.ContainerMemoryViewRequest {
	p := containerMemoryHTTPParent()
	return measurements.ContainerMemoryViewRequest{Parent: p, WorkingSet: containerMemoryHTTPResult(t, p, "working-set", "50"), Request: containerMemoryHTTPResult(t, p, "request", "100"), Limit: containerMemoryHTTPResult(t, p, "limit", "200")}
}
func containerMemoryHTTPBody(t *testing.T, edit func(*measurements.ContainerMemoryViewRequest)) string {
	r := containerMemoryHTTPRequest(t)
	if edit != nil {
		edit(&r)
	}
	b, _ := json.Marshal(r)
	return string(b)
}
func containerMemoryWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	registerContainerMemoryView(mux)
	body := c.Body + c.Suffix
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	return w
}
func TestContainerMemoryHTTPFrozenContracts(t *testing.T) {
	full := containerMemoryHTTPBody(t, nil)
	cases := []planningWireCase{
		{Name: "aligned-three-signals", Body: full, Status: 200},
		{Name: "all-inputs-missing", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) { r.WorkingSet = nil; r.Request = nil; r.Limit = nil }), Status: 200},
		{Name: "explicit-zero-denominator", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) {
			r.Limit = containerMemoryHTTPResult(t, r.Parent, "limit", "0")
		}), Status: 200},
		{Name: "independent-resource-schema", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) {
			p := r.Parent
			s := "independent"
			p.ResourceSchemaURL = &s
			r.Request = containerMemoryHTTPResult(t, p, "request", "100")
		}), Status: 200},
		{Name: "invalid-raw-fraction-rounded-by-float", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) {
			r.WorkingSet = containerMemoryHTTPResult(t, r.Parent, "working-set", "1.00000000000000001")
		}), Status: 200},
		{Name: "wrong-valid-container-context", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) {
			p := containerMemoryHTTPParent()
			p.Resource["container.id"] = "replacement"
			r.Limit = containerMemoryHTTPResult(t, p, "limit", "200")
		}), Status: 400},
		{Name: "forged-derived-value", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) { r.Limit.Gauges.Series[0].Samples[0].Value = nil }), Status: 400},
		{Name: "parent-context-too-long", Body: containerMemoryHTTPBody(t, func(r *measurements.ContainerMemoryViewRequest) {
			r.Parent.Resource["container.id"] = strings.Repeat("x", 513)
		}), Status: 400},
		{Name: "unknown-field", Body: `{"unexpected":true}`, Status: 400},
		{Name: "second-document", Body: full, Suffix: " {}", Status: 400},
		{Name: "oversize", Body: full, PadTo: (8 << 20) + 1, Status: 400},
		{Name: "method", Method: "GET", Status: 405},
	}
	for n := range cases {
		cases[n].Path = "/container-memory/view"
		if cases[n].Method == "" {
			cases[n].Method = "POST"
		}
	}
	const file = "../../testdata/http/container-memory.json"
	if os.Getenv("GOO_UPDATE_CONTAINER_MEMORY_HTTP_FIXTURES") == "1" {
		for n := range cases {
			w := containerMemoryWireResponse(t, cases[n])
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
		w := containerMemoryWireResponse(t, c)
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
func TestContainerMemoryPureHTTPAndOfflineNativeCapture(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	dir := t.TempDir()
	calls := 0
	held := containerMemoryHTTPRequest(t)
	raw := held.WorkingSet.Rows[0]
	source := measurements.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(raw) + "\n"))}, nil
	})}}
	mux := http.NewServeMux()
	registerMeasurements(mux, source, dir)
	body := containerMemoryHTTPBody(t, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/container-memory/view", strings.NewReader(body)))
	if w.Code != 200 || calls != 0 {
		t.Fatal(w.Code, w.Body.String(), calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/container-memory/view", strings.NewReader(body)).WithContext(ctx))
	if w.Code != 400 || calls != 0 {
		t.Fatal("pure cancellation", w.Code, calls)
	}
	// Capture retains one native signal through the existing fresh-read path;
	// the composer itself neither publishes nor rereads evidence.
	q := held.WorkingSet.Question
	request, _ := json.Marshal(map[string]any{"title": "Container working set", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}})
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
