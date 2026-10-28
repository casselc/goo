package main

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
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

const hostDiscoveryHTTPBody = `{"environment":"prod","fromMs":100000,"toMs":104000}`

func hostHTTPMetadata() string {
	rows := []string{}
	resource := map[string]string{"host.id": "machine-1", "host.name": "display-label", "deployment.environment.name": "prod"}
	for _, metric := range []string{"system.cpu.utilization", "system.memory.usage", "system.cpu.time"} {
		kind, unit, tempo, mono := "gauge", "1", 0, false
		attrs := map[string]string{"cpu.mode": "user", "cpu.logical_number": "0"}
		if metric == "system.memory.usage" {
			kind, unit, tempo = "counter", "By", 2
			attrs = map[string]string{"system.memory.state": "used"}
		}
		if metric == "system.cpu.time" {
			kind, unit, tempo, mono = "counter", "s", 2, true
		}
		raw, _ := json.Marshal(map[string]any{"kind": kind, "scope": "host-sdk", "version": "", "schemaURL": measurements.HostMetricsSchema, "metric": metric, "unit": unit, "temporality": tempo, "monotonic": mono, "resource": resource, "resourceSchemaURL": "resource-schema", "attributes": attrs})
		rows = append(rows, string(raw))
	}
	return strings.Join(rows, "\n") + "\n"
}
func hostHTTPSource(payload string, reads *int) measurements.Source {
	return measurements.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		*reads++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}}
}
func hostHTTPPlanBody(t *testing.T, change func(*measurements.HostDashboardRequest)) string {
	t.Helper()
	reads := 0
	out, e := hostHTTPSource(hostHTTPMetadata(), &reads).DiscoverHosts(context.Background(), measurements.HostDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 104000})
	if e != nil || len(out.Hosts) != 1 {
		t.Fatal(out, e)
	}
	in := measurements.HostDashboardRequest{Question: out.Question, Host: out.Hosts[0]}
	if change != nil {
		change(&in)
	}
	raw, _ := json.Marshal(in)
	return string(raw)
}
func TestHostHTTPWireContracts(t *testing.T) {
	plan := hostHTTPPlanBody(t, nil)
	cases := []planningWireCase{
		{Name: "host-inventory", Path: "/host-discovery", Body: hostDiscoveryHTTPBody, Status: 200},
		{Name: "host-inventory-invalid", Path: "/host-discovery", Body: `{}`, Status: 400},
		{Name: "host-inventory-unknown", Path: "/host-discovery", Body: strings.TrimSuffix(hostDiscoveryHTTPBody, "}") + `,"service":"fake"}`, Status: 400},
		{Name: "host-inventory-exact-size", Path: "/host-discovery", Body: hostDiscoveryHTTPBody, PadTo: 16384, Status: 200},
		{Name: "host-inventory-oversize", Path: "/host-discovery", Body: hostDiscoveryHTTPBody, PadTo: 16385, Status: 400},
		{Name: "host-inventory-source-failure", Path: "/host-discovery", Body: hostDiscoveryHTTPBody, Status: 502},
		{Name: "host-plan-no-candidates", Path: "/host-dashboard/plan", Body: hostHTTPPlanBody(t, func(in *measurements.HostDashboardRequest) { in.Host.Instruments = nil }), Status: 200},
		{Name: "host-plan", Path: "/host-dashboard/plan", Body: plan, Status: 200},
		{Name: "host-plan-wrong-key", Path: "/host-dashboard/plan", Body: hostHTTPPlanBody(t, func(in *measurements.HostDashboardRequest) { in.Host.Key = "wrong" }), Status: 400},
		{Name: "host-plan-bad-omissions", Path: "/host-dashboard/plan", Body: hostHTTPPlanBody(t, func(in *measurements.HostDashboardRequest) {
			in.Host.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = -1
		}), Status: 400},
		{Name: "host-plan-exact-size", Path: "/host-dashboard/plan", Body: plan, PadTo: 65536, Status: 200},
		{Name: "host-plan-oversize", Path: "/host-dashboard/plan", Body: plan, PadTo: 65537, Status: 400},
		{Name: "host-inventory-method", Method: "GET", Path: "/host-discovery", Status: 405},
		{Name: "host-plan-method", Method: "GET", Path: "/host-dashboard/plan", Status: 405},
	}
	for i := range cases {
		if cases[i].Method == "" {
			cases[i].Method = "POST"
		}
	}
	run := func(c planningWireCase) *httptest.ResponseRecorder {
		reads := 0
		mux := http.NewServeMux()
		source := hostHTTPSource(hostHTTPMetadata(), &reads)
		if c.Name == "host-inventory-source-failure" {
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
		if c.Path == "/host-discovery" && (c.Status == 200 || c.Status == 502) {
			want = 1
		}
		if reads != want {
			t.Fatal("wrong host source effects", reads, want)
		}
		return w
	}
	const filename = "../../testdata/http/host.json"
	if os.Getenv("GOO_UPDATE_HOST_HTTP_FIXTURES") == "1" {
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
					t.Fatal("host header changed", k)
				}
			}
		})
	}
}
func TestHostHTTPSourceAdmissionAndPurePlanning(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	cancel := httptest.NewRecorder()
	budget.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	req := httptest.NewRequest("POST", "/host-discovery", strings.NewReader(hostDiscoveryHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, req)
	if cancel.Code != 202 || w.Code != 409 || called {
		t.Fatal("cancelled host source admitted", w.Code)
	}
	reads := 0
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, hostHTTPSource(hostHTTPMetadata(), &reads))
	budget = newQueryBudget(mux, 0)
	req = httptest.NewRequest("POST", "/host-dashboard/plan", strings.NewReader(hostHTTPPlanBody(t, nil)))
	req.Header.Set("X-Goo-Query-ID", "not-a-source-capability")
	w = httptest.NewRecorder()
	budget.ServeHTTP(w, req)
	if w.Code != 200 || reads != 0 {
		t.Fatal("pure host planner source admission", w.Code, reads)
	}
	ctx, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/host-dashboard/plan", strings.NewReader(hostHTTPPlanBody(t, nil)))
	mux.ServeHTTP(w, req.WithContext(ctx))
	if w.Code != 400 {
		t.Fatal("expired host plan returned candidates", w.Code)
	}
}

