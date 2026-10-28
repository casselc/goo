package main

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/measurements"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func podMetricQuestion() measurements.PodMetricDiscoveryQuestion {
	schema := ""
	return measurements.PodMetricDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 104000, Resource: map[string]string{"k8s.pod.uid": "pod-1", "deployment.environment.name": "prod"}, ResourceSchemaURL: &schema}
}

var podMetricHTTPBody = func() string { raw, _ := json.Marshal(podMetricQuestion()); return string(raw) }()

func podMetricHTTPMetadata() string {
	rows := []string{}
	resource := podMetricQuestion().Resource
	for _, metric := range []string{"k8s.pod.memory.usage", "k8s.pod.memory.working_set"} {
		raw, _ := json.Marshal(map[string]any{"kind": "gauge", "scope": "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver", "version": "0.137.0", "schemaURL": "", "metric": metric, "unit": "By", "temporality": 0, "monotonic": false, "resource": resource, "resourceSchemaURL": "", "attributes": map[string]string{}})
		rows = append(rows, string(raw))
	}
	return strings.Join(rows, "\n") + "\n"
}
func podMetricHTTPSource(payload string, reads *int) measurements.Source {
	return measurements.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		*reads++

		sql, _ := io.ReadAll(r.Body)
		if strings.Contains(string(sql), "ServiceName") || strings.Contains(string(sql), "mapUpdate") || !strings.Contains(string(sql), "JSONExtractKeysAndValues") || r.URL.Query().Get("param_resourceSchema") != "" {
			panic("pod discovery identity SQL mismatch")
		}
		if path := os.Getenv("GOO_POD_METRIC_SQL"); path != "" {
			if err := os.WriteFile(path, sql, 0600); err != nil {
				panic(err)
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}}
}
func podMetricHTTPPlanBody(t *testing.T, change func(*measurements.PodDashboardRequest)) string {
	t.Helper()
	reads := 0
	out, e := podMetricHTTPSource(podMetricHTTPMetadata(), &reads).DiscoverPodMetrics(context.Background(), podMetricQuestion())
	if e != nil || len(out.Instruments) != 2 {
		t.Fatal(out, e)
	}
	in := measurements.PodDashboardRequest{Discovery: *out}
	if change != nil {
		change(&in)
	}
	raw, _ := json.Marshal(in)
	return string(raw)
}
func TestPodMetricHTTPWireContracts(t *testing.T) {
	plan := podMetricHTTPPlanBody(t, nil)
	cases := []planningWireCase{
		{Name: "pod-metric-inventory", Path: "/pod-metric-discovery", Body: podMetricHTTPBody, Status: 200},
		{Name: "pod-metric-inventory-invalid", Path: "/pod-metric-discovery", Body: `{}`, Status: 400},
		{Name: "pod-metric-inventory-unknown", Path: "/pod-metric-discovery", Body: strings.TrimSuffix(podMetricHTTPBody, "}") + `,"service":"fake"}`, Status: 400},
		{Name: "pod-metric-inventory-exact-size", Path: "/pod-metric-discovery", Body: podMetricHTTPBody, PadTo: 65536, Status: 200},
		{Name: "pod-metric-inventory-oversize", Path: "/pod-metric-discovery", Body: podMetricHTTPBody, PadTo: 65537, Status: 400},
		{Name: "pod-metric-inventory-source-failure", Path: "/pod-metric-discovery", Body: podMetricHTTPBody, Status: 502},
		{Name: "pod-metric-plan-no-candidates", Path: "/pod-dashboard/plan", Body: podMetricHTTPPlanBody(t, func(in *measurements.PodDashboardRequest) { in.Discovery.Instruments = nil }), Status: 200},
		{Name: "pod-metric-plan", Path: "/pod-dashboard/plan", Body: plan, Status: 200},
		{Name: "pod-metric-plan-wrong-key", Path: "/pod-dashboard/plan", Body: podMetricHTTPPlanBody(t, func(in *measurements.PodDashboardRequest) { in.Discovery.Key = "wrong" }), Status: 400},
		{Name: "pod-metric-plan-bad-omissions", Path: "/pod-dashboard/plan", Body: podMetricHTTPPlanBody(t, func(in *measurements.PodDashboardRequest) {
			in.Discovery.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = -1
		}), Status: 400},
		{Name: "pod-metric-plan-exact-size", Path: "/pod-dashboard/plan", Body: plan, PadTo: nativehttp.PodDashboardBodyLimit, Status: 200},
		{Name: "pod-metric-plan-oversize", Path: "/pod-dashboard/plan", Body: plan, PadTo: nativehttp.PodDashboardBodyLimit + 1, Status: 400},
		{Name: "pod-metric-inventory-method", Method: "GET", Path: "/pod-metric-discovery", Status: 405},
		{Name: "pod-metric-plan-method", Method: "GET", Path: "/pod-dashboard/plan", Status: 405},
	}

	for _, mode := range []string{"null-schema", "missing-schema", "missing-uid", "wrong-environment"} {
		var body map[string]any
		json.Unmarshal([]byte(podMetricHTTPBody), &body)
		switch mode {
		case "null-schema":
			body["resourceSchemaURL"] = nil
		case "missing-schema":
			delete(body, "resourceSchemaURL")
		case "missing-uid":
			delete(body["resource"].(map[string]any), "k8s.pod.uid")
		case "wrong-environment":
			body["environment"] = "other"
		}
		raw, _ := json.Marshal(body)
		cases = append(cases, planningWireCase{Name: mode, Path: "/pod-metric-discovery", Body: string(raw), Status: 400})
	}

	for _, name := range []string{"escaped-resource", "escaped-schema", "source-row-sentinel"} {
		cases = append(cases, planningWireCase{Name: name, Path: "/pod-metric-discovery", Body: podMetricHTTPBody, Status: 502})
	}
	for i := range cases {
		if cases[i].Method == "" {
			cases[i].Method = "POST"
		}
	}
	run := func(c planningWireCase) *httptest.ResponseRecorder {
		reads := 0
		mux := http.NewServeMux()
		source := podMetricHTTPSource(podMetricHTTPMetadata(), &reads)
		if c.Name == "escaped-resource" {
			source = podMetricHTTPSource(strings.ReplaceAll(podMetricHTTPMetadata(), "pod-1", "another-pod"), &reads)
		}
		if c.Name == "escaped-schema" {
			source = podMetricHTTPSource(strings.ReplaceAll(podMetricHTTPMetadata(), `"resourceSchemaURL":""`, `"resourceSchemaURL":"other"`), &reads)
		}
		if c.Name == "source-row-sentinel" {
			source = podMetricHTTPSource(strings.Repeat(strings.Split(podMetricHTTPMetadata(), "\n")[0]+"\n", 10001), &reads)
		}

		if c.Name == "pod-metric-inventory-source-failure" {
			source.Client = &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
				reads++
				return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("fixture unavailable")), Header: make(http.Header)}, nil
			})}
		}
		nativehttp.RegisterMeasurements(mux, source)
		body := c.Body
		if c.PadTo > len(body) {
			body += strings.Repeat(" ", c.PadTo-len(body))
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
		want := 0
		if c.Path == "/pod-metric-discovery" && (c.Status == 200 || c.Status == 502) {
			want = 1
		}
		if reads != want {
			t.Fatal("wrong pod metric source effects", reads, want)
		}
		return w
	}
	const filename = "../../testdata/http/pod-metrics.json"
	if os.Getenv("GOO_UPDATE_POD_METRIC_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := run(cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		raw, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e := os.WriteFile(filename, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	fixture := dashboardProfileFixture(t, filename, len(cases))
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := run(c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("pod metric header changed", k)
				}
			}
		})
	}
}
func TestPodMetricHTTPSourceAdmissionAndPurePlanning(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	cancel := httptest.NewRecorder()
	budget.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	req := httptest.NewRequest("POST", "/pod-metric-discovery", strings.NewReader(podMetricHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, req)
	if cancel.Code != 202 || w.Code != 409 || called {
		t.Fatal("cancelled pod metric source admitted", w.Code)
	}
	reads := 0
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, podMetricHTTPSource(podMetricHTTPMetadata(), &reads))
	budget = newQueryBudget(mux, 0)
	req = httptest.NewRequest("POST", "/pod-dashboard/plan", strings.NewReader(podMetricHTTPPlanBody(t, nil)))
	req.Header.Set("X-Goo-Query-ID", "not-a-source-capability")
	w = httptest.NewRecorder()
	budget.ServeHTTP(w, req)
	if w.Code != 200 || reads != 0 {
		t.Fatal("pure pod metric planner source admission", w.Code, reads)
	}
	ctx, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/pod-dashboard/plan", strings.NewReader(podMetricHTTPPlanBody(t, nil)))
	mux.ServeHTTP(w, req.WithContext(ctx))
	if w.Code != 400 {
		t.Fatal("expired pod metric plan returned candidates", w.Code)
	}
}

