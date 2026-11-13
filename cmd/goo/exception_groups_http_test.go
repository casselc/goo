package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

const exceptionGroupsHTTPBody416 = `{"signal":"logs","environment":"prod","fromMs":1000,"toMs":61000,"limit":100,"offset":0}`

func exceptionGroupsWire416(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 55*time.Second || time.Until(deadline) < 50*time.Second || r.URL.Query().Get("readonly") != "1" {
			t.Fatal("exception read policy changed")
		}
		status, payload := 503, "unavailable"
		if strings.HasPrefix(c.Name, "success-") || c.Name == "malformed-groups" || c.Name == "late-groups-unavailable" {
			status, payload = 200, "{\"count\":\"2\"}\n"
			if c.Name == "success-no-data" {
				payload = "{\"count\":\"0\"}\n"
			}
			if c.Name == "success-changed" {
				payload = "{\"count\":\"3\"}\n"
			}
			if reads == 2 {
				if c.Name == "success-no-data" {
					payload = ""
				} else if c.Name == "late-groups-unavailable" {
					status, payload = 503, "unavailable"
				} else {
					identity := exploration.ReferenceLogExceptionFingerprint(map[string]string{"exception.type": "TypeError", "exception.stacktrace": "at worker:42"})
					value := map[string]any{"fingerprint": identity.Fingerprint, "count": "2", "first_ns": "1000000001", "last_ns": "1000000002", "type_min": "TypeError", "type_max": "TypeError", "stack_min": "at worker:42", "stack_max": "at worker:42", "affected": [][]any{{"worker", 0, "", 1, "v1"}}, "affected_count": "1"}
					if c.Name == "malformed-groups" {
						value["stack_max"] = "other"
					}
					raw, _ := json.Marshal(value)
					payload = string(raw) + "\n"
				}
			}
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Endpoint: "http://fixture.invalid", Client: client}, &exploration.ResultSnapshots{}, nil)
	body := c.Body + c.Suffix
	if c.PadTo > 0 {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path, strings.NewReader(body)))
	want := 0
	if c.Status == 502 {
		want = 1
	}
	if c.Status == 200 || c.Name == "malformed-groups" || c.Name == "late-groups-unavailable" {
		want = 2
	}
	if reads != want {
		t.Fatal("exception source effects", reads, want)
	}
	if w.Code == 200 {
		var result map[string]any
		if e := json.Unmarshal(w.Body.Bytes(), &result); e != nil {
			t.Fatal(e)
		}
		wantState := map[string]string{"success-observed": "observed", "success-no-data": "no_data", "success-changed": "changed_between_reads"}[c.Name]
		if result["version"] != exploration.ExceptionGroupsVersion || result["state"] != wantState {
			t.Fatal(result)
		}
		for _, key := range []string{"rows", "fields", "resultToken", "plan", "hasMore"} {
			if _, ok := result[key]; ok {
				t.Fatal("exception groups leaked inspection field", key)
			}
		}
		normalize := func(v map[string]any, key string) {
			s, ok := v[key].(string)
			if !ok {
				t.Fatal("missing clock", key)
			}
			if _, e := time.Parse(time.RFC3339Nano, s); e != nil {
				t.Fatal(e)
			}
			v[key] = "2026-11-13T00:00:00Z"
		}
		normalize(result, "capturedAt")
		for _, key := range []string{"count", "groups"} {
			normalize(result["reads"].(map[string]any)[key].(map[string]any), "readAt")
		}
		raw, _ := json.Marshal(result)
		w.Body.Reset()
		w.Body.Write(append(raw, '\n'))
	}
	return w
}
func TestExceptionGroupsHTTPWire416(t *testing.T) {
	cases := []planningWireCase{
		{Name: "empty", Method: "POST", Path: "/exception-groups", Status: 400},
		{Name: "unknown", Method: "POST", Path: "/exception-groups", Body: `{"extra":true}`, Status: 400},
		{Name: "trailing", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Suffix: ` {}`, Status: 400},
		{Name: "unsupported-projection", Method: "POST", Path: "/exception-groups", Body: strings.Replace(exceptionGroupsHTTPBody416, `"offset":0`, `"offset":0,"groupBy":[{"kind":"column","key":"service"}]`, 1), Status: 400},
		{Name: "source-unavailable", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Status: 502},
		{Name: "exact-size", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, PadTo: 16384, Status: 502},
		{Name: "oversize", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, PadTo: 16385, Status: 400},
		{Name: "success-observed", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Status: 200},
		{Name: "success-no-data", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Status: 200},
		{Name: "success-changed", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Status: 200},
		{Name: "late-groups-unavailable", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Status: 502},
		{Name: "malformed-groups", Method: "POST", Path: "/exception-groups", Body: exceptionGroupsHTTPBody416, Status: 502},
		{Name: "method", Method: "GET", Path: "/exception-groups", Status: 405},
	}
	const path = "../../testdata/http/exception-groups.json"
	if os.Getenv("GOO_UPDATE_EXCEPTION_GROUPS_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := exceptionGroupsWire416(t, cases[i])
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[key] = w.Header().Get(key)
			}
		}
		raw, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e := os.WriteFile(path, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var frozen planningWireFixtures
	if e = json.Unmarshal(raw, &frozen); e != nil || frozen.Version != 1 || len(frozen.Cases) != len(cases) {
		t.Fatal("exception HTTP coverage changed", e)
	}
	for _, c := range frozen.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := exceptionGroupsWire416(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatalf("HTTP wire changed: %d %s", w.Code, w.Body.String())
			}
			for key, value := range c.Headers {
				if w.Header().Get(key) != value {
					t.Fatal("HTTP header changed", key)
				}
			}
		})
	}
}
