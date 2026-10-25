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

const countedValuesHTTPBody = `{"signal":"logs","environment":"prod","fromMs":1000,"toMs":61000,"limit":20,"groupBy":[{"kind":"column","key":"severity"}]}`

func countedValuesWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		deadline, ok := r.Context().Deadline()
		remaining := time.Until(deadline)
		if !ok || remaining > 55*time.Second || remaining < 50*time.Second {
			t.Fatal("counted values source deadline", remaining)
		}
		if r.URL.Query().Get("readonly") != "1" {
			t.Fatal("counted values source not readonly")
		}

		if strings.HasPrefix(c.Name, "success-") {
			payload := `{"count":"2"}` + "\n"
			if c.Name == "success-no-data" {
				payload = `{"count":"0"}` + "\n"
			}
			if reads == 2 {
				payload = `{"g0":"ERROR","count":"2"}` + "\n"
				if c.Name == "success-changed" {
					payload = `{"g0":"ERROR","count":"1"}` + "\n"
				}
				if c.Name == "success-no-data" {
					payload = ""
				}
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("fixture unavailable")), Header: make(http.Header)}, nil
	})}
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Endpoint: "http://source", Client: client})
	body := c.Body + strings.Repeat("x", c.FillBytes) + c.Suffix
	if c.PadTo > 0 {
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	request := httptest.NewRequest(c.Method, c.Path, strings.NewReader(body))
	if c.ContentType != "" {
		request.Header.Set("Content-Type", c.ContentType)
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	want := 0
	if c.Status == 502 {
		want = 1
	}
	if c.Status == 200 {
		want = 2
	}
	if reads != want {
		t.Fatal("unexpected source effects", reads, want)
	}

	if c.Status == 200 {
		var result map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		expectedState := map[string]string{"success-observed": "observed", "success-no-data": "no_data", "success-changed": "changed_between_reads"}[c.Name]
		if result["version"] != "goo/counted-values/v1" || result["state"] != expectedState {
			t.Fatal("unexpected counted state", result)
		}
		if _, ok := result["resultToken"]; ok {
			t.Fatal("counted values issued retention")
		}
		// Clocks are checked at the real handler boundary, then normalized only for
		// this frozen fixture. Content/read bases remain unmodified and deterministic.
		clock := func(obj map[string]any, key string) {
			value, ok := obj[key].(string)
			if !ok {
				t.Fatal("missing clock", key)
			}
			if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
				t.Fatal(err)
			}
			obj[key] = "2026-10-25T00:00:00Z"
		}
		clock(result, "capturedAt")
		for _, key := range []string{"count", "groups"} {
			clock(result["reads"].(map[string]any)[key].(map[string]any), "readAt")
		}
		raw, _ := json.Marshal(result)
		response.Body.Reset()
		response.Body.Write(append(raw, '\n'))
	}
	return response
}
func TestCountedValuesHTTPWireFixtures(t *testing.T) {
	cases := []planningWireCase{
		{Name: "empty", Method: "POST", Path: "/explore/counted-values", Status: 400},
		{Name: "unknown", Method: "POST", Path: "/explore/counted-values", Body: `{"unexpected":true}`, Status: 400},
		{Name: "trailing", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody + ` {}`, Status: 400},
		{Name: "missing grouping", Method: "POST", Path: "/explore/counted-values", Body: `{}`, Status: 400},
		{Name: "source unavailable", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody, Status: 502},
		{Name: "exact size", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody, PadTo: 16384, Status: 502},
		{Name: "oversize", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody, PadTo: 16385, Status: 400},
		{Name: "success-observed", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody, Status: 200},
		{Name: "success-no-data", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody, Status: 200},
		{Name: "success-changed", Method: "POST", Path: "/explore/counted-values", Body: countedValuesHTTPBody, Status: 200},
		{Name: "method", Method: "GET", Path: "/explore/counted-values", Status: 405},
	}
	const filename = "../../testdata/http/counted-values.json"
	if os.Getenv("GOO_UPDATE_COUNTED_VALUES_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := countedValuesWireResponse(t, cases[i])
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[key] = w.Header().Get(key)
			}
		}
		out, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		out = append(out, '\n')
		if err := os.WriteFile(filename, out, 0644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var frozen planningWireFixtures
	if err = json.Unmarshal(raw, &frozen); err != nil {
		t.Fatal(err)
	}
	if frozen.Version != 1 || len(frozen.Cases) != 11 {
		t.Fatal("counted values contract coverage changed")
	}
	for _, c := range frozen.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := countedValuesWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatalf("wire changed %d %s", w.Code, w.Body.String())
			}
			for _, key := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				if w.Header().Get(key) != c.Headers[key] {
					t.Fatalf("header %s changed", key)
				}
			}
		})
	}
}

func TestCountedValuesUsesSharedCancellationAdmission(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	cancel := httptest.NewRecorder()
	budget.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	request := httptest.NewRequest("POST", "/explore/counted-values", strings.NewReader(countedValuesHTTPBody))
	request.Header.Set("X-Goo-Query-ID", id)
	response := httptest.NewRecorder()
	budget.ServeHTTP(response, request)
	if cancel.Code != 202 || response.Code != 409 || called {
		t.Fatal("cancelled counted values admitted", cancel.Code, response.Code, called)
	}
}
func TestCountedValuesHTTPRejectsUnsupportedModesBeforeSource(t *testing.T) {
	reads := 0
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("unexpected")), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source)
	for _, extra := range []string{`,"offset":1`, `,"bucketMs":1000`, `,"summary":"span-duration"`, `,"spanStatus":true`, `,"groupTimeSeries":true`} {
		body := strings.TrimSuffix(countedValuesHTTPBody, "}") + extra + "}"
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/explore/counted-values", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal("unsupported counted mode", extra, w.Code, w.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("invalid counted mode read source", reads)
	}
}
