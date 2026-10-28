package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCPUTimeHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
	testHostIntervalCounterHTTP(t, "system.cpu.time")
}
func TestNetworkHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
	testHostIntervalCounterHTTP(t, "system.network.io")
}
func TestDiskHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
	testHostIntervalCounterHTTP(t, "system.disk.io")
}
func testHostIntervalCounterHTTP(t *testing.T, metric string) {
	resultSnapshots := &exploration.ResultSnapshots{}
	network := metric != "system.cpu.time"
	s := metrics.Stream{Resource: map[string]string{"host.id": "host", "deployment.environment.name": "prod"}, Scope: "hostmetrics", SchemaURL: measurements.HostMetricsSchema, Name: "system.cpu.time", Unit: "s", Attributes: map[string]string{"cpu.mode": "user", "cpu.logical_number": "0"}}
	if network {
		s.Name = "system.network.io"
		s.Unit = "By"
		s.Attributes = map[string]string{"network.interface.name": "eth0", "network.io.direction": "receive"}
	}
	if metric == "system.disk.io" {
		s.Name = metric
		s.Attributes = map[string]string{"system.device": "sda", "disk.io.direction": "read"}
	}
	first, last := "1.25", "1.75"
	if network {
		first, last = "100", "612"
	}
	q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "cpu-time", Title: "CPU time", Kind: "counter", CounterDomain: metrics.DecimalCounterDomain, Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "host", ID: "host"}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 3000}
	rows := func() string {
		var b strings.Builder
		for i, end := range []string{"1000000000", "3000000000"} {
			raw, _ := json.Marshal(map[string]any{"resource": s.Resource, "resource_schema_url": "", "scope": s.Scope, "version": "", "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "attributes": s.Attributes, "flags": 0, "monotonic": true, "temporality": 2, "start_ns": "0", "end_ns": end, "value": []string{first, last}[i]})
			b.Write(raw)
			b.WriteByte('\n')
		}
		return b.String()
	}
	mode := "ok"
	offline := false
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if offline {
			t.Fatal("offline reopen read source")
		}
		reads++
		sql, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(sql), "otel_metrics_sum") || strings.Contains(string(sql), "ServiceName IN") {
			t.Fatal("CPU-time authority/adapter", string(sql))
		}
		body := rows()
		switch mode {
		case "fractional-bytes":
			body = strings.ReplaceAll(body, `"value":"`+first+`"`, `"value":"100.000000000000000001"`)
		case "unsafe-bytes":
			body = strings.ReplaceAll(body, `"value":"`+first+`"`, `"value":"9007199254740992"`)
		case "negative":
			body = strings.ReplaceAll(body, `"value":"`+first+`"`, `"value":"-1"`)
		case "flags-missing":
			body = strings.ReplaceAll(body, `"flags":0,`, ``)
		case "delta-profile":
			body = strings.ReplaceAll(body, `"temporality":2`, `"temporality":1`)
		case "fractional-temporality":
			body = strings.ReplaceAll(body, `"temporality":2`, `"temporality":2.0000000000000000001`)
		case "fractional-flags":
			body = strings.ReplaceAll(body, `"flags":0`, `"flags":0.00000000000000001`)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	dir := t.TempDir()
	makeMux := func() *http.ServeMux {
		mux := http.NewServeMux()
		registerMeasurements(mux, measurements.Source{Endpoint: "http://source", Client: client}, dir)
		capturehttp.Register(mux, exploration.Source{Endpoint: "http://source", Client: client}, dir, resultSnapshots)
		return mux
	}
	mux := makeMux()
	observed := []apiObservation{}
	call := func(method, path, pattern string, input any, want int) []byte {
		var body []byte
		if input != nil {
			body, _ = json.Marshal(input)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(body))))
		if w.Code != want {
			t.Fatal(method, path, w.Code, w.Body.String())
		}
		observed = append(observed, apiObservation{method, pattern, w.Code, body, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	raw := call("POST", "/measurements", "/measurements", q, 200)
	var result measurements.Result
	if json.Unmarshal(raw, &result) != nil || measurements.ValidateResult(&result) != nil || result.Counters == nil {
		t.Fatal(string(raw))
	}
	p := result.Counters.Series[0].Intervals[0]
	delta, exact, rate := "0.5", "1/4", .25
	if network {
		delta, exact, rate = "512", "256", 256
	}
	if p.Delta != delta || p.ExactPerSecond != exact || p.PerSecond == nil || *p.PerSecond != rate {
		t.Fatal(p)
	}
	raw = call("POST", "/captures", "/captures", map[string]any{"title": "CPU activity", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}}, 201)
	var page exploration.CapturePage
	if json.Unmarshal(raw, &page) != nil || page.Measurement == nil {
		t.Fatal(string(raw))
	}
	frozen, _ := json.Marshal(page.Measurement)
	before := reads
	offline = true
	mux = makeMux()
	raw = call("GET", "/captures/"+page.Receipt.ID, "/captures/{id}", nil, 200)
	var reopened exploration.CapturePage
	_ = json.Unmarshal(raw, &reopened)
	again, _ := json.Marshal(reopened.Measurement)
	if string(frozen) != string(again) || reads != before {
		t.Fatal("frozen CPU-time evidence changed/reread")
	}
	offline = false
	for _, bad := range []string{"negative", "flags-missing", "fractional-flags", "fractional-temporality", "delta-profile"} {
		mode = bad
		call("POST", "/measurements", "/measurements", q, 502)
	}
	if network {
		for _, bad := range []string{"fractional-bytes", "unsafe-bytes"} {
			mode = bad
			call("POST", "/measurements", "/measurements", q, 502)
		}
	}
	bad := q
	bad.Definition.CounterDomain = ""
	call("POST", "/measurements", "/measurements", bad, 400)
	variable := "GOO_CPU_TIME_API_OBSERVATIONS"
	if network {
		variable = "GOO_NETWORK_API_OBSERVATIONS"
	}
	if metric == "system.disk.io" {
		variable = "GOO_DISK_API_OBSERVATIONS"
	}
	if file := os.Getenv(variable); file != "" {
		b, _ := json.MarshalIndent(observed, "", "  ")
		if e := os.WriteFile(file, b, 0644); e != nil {
			t.Fatal(e)
		}
	}
}