func TestHostHTTPNativeExecuteCaptureOfflineReopen(t *testing.T) {
	dir := t.TempDir()
	reads := 0
	offline := false
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if offline {
			t.Fatal("offline host evidence read telemetry")
		}
		reads++
		rawSQL, _ := io.ReadAll(r.Body)
		sql := string(rawSQL)
		if strings.Contains(sql, "ServiceName") || strings.Contains(sql, "mapUpdate") {
			t.Fatal("host query manufactured service identity", sql)
		}
		payload := hostHTTPMetadata()
		if !strings.Contains(sql, "UNION ALL") {
			name := ""
			for _, values := range r.URL.Query() {
				for _, value := range values {
					if value == "system.cpu.utilization" || value == "system.memory.usage" {
						name = value
					}
				}
			}
			resource := map[string]string{"host.id": "machine-1", "host.name": "display-label", "deployment.environment.name": "prod"}
			attrs := map[string]string{"cpu.mode": "user", "cpu.logical_number": "0"}
			unit, value, tempo := "1", "0.25", 0
			if name == "system.memory.usage" {
				attrs = map[string]string{"system.memory.state": "used"}
				unit, value, tempo = "By", "100", 2
			}
			row := map[string]any{"resource": resource, "resource_schema_url": "resource-schema", "attributes": attrs, "scope": "host-sdk", "version": "", "schema_url": measurements.HostMetricsSchema, "name": name, "unit": unit, "value": value, "start_ns": "100000000000", "end_ns": "101000000000", "temporality": tempo, "monotonic": false, "flags": 0}
			raw, _ := json.Marshal(row)
			payload = string(raw) + "\n"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	makeMux := func() *http.ServeMux {
		m := http.NewServeMux()
		registerMeasurements(m, measurements.Source{Endpoint: "http://source", Client: client}, dir)
		registerCaptures(m, exploration.Source{Endpoint: "http://source", Client: client}, dir)
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
	var inventory measurements.HostDiscovery
	raw := call("POST", "/host-discovery", "/host-discovery", measurements.HostDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 104000}, 200)
	if json.Unmarshal(raw, &inventory) != nil || len(inventory.Hosts) != 1 || reads != 1 {
		t.Fatal(string(raw), reads)
	}
	var plan measurements.HostDashboardPlan
	raw = call("POST", "/host-dashboard/plan", "/host-dashboard/plan", measurements.HostDashboardRequest{Question: inventory.Question, Host: inventory.Hosts[0]}, 200)
	if json.Unmarshal(raw, &plan) != nil || len(plan.Widgets) != 2 || reads != 1 {
		t.Fatal(string(raw), reads)
	}
	for _, widget := range plan.Widgets {
		raw := call("POST", "/measurements", "/measurements", widget.Question, 200)
		var result measurements.Result
		if json.Unmarshal(raw, &result) != nil || measurements.ValidateResult(&result) != nil {
			t.Fatal("invalid exacthost result", string(raw))
		}
		stream := widget.Question.Definition.Source.ExactStream
		if stream.Resource["service.name"] != "" || stream.Resource["service.instance.id"] != "" {
			t.Fatal("host needs service identity")
		}
		if widget.Question.Definition.Kind == "stock" && result.Stocks == nil || widget.Question.Definition.Kind == "gauge" && result.Gauges == nil {
			t.Fatal("wrong host kernel", string(raw))
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
			t.Fatal("host evidence changed/reread on offline reopen")
		}
	}
	if path := os.Getenv("GOO_HOST_API_OBSERVATIONS"); path != "" {
		raw, _ := json.MarshalIndent(observed, "", "  ")
		if e := os.WriteFile(path, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}
