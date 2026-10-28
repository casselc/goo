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

func TestFilesystemUtilizationHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	s := metrics.Stream{Resource: map[string]string{"host.id": "host", "deployment.environment.name": "prod"}, Scope: "hostmetrics", SchemaURL: measurements.HostMetricsSchema, Name: "system.filesystem.utilization", Unit: "1", Attributes: map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.state": "used"}}
	q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "filesystem", Title: "Filesystem", Kind: "gauge", Environment: "prod", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Resource: &measurements.ResourceBinding{Kind: "host", ID: "host"}, Services: []string{}, Scope: s.Scope, SchemaURL: s.SchemaURL, Metric: s.Name, Unit: s.Unit, ExactStream: &s}}, FromMS: 1000, ToMS: 3000}
	first, last := "0.25", "0.75"
	rows := func() string {
		var b strings.Builder
		for i, end := range []string{"1000000000", "2000000000"} {
			raw, _ := json.Marshal(map[string]any{"resource": s.Resource, "resource_schema_url": "", "scope": s.Scope, "version": "", "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "attributes": s.Attributes, "flags": 0, "end_ns": end, "value": []string{first, last}[i]})
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
		if !strings.Contains(string(sql), "otel_metrics_gauge") || strings.Contains(string(sql), "ServiceName IN") {
			t.Fatal("filesystem authority/adapter", string(sql))
		}
		body := rows()

		switch mode {
		case "negative":
			body = strings.ReplaceAll(body, `"value":"0.25"`, `"value":"-1"`)
		case "above-one":
			body = strings.ReplaceAll(body, `"value":"0.25"`, `"value":"1.1"`)
		case "rounded-above":
			body = strings.ReplaceAll(body, `"value":"0.25"`, `"value":"1.00000000000000000001"`)
		case "tiny-negative":
			body = strings.ReplaceAll(body, `"value":"0.25"`, `"value":"-0.00000000000000000001"`)
		case "flags-missing":
			body = strings.ReplaceAll(body, `"flags":0,`, ``)
		case "fractional-flags":
			body = strings.ReplaceAll(body, `"flags":0`, `"flags":0.000000000000001`)
		case "maps-null":
			body = strings.ReplaceAll(body, `"attributes":{`, `"attributes":null,"ignored":{`)
		case "unit-drift":
			body = strings.ReplaceAll(body, `"unit":"1"`, `"unit":"percent"`)
		case "null-scope":
			body = strings.ReplaceAll(body, `"scope":"hostmetrics"`, `"scope":null`)
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
	if json.Unmarshal(raw, &result) != nil || measurements.ValidateResult(&result) != nil || result.Gauges == nil {
		t.Fatal(string(raw))
	}
	if result.Gauges == nil || len(result.Gauges.Series[0].Samples) != 2 || *result.Gauges.Series[0].Samples[1].Value != 0.75 {
		t.Fatal("reported fraction lost", result)
	}
	raw = call("POST", "/captures", "/captures", map[string]any{"title": "Filesystem usage", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}}, 201)
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
		t.Fatal("frozen filesystem evidence changed/reread")
	}
	offline = false
	for _, bad := range []string{"negative", "above-one", "rounded-above", "tiny-negative", "flags-missing", "fractional-flags", "maps-null", "unit-drift", "null-scope"} {
		mode = bad
		call("POST", "/measurements", "/measurements", q, 502)
	}
	bad := q
	bad.Definition.Source.Unit = "bytes"
	call("POST", "/measurements", "/measurements", bad, 400)
	if file := os.Getenv("GOO_FILESYSTEM_UTILIZATION_API_OBSERVATIONS"); file != "" {
		b, _ := json.MarshalIndent(observed, "", "  ")
		if e := os.WriteFile(file, b, 0644); e != nil {
			t.Fatal(e)
		}
	}
}
