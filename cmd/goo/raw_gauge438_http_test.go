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

const rawGauge438HTTPBody = `{"question":{"environment":"prod","service":"api","fromMs":1000,"toMs":2000},"instrument":{"kind":"gauge","scope":"","version":"","schemaURL":"","metric":"temperature","unit":"Cel"}}`
const rawGauge438HTTPRow = `{"resource":{"service.name":"api","deployment.environment.name":"prod"},"attributes":{},"resource_schema_url":"","scope":"","version":"","schema_url":"","name":"temperature","unit":"Cel","start_ns":"0","end_ns":"1500000000","value":"12.5","flags":0}`

func rawGauge438HTTPResponse(t *testing.T, c planningWireCase) (*httptest.ResponseRecorder, int) {
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
		case strings.HasPrefix(string(sql), "DESCRIBE TABLE otel_metrics_gauge "):
			for _, row := range columns["gauge"] {
				compact, _ := json.Marshal(row)
				payload += string(compact) + "\n"
			}
		case strings.HasSuffix(string(sql), " LIMIT 0 FORMAT JSONEachRow"):
		default:
			payload = rawGauge438HTTPRow + "\n"
			if c.Name == "empty" {
				payload = ""
			}
			if c.Name == "source-failure" {
				status = 403
				payload = "not authorized"
			}
			if c.Name == "malformed-end" {
				payload = strings.ReplaceAll(payload, `"1500000000"`, `"2000000000"`)
			}
			if c.Name == "no-recorded" {
				payload = strings.ReplaceAll(payload, `"flags":0`, `"flags":1`)
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
func TestRawGauge438HTTPWire(t *testing.T) {
	cases := []planningWireCase{
		{Name: "observed", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, Status: 200},
		{Name: "empty", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, Status: 200},
		{Name: "no-recorded", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, Status: 200},
		{Name: "source-failure", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, Status: 502},
		{Name: "malformed-end", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, Status: 502},
		{Name: "wrong-kind", Method: "POST", Path: "/metric-observations", Body: strings.ReplaceAll(rawGauge438HTTPBody, `"gauge"`, `"counter"`), Status: 400},
		{Name: "missing-unit", Method: "POST", Path: "/metric-observations", Body: strings.ReplaceAll(rawGauge438HTTPBody, `,"unit":"Cel"`, ``), Status: 400},
		{Name: "trailing", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, Suffix: `{}`, Status: 400},
		{Name: "unknown", Method: "POST", Path: "/metric-observations", Body: strings.ReplaceAll(rawGauge438HTTPBody, `"kind":"gauge"`, `"kind":"gauge","extra":true`), Status: 400},
		{Name: "oversize", Method: "POST", Path: "/metric-observations", Body: rawGauge438HTTPBody, PadTo: 65537, Status: 400},
		{Name: "method", Method: "GET", Path: "/metric-observations", Body: "", Status: 405},
	}
	bank := planningWireFixtures{Version: 1, Cases: cases}
	for index, c := range bank.Cases {
		w, reads := rawGauge438HTTPResponse(t, c)
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
	path := "../../testdata/http/raw-gauge438.json"
	if os.Getenv("GOO_UPDATE_RAW_GAUGE438_HTTP_FIXTURES") == "1" {
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil || string(raw) != string(want) {
		t.Fatal("raw gauge wire fixture changed", err)
	}
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, measurements.Source{})
	budget := queryhttp.NewBudget(mux, 0, func(r *http.Request) bool {
		p, ok := nativehttp.PolicyFor(r.Method, r.URL.Path)
		return ok && p.Effect != queryhttp.Pure
	})
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/metric-observations", strings.NewReader(rawGauge438HTTPBody)))
	if w.Code != 429 {
		t.Fatal("source admission bypass", w.Code)
	}
}
