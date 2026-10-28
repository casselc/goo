package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func stockHTTPQuestion() measurements.Question {
	return measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "pool-stock", Title: "Pool observations", Kind: "stock", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Namespace: &identity.Namespace{Mode: "value"}, Services: []string{"api"}, Scope: "ordinary-sdk", Metric: "db.client.connection.count", Unit: "{connection}"}}, FromMS: 100000, ToMS: 104000}
}
func stockHTTPRows(temporality int, monotonic bool) string {
	rows := []string{}
	for _, point := range []struct{ start, end, value string }{{"100000000000", "100500000000", "10"}, {"100000000000", "101500000000", "7.5"}, {"102000000000", "102500000000", "4"}} {
		row := map[string]any{"resource": map[string]string{"service.name": "api", "service.instance.id": "sdk-1", "deployment.environment.name": "prod", "service.namespace": ""}, "scope": "ordinary-sdk", "version": "", "schema_url": "", "name": "db.client.connection.count", "unit": "{connection}", "attributes": map[string]string{"db.client.connection.pool.name": "main", "db.client.connection.state": "idle"}, "start_ns": point.start, "end_ns": point.end, "value": point.value, "temporality": temporality, "monotonic": monotonic, "flags": 0}
		raw, _ := json.Marshal(row)
		rows = append(rows, string(raw))
	}
	return strings.Join(rows, "\n") + "\n"
}
func TestStockHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	dir := t.TempDir()
	reads := 0
	offline := false
	mode := "stock"
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if offline {
			t.Fatal("reopening stock capture read source")
		}
		reads++
		sql, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(sql), "otel_metrics_sum") || strings.Contains(string(sql), "otel_metrics_gauge") {
			t.Fatal("stock used wrong table", string(sql))
		}
		payload := stockHTTPRows(2, false)
		if mode == "delta" {
			payload = stockHTTPRows(1, false)
		}
		if mode == "monotonic" {
			payload = stockHTTPRows(2, true)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	makeMux := func() *http.ServeMux {
		mux := http.NewServeMux()
		registerMeasurements(mux, measurements.Source{Endpoint: "http://source", Client: client}, dir)
		registerCaptures(mux, exploration.Source{Endpoint: "http://source", Client: client}, dir, resultSnapshots)
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
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("missing no-store")
		}
		observed = append(observed, apiObservation{method, pattern, w.Code, body, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	q := stockHTTPQuestion()
	raw := call("POST", "/measurements", "/measurements", q, 200)
	var result measurements.Result
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if err := measurements.ValidateResult(&result); err != nil {
		t.Fatal(err)
	}
	if result.Stocks == nil || result.Gauges != nil || result.Counters != nil || result.Histograms != nil || len(result.Stocks.Series) != 1 {
		t.Fatal("stock result routed to other kernel", string(raw))
	}
	samples := result.Stocks.Series[0].Samples
	if len(samples) != 3 || samples[1].Value == nil || *samples[1].Value != 7.5 || samples[2].StartNS != "102000000000" {
		t.Fatal("release/epoch lost", samples)
	}
	capture := call("POST", "/captures", "/captures", map[string]any{"title": "Stock evidence", "question": exploration.CaptureQuestion{Kind: "measurement", Measurement: &q}}, 201)
	var page exploration.CapturePage
	if err := json.Unmarshal(capture, &page); err != nil || page.Measurement == nil {
		t.Fatal("stock capture", err, string(capture))
	}
	frozen, _ := json.Marshal(page.Measurement)
	before := reads
	offline = true
	mux = makeMux()
	reopened := call("GET", "/captures/"+page.Receipt.ID, "/captures/{id}", nil, 200)
	var reopenedPage exploration.CapturePage
	if err := json.Unmarshal(reopened, &reopenedPage); err != nil {
		t.Fatal(err)
	}
	again, _ := json.Marshal(reopenedPage.Measurement)
	if string(frozen) != string(again) || reads != before {
		t.Fatal("stock evidence changed or reread")
	}
	offline = false
	mode = "delta"
	call("POST", "/measurements", "/measurements", q, 502)
	mode = "monotonic"
	call("POST", "/measurements", "/measurements", q, 502)
	invalid := q
	invalid.Definition.Kind = "unsupported-stock"
	before = reads
	call("POST", "/measurements", "/measurements", invalid, 400)
	if reads != before {
		t.Fatal("invalid stock read source")
	}
	if path := os.Getenv("GOO_STOCK_API_OBSERVATIONS"); path != "" {
		out, _ := json.MarshalIndent(observed, "", "  ")
		if err := os.WriteFile(path, append(out, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStockHTTPFrozenContracts(t *testing.T) {
	q := stockHTTPQuestion()
	body, _ := json.Marshal(q)
	definition, _ := yaml.Marshal(q.Definition)
	preview, _ := json.Marshal(map[string]any{"yaml": string(definition), "fromMs": q.FromMS, "toMs": q.ToMS})
	cases := []planningWireCase{{Name: "stock-package-plan", Method: "POST", Path: "/measurement-packages/plan", Body: string(preview), Status: 200}, {Name: "stock-delta-refused", Method: "POST", Path: "/measurements", Body: string(body), Status: 502}, {Name: "stock-monotonic-refused", Method: "POST", Path: "/measurements", Body: string(body), Status: 502}, {Name: "stock-unknown-field", Method: "POST", Path: "/measurements", Body: strings.TrimSuffix(string(body), "}") + `,"stocks":{}}`, Status: 400}}
	run := func(c planningWireCase) *httptest.ResponseRecorder {
		reads := 0
		client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
			reads++
			temporality, monotonic := 1, false
			if c.Name == "stock-monotonic-refused" {
				temporality, monotonic = 2, true
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(stockHTTPRows(temporality, monotonic))), Header: make(http.Header)}, nil
		})}
		mux := http.NewServeMux()
		registerMeasurements(mux, measurements.Source{Endpoint: "http://source", Client: client}, t.TempDir())
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(c.Body)))
		want := 0
		if c.Status == 502 {
			want = 1
		}
		if reads != want {
			t.Fatal("stock unexpected source effect", reads, want)
		}
		return w
	}
	const filename = "../../testdata/http/stock.json"
	if os.Getenv("GOO_UPDATE_STOCK_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := run(cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		out, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if err := os.WriteFile(filename, append(out, '\n'), 0644); err != nil {
			t.Fatal(err)
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
					t.Fatal("header changed", k)
				}
			}
		})
	}
}
