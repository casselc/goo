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

const rawSum441HTTPBody = `{"question":{"environment":"prod","service":"api","fromMs":1000,"toMs":2000},"instrument":{"kind":"counter","scope":"","version":"","schemaURL":"","metric":"jobs.total","unit":"{job}","temporality":1,"monotonic":true}}`
const rawSum441HTTPRow = `{"resource":{"service.name":"api","deployment.environment.name":"prod"},"attributes":{},"resource_schema_url":"","scope":"","version":"","schema_url":"","name":"jobs.total","unit":"{job}","start_ns":"0","end_ns":"1500000000","value":"12.5","flags":0,"temporality":1,"monotonic":true}`

func rawSum441HTTPResponse(t *testing.T, c planningWireCase) (*httptest.ResponseRecorder, int) {
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
		case strings.HasPrefix(string(sql), "DESCRIBE TABLE otel_metrics_sum "):
			for _, row := range columns["counter"] {
				compact, _ := json.Marshal(row)
				payload += string(compact) + "\n"
			}
		case strings.HasSuffix(string(sql), " LIMIT 0 FORMAT JSONEachRow"):
		default:
			payload = rawSum441HTTPRow + "\n"
			if c.Name == "stock" {
				payload = strings.ReplaceAll(strings.ReplaceAll(payload, `"temporality":1`, `"temporality":2`), `"monotonic":true`, `"monotonic":false`)
			}
			if c.Name == "unknown-temporality" {
				payload = strings.ReplaceAll(payload, `"temporality":1`, `"temporality":7`)
			}
			if c.Name == "metadata-mismatch" {
				payload = strings.ReplaceAll(payload, `"temporality":1`, `"temporality":2`)
			}
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
func TestRawSum441HTTPWire(t *testing.T) {
	cases := []planningWireCase{
		{Name: "observed", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Status: 200},
		{Name: "stock", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(rawSum441HTTPBody, `"counter"`, `"stock"`), `"temporality":1`, `"temporality":2`), `"monotonic":true`, `"monotonic":false`), Status: 200},
		{Name: "unknown-temporality", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `"temporality":1`, `"temporality":7`), Status: 200},
		{Name: "missing-monotonic", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `,"monotonic":true`, ``), Status: 400},
		{Name: "null-temporality", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `"temporality":1`, `"temporality":null`), Status: 400},
		{Name: "mislabeled-stock", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `"counter"`, `"stock"`), Status: 400},
		{Name: "mislabeled-counter", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(strings.ReplaceAll(rawSum441HTTPBody, `"temporality":1`, `"temporality":2`), `"monotonic":true`, `"monotonic":false`), Status: 400},
		{Name: "all-scope", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `"environment":"prod"`, `"environmentSelector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"all"}}`), Status: 400},
		{Name: "metadata-mismatch", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Status: 502},
		{Name: "empty", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Status: 200},
		{Name: "no-recorded", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Status: 200},
		{Name: "source-failure", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Status: 502},
		{Name: "malformed-end", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Status: 502},
		{Name: "wrong-kind", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `"counter"`, `"gauge"`), Status: 400},
		{Name: "missing-unit", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `,"unit":"{job}"`, ``), Status: 400},
		{Name: "trailing", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, Suffix: `{}`, Status: 400},
		{Name: "unknown", Method: "POST", Path: "/sum-observations", Body: strings.ReplaceAll(rawSum441HTTPBody, `"kind":"counter"`, `"kind":"counter","extra":true`), Status: 400},
		{Name: "oversize", Method: "POST", Path: "/sum-observations", Body: rawSum441HTTPBody, PadTo: 65537, Status: 400},
		{Name: "method", Method: "GET", Path: "/sum-observations", Body: "", Status: 405},
	}
	bank := planningWireFixtures{Version: 1, Cases: cases}
	for index, c := range bank.Cases {
		w, reads := rawSum441HTTPResponse(t, c)
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
	path := "../../testdata/http/raw-sum441.json"
	if os.Getenv("GOO_UPDATE_RAW_SUM441_HTTP_FIXTURES") == "1" {
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
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/sum-observations", strings.NewReader(rawSum441HTTPBody)))
	if w.Code != 429 {
		t.Fatal("source admission bypass", w.Code)
	}
}
