package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestStockDomainHTTPExecuteCaptureAndOfflineReopen(t *testing.T) {
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
		payload := strings.ReplaceAll(stockHTTPRows(2, false), `"value":"7.5"`, `"value":"7"`)
		if mode == "negative" {
			payload = strings.ReplaceAll(payload, `"value":"7"`, `"value":"-1"`)
		}
		if mode == "fractional" {
			payload = strings.ReplaceAll(payload, `"value":"7"`, `"value":"7.5"`)
		}
		if mode == "unsafe" {
			payload = strings.ReplaceAll(payload, `"value":"7"`, `"value":"9007199254740992"`)
		}
		if mode == "safe-boundary" {
			payload = strings.ReplaceAll(payload, `"value":"7"`, `"value":"9007199254740991"`)
		}
		if mode == "missing" {
			payload = strings.ReplaceAll(payload, `"flags":0`, `"flags":1`)
			payload = strings.ReplaceAll(payload, `"value":"7"`, `"value":"-1"`)
		}

		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
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
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("missing no-store")
		}
		observed = append(observed, apiObservation{method, pattern, w.Code, body, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	q := stockHTTPQuestion()
	q.Definition.StockDomain = "nonnegative-integer"
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
	if len(samples) != 3 || samples[1].Value == nil || *samples[1].Value != 7 || samples[2].StartNS != "102000000000" {
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
	for _, m := range []string{"negative", "fractional", "unsafe"} {
		mode = m
		call("POST", "/measurements", "/measurements", q, 502)
	}
	mode = "safe-boundary"
	call("POST", "/measurements", "/measurements", q, 200)
	mode = "missing"
	missingRaw := call("POST", "/measurements", "/measurements", q, 200)
	var missing measurements.Result
	_ = json.Unmarshal(missingRaw, &missing)
	if missing.Stocks == nil || missing.Stocks.Series[0].State != "unknown" || missing.Stocks.Series[0].Samples[1].Value != nil {
		t.Fatal("missing flag became a domain value", string(missingRaw))
	}
	invalid := q
	invalid.Definition.Kind = "gauge"
	before = reads
	call("POST", "/measurements", "/measurements", invalid, 400)
	if reads != before {
		t.Fatal("invalid stock read source")
	}
	invalid = q
	invalid.Definition.StockDomain = "unknown-domain"
	call("POST", "/measurements", "/measurements", invalid, 400)
	if reads != before {
		t.Fatal("invalid domain read source")
	}
	if path := os.Getenv("GOO_STOCK_DOMAIN_API_OBSERVATIONS"); path != "" {
		out, _ := json.MarshalIndent(observed, "", "  ")
		if err := os.WriteFile(path, append(out, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStockDomainHTTPFrozenContracts(t *testing.T) {
	q := stockHTTPQuestion()
	q.Definition.StockDomain = "nonnegative-integer"
	body, _ := json.Marshal(q)
	definition, _ := yaml.Marshal(q.Definition)
	preview, _ := json.Marshal(map[string]any{"yaml": string(definition), "fromMs": q.FromMS, "toMs": q.ToMS})
	wrongKind := q
	wrongKind.Definition.Kind = "gauge"
	wrongKindBody, _ := json.Marshal(wrongKind)
	wrongPolicy := q
	wrongPolicy.Definition.StockDomain = "unknown"
	wrongPolicyBody, _ := json.Marshal(wrongPolicy)
	wrongType := strings.Replace(string(body), `"stockDomain":"nonnegative-integer"`, `"stockDomain":true`, 1)
	cases := []planningWireCase{{Name: "stock-domain-plan", Method: "POST", Path: "/measurement-packages/plan", Body: string(preview), Status: 200}, {Name: "stock-domain-negative", Method: "POST", Path: "/measurements", Body: string(body), Status: 502}, {Name: "stock-domain-fractional", Method: "POST", Path: "/measurements", Body: string(body), Status: 502}, {Name: "stock-domain-unsafe", Method: "POST", Path: "/measurements", Body: string(body), Status: 502}, {Name: "stock-domain-wrong-kind", Method: "POST", Path: "/measurements", Body: string(wrongKindBody), Status: 400}, {Name: "stock-domain-unknown", Method: "POST", Path: "/measurements", Body: string(wrongPolicyBody), Status: 400}, {Name: "stock-domain-wrong-type", Method: "POST", Path: "/measurements", Body: wrongType, Status: 400}}
	run := func(c planningWireCase) *httptest.ResponseRecorder {
		reads := 0
		client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
			reads++
			payload := stockHTTPRows(2, false)
			if c.Name == "stock-domain-negative" {
				payload = strings.ReplaceAll(payload, `"value":"7.5"`, `"value":"-1"`)
			}
			if c.Name == "stock-domain-unsafe" {
				payload = strings.ReplaceAll(payload, `"value":"7.5"`, `"value":"9007199254740992"`)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
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
			t.Fatal("domain effects", reads, want)
		}
		return w
	}
	const filename = "../../testdata/http/stock-domain.json"
	if os.Getenv("GOO_UPDATE_STOCK_DOMAIN_HTTP_FIXTURES") == "1" {
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