func TestPodMetricHTTPNativeExecuteCaptureOfflineReopen(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	dir := t.TempDir()
	reads := 0
	offline := false
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if offline {
			t.Fatal("offline pod evidence read telemetry")
		}
		reads++
		rawSQL, _ := io.ReadAll(r.Body)
		sql := string(rawSQL)
		if strings.Contains(sql, "ServiceName") || strings.Contains(sql, "mapUpdate") {
			t.Fatal("pod query manufactured service identity", sql)
		}
		payload := podMetricHTTPMetadata()
		if !strings.Contains(sql, "UNION ALL") {
			name := ""
			for _, values := range r.URL.Query() {
				for _, value := range values {
					if value == "k8s.pod.memory.usage" || value == "k8s.pod.memory.working_set" {
						name = value
					}
				}
			}
			resource := podMetricQuestion().Resource
			attrs := map[string]string{}
			unit, value, tempo := "By", "100", 0
			row := map[string]any{"resource": resource, "resource_schema_url": "", "attributes": attrs, "scope": "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver", "version": "0.137.0", "schema_url": "", "name": name, "unit": unit, "value": value, "start_ns": "100000000000", "end_ns": "101000000000", "temporality": tempo, "monotonic": false, "flags": 0}
			raw, _ := json.Marshal(row)
			payload = string(raw) + "\n"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	makeMux := func() *http.ServeMux {
		m := http.NewServeMux()
		registerMeasurements(m, measurements.Source{Endpoint: "http://source", Client: client}, dir)
		capturehttp.Register(m, exploration.Source{Endpoint: "http://source", Client: client}, dir, resultSnapshots)
		return m
	}
	mux := makeMux()
	observed := []apiObservation{}
	call := func(method, path, pattern string, input any, want int) []byte {
		var raw []byte
		if input != nil {
			raw, _ = json.Marshal(input)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(raw))))
		if w.Code != want {
			t.Fatal(method, path, w.Code, w.Body.String())
		}
		observed = append(observed, apiObservation{method, pattern, w.Code, raw, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	var inventory measurements.PodMetricDiscovery
	raw := call("POST", "/pod-metric-discovery", "/pod-metric-discovery", podMetricQuestion(), 200)
	if json.Unmarshal(raw, &inventory) != nil || len(inventory.Instruments) != 2 || reads != 1 {
		t.Fatal(string(raw), reads)
	}
	var plan measurements.PodDashboardPlan
	raw = call("POST", "/pod-dashboard/plan", "/pod-dashboard/plan", measurements.PodDashboardRequest{Discovery: inventory}, 200)
	if json.Unmarshal(raw, &plan) != nil || len(plan.Widgets) != 2 || reads != 1 {
		t.Fatal(string(raw), reads)
	}
	for _, widget := range plan.Widgets {
		raw := call("POST", "/measurements", "/measurements", widget.Question, 200)
		var result measurements.Result
		if json.Unmarshal(raw, &result) != nil || measurements.ValidateResult(&result) != nil {
			t.Fatal("invalid exactpod result", string(raw))
		}
		stream := widget.Question.Definition.Source.ExactStream
		if stream.Resource["service.name"] != "" || stream.Resource["service.instance.id"] != "" {
			t.Fatal("pod needs service identity")
		}
		if widget.Question.Definition.Kind == "stock" && result.Stocks == nil || widget.Question.Definition.Kind == "gauge" && result.Gauges == nil {
			t.Fatal("wrong pod kernel", string(raw))
		}
	}
	pages := []exploration.CapturePage{}
	native := [][]byte{}
	for _, widget := range plan.Widgets {
		q := widget.Question
		raw := call("POST", "/captures", "/captures", map[string]any{"title": "Host exact stream evidence", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}}, 201)
		var page exploration.CapturePage
		if json.Unmarshal(raw, &page) != nil || page.Measurement == nil {
			t.Fatal(string(raw))
		}
		pages = append(pages, page)
		held, _ := json.Marshal(page.Measurement)
		native = append(native, held)
	}
	before := reads
	offline = true
	mux = makeMux()
	for i, page := range pages {
		raw := call("GET", "/captures/"+page.Receipt.ID, "/captures/{id}", nil, 200)
		var reopened exploration.CapturePage
		_ = json.Unmarshal(raw, &reopened)
		held, _ := json.Marshal(reopened.Measurement)
		if string(held) != string(native[i]) || reads != before {
			t.Fatal("pod evidence changed/reread on offline reopen")
		}
	}
	if path := os.Getenv("GOO_POD_METRIC_API_OBSERVATIONS"); path != "" {
		raw, _ := json.MarshalIndent(observed, "", "  ")
		if e := os.WriteFile(path, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}
func TestPodMetricSharedAdmissionAndInflightCancellation(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	cancelled := httptest.NewRecorder()
	budget.ServeHTTP(cancelled, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	req := httptest.NewRequest("POST", "/pod-metric-discovery", strings.NewReader(podMetricHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, req)
	if cancelled.Code != 202 || w.Code != 409 || called {
		t.Fatal("pre-cancelled pod query admitted")
	}
	started := make(chan struct{}, 1)
	source := measurements.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		started <- struct{}{}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}}
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, source)
	budget = newQueryBudget(mux, 1)
	id = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	req = httptest.NewRequest("POST", "/pod-metric-discovery", strings.NewReader(podMetricHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { w := httptest.NewRecorder(); budget.ServeHTTP(w, req); done <- w }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("pod source not started")
	}
	busy := httptest.NewRecorder()
	budget.ServeHTTP(busy, httptest.NewRequest("POST", "/explore", strings.NewReader(`{}`)))
	if busy.Code != 429 {
		t.Fatal("pod discovery has independent source admission")
	}
	cancelled = httptest.NewRecorder()
	budget.ServeHTTP(cancelled, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	select {
	case w := <-done:
		if cancelled.Code != 202 || w.Code != 502 {
			t.Fatal("pod cancel became data", w.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("pod source cancellation not propagated")
	}
}

func TestPodMetricHTTPPlanAcceptsLargeLegitimateDiscovery(t *testing.T) {
	// Real source decoding builds retained metadata with 25 distinct previews per
	// instrument. Repeated raw resource/stream descriptions legitimately exceed
	// the original 64 KiB transport budget without exceeding semantic limits.
	rows := []string{}
	for _, line := range strings.Split(strings.TrimSpace(podMetricHTTPMetadata()), "\n") {
		for index := 0; index < 25; index++ {
			var row map[string]any
			if err := json.Unmarshal([]byte(line), &row); err != nil {
				t.Fatal(err)
			}
			row["attributes"] = map[string]string{"variant": strings.Repeat("x", 1500) + strings.Repeat("y", index)}
			raw, _ := json.Marshal(row)
			rows = append(rows, string(raw))
		}
	}
	reads := 0
	discovery, err := podMetricHTTPSource(strings.Join(rows, "\n")+"\n", &reads).DiscoverPodMetrics(context.Background(), podMetricQuestion())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(measurements.PodDashboardRequest{Discovery: *discovery})
	if err != nil || len(raw) <= 65536 || len(raw) > nativehttp.PodDashboardBodyLimit {
		t.Fatal("invalid large metadata fixture", len(raw), err)
	}
	sourceReads := 0
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, podMetricHTTPSource("", &sourceReads))
	budget := newQueryBudget(mux, 0)
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/pod-dashboard/plan", strings.NewReader(string(raw))))
	var plan measurements.PodDashboardPlan
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || len(plan.Widgets) != 2 || sourceReads != 0 {
		t.Fatal(len(raw), w.Code, w.Body.String(), sourceReads)
	}
	t.Logf("Accepted %d-byte real retained discovery without source admission", len(raw))
}
