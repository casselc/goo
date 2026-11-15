package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/measurements"
)

const rawHistogram447HTTPBody = `{"question":{"environment":"prod","service":"api","fromMs":1000,"toMs":2000},"instrument":{"kind":"histogram","scope":"","version":"","schemaURL":"","metric":"request.duration","unit":"s","temporality":1}}`
const rawHistogram447HTTPRow = `{"resource":{"service.name":"api","deployment.environment.name":"prod"},"attributes":{},"resource_schema_url":"","scope":"","version":"","schema_url":"","name":"request.duration","unit":"s","start_ns":"0","end_ns":"1500000000","flags":0,"temporality":1,"count":"3","bounds":["1"],"bucket_counts":["1","2"],"sum":"0","min":"0","max":"0"}`

func rawHistogram447HTTPResponse(t *testing.T, c planningWireCase) (*httptest.ResponseRecorder, int) {
	t.Helper()
	reads := 0
	schema, err := os.ReadFile("../../internal/sourcecap/testdata/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var columns map[string][]json.RawMessage
	if json.Unmarshal(schema, &columns) != nil {
		t.Fatal("schema fixture")
	}
	source := measurements.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		sql, _ := io.ReadAll(r.Body)
		payload := ""
		status := 200
		switch {
		case strings.HasPrefix(string(sql), "DESCRIBE TABLE otel_metrics_histogram "):
			for _, row := range columns["histogram"] {
				compact, _ := json.Marshal(row)
				payload += string(compact) + "\n"
			}
			if c.Name != "missing-stat-columns" {
				for _, name := range []string{"Sum", "Min", "Max"} {
					payload += "{\"name\":\"" + name + "\",\"type\":\"Float64\"}\n"
				}
			}
		case strings.HasSuffix(string(sql), " LIMIT 0 FORMAT JSONEachRow"):
		default:
			payload = rawHistogram447HTTPRow + "\n"
			switch c.Name {
			case "unknown-temporality":
				payload = strings.ReplaceAll(payload, `"temporality":1`, `"temporality":7`)
			case "metadata-mismatch":
				payload = strings.ReplaceAll(payload, `"temporality":1`, `"temporality":2`)
			case "empty":
				payload = ""
			case "source-failure":
				status = 403
				payload = "not authorized"
			case "malformed-end":
				payload = strings.ReplaceAll(payload, `"1500000000"`, `"2000000000"`)
			case "no-recorded":
				payload = strings.ReplaceAll(payload, `"flags":0`, `"flags":1`)
			case "unknown-flags":
				payload = strings.ReplaceAll(payload, `"flags":0`, `"flags":8`)
			case "count-mismatch":
				payload = strings.ReplaceAll(payload, `"count":"3"`, `"count":"4"`)
			case "count-only":
				payload = strings.ReplaceAll(strings.ReplaceAll(payload, `"bounds":["1"]`, `"bounds":[]`), `"bucket_counts":["1","2"]`, `"bucket_counts":[]`)
			case "all-values":
				payload = strings.ReplaceAll(strings.ReplaceAll(payload, `"bounds":["1"]`, `"bounds":[]`), `"bucket_counts":["1","2"]`, `"bucket_counts":["3"]`)
			case "nonfinite-stat":
				payload = strings.ReplaceAll(payload, `"sum":"0"`, `"sum":"nan"`)
			case "missing-stat-columns":
				payload = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(payload, `"sum":"0"`, `"sum":null`), `"min":"0"`, `"min":null`), `"max":"0"`, `"max":null`)
			case "rows-budget":
				payload = strings.Repeat(payload, 10001)
			}

		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, source)
	body := c.Body
	if c.PadTo > len(body) {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	body += c.Suffix
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	if w.Code == 200 {
		var value map[string]any
		if json.Unmarshal(w.Body.Bytes(), &value) != nil {
			t.Fatal("wire result")
		}
		value["readAt"] = "2026-11-14T00:00:00Z"
		fixed, _ := json.Marshal(value)
		w.Body.Reset()
		w.Body.Write(append(fixed, '\n'))
	}
	return w, reads
}
func TestRawHistogram447HTTPWire(t *testing.T) {
	cases := []planningWireCase{
		{Name: "observed", Method: "POST", Path: "/histogram-observations", Body: rawHistogram447HTTPBody, Status: 200},
		{Name: "unknown-temporality", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `"temporality":1`, `"temporality":7`), Status: 200},
		{Name: "missing-temporality", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `,"temporality":1`, ``), Status: 400},
		{Name: "null-temporality", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `"temporality":1`, `"temporality":null`), Status: 400},
		{Name: "wrong-kind", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `"histogram"`, `"exponential-histogram"`), Status: 400},
		{Name: "missing-unit", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `,"unit":"s"`, ``), Status: 400},
		{Name: "wrongcase", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `"scope"`, `"Scope"`), Status: 400},
		{Name: "all-scope", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `"environment":"prod"`, `"environmentSelector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"all"}}`), Status: 400},
		{Name: "trailing", Method: "POST", Path: "/histogram-observations", Body: rawHistogram447HTTPBody, Suffix: `{}`, Status: 400},
		{Name: "unknown", Method: "POST", Path: "/histogram-observations", Body: strings.ReplaceAll(rawHistogram447HTTPBody, `"kind":"histogram"`, `"kind":"histogram","extra":true`), Status: 400},
		{Name: "oversize", Method: "POST", Path: "/histogram-observations", Body: rawHistogram447HTTPBody, PadTo: 65537, Status: 400},
		{Name: "method", Method: "GET", Path: "/histogram-observations", Status: 405},
	}
	for _, name := range []string{"metadata-mismatch", "source-failure", "malformed-end", "rows-budget"} {
		cases = append(cases, planningWireCase{Name: name, Method: "POST", Path: "/histogram-observations", Body: rawHistogram447HTTPBody, Status: 502})
	}
	for _, name := range []string{"empty", "no-recorded", "unknown-flags", "count-mismatch", "count-only", "all-values", "nonfinite-stat", "missing-stat-columns"} {
		cases = append(cases, planningWireCase{Name: name, Method: "POST", Path: "/histogram-observations", Body: rawHistogram447HTTPBody, Status: 200})
	}

	bank := planningWireFixtures{Version: 1, Cases: cases}
	for index, c := range bank.Cases {
		w, reads := rawHistogram447HTTPResponse(t, c)
		wantReads := 0
		if c.Status == 200 || c.Status == 502 {
			wantReads = 3
		}
		if reads != wantReads || w.Code != c.Status {
			t.Fatalf("%s: code %d reads %d body %s", c.Name, w.Code, reads, w.Body.String())
		}
		bank.Cases[index].Headers = map[string]string{}
		for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
			bank.Cases[index].Headers[key] = w.Header().Get(key)
		}
		bank.Cases[index].Response = w.Body.String()
	}
	raw, err := json.MarshalIndent(bank, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	path := "../../testdata/http/raw-histogram447.json"
	if os.Getenv("GOO_UPDATE_RAW_HISTOGRAM447_HTTP_FIXTURES") == "1" {
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil || string(raw) != string(want) {
		t.Fatal("raw sum wire fixture changed", err)
	}
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, measurements.Source{})
	budget := queryhttp.NewBudget(mux, 0, func(r *http.Request) bool {
		p, ok := nativehttp.PolicyFor(r.Method, r.URL.Path)
		return ok && p.Effect != queryhttp.Pure
	})
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/histogram-observations", strings.NewReader(rawHistogram447HTTPBody)))
	if w.Code != 429 {
		t.Fatal("source admission bypass", w.Code)
	}
}
