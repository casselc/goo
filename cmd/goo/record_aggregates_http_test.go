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

const recordAggregatesRichHTTPBody = `{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":20,"groupBy":[{"kind":"column","key":"service"}],"summary":"span-duration","spanStatus":true,"bucketMs":10000}`

const recordAggregatesHTTPBody = `{"signal":"logs","environment":"prod","fromMs":1000,"toMs":61000,"limit":20,"groupBy":[{"kind":"column","key":"severity"}]}`

func recordAggregatesWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	resultSnapshots := &exploration.ResultSnapshots{}
	t.Helper()
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		deadline, ok := r.Context().Deadline()
		remaining := time.Until(deadline)
		if !ok || remaining > 55*time.Second || remaining < 50*time.Second {
			t.Fatal("record aggregates source deadline", remaining)
		}
		if r.URL.Query().Get("readonly") != "1" {
			t.Fatal("record aggregates source not readonly")
		}

		if c.Name == "late-groups-unavailable" && reads == 2 || c.Name == "late-series-unavailable" && reads == 3 {
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("fixture unavailable")), Header: make(http.Header)}, nil
		}
		if strings.HasPrefix(c.Name, "success-") || strings.HasPrefix(c.Name, "late-") {
			payload := `{"count":"2"}` + "\n"
			if c.Name == "success-rich-series" || c.Name == "late-series-unavailable" {
				aggregate := `"count":"2","sum_ns":"20","min_ns":"10","max_ns":"10","quantiles_ns":["10","10","10"],"status_error":"1","status_ok":"1","status_unset":"0","status_unknown":"0"`
				prefix := ""
				if reads == 2 {
					prefix = `"g0":"worker",`
				}
				if reads == 3 {
					prefix = `"from_ms":"1000",`
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{" + prefix + aggregate + "}\n")), Header: make(http.Header)}, nil
			}
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
	registerExploration(mux, exploration.Source{Endpoint: "http://source", Client: client}, resultSnapshots, nil)
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
	if c.Name == "success-rich-series" {
		want = 3
	}
	if c.Name == "late-groups-unavailable" {
		want = 2
	}
	if c.Name == "late-series-unavailable" {
		want = 3
	}
	if reads != want {
		t.Fatal("unexpected source effects", reads, want)
	}

	if c.Status == 200 {
		var result map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		expectedState := map[string]string{"success-observed": "observed", "success-no-data": "no_data", "success-changed": "changed_between_reads", "success-rich-series": "observed"}[c.Name]
		if result["version"] != "goo/record-aggregates/v1" || result["state"] != expectedState {
			t.Fatal("unexpected counted state", result)
		}
		for _, key := range []string{"resultToken", "rows", "fields", "plan", "hasMore"} {
			if _, ok := result[key]; ok {
				t.Fatal("record aggregates leaked full fetch field", key)
			}
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
		for _, key := range []string{"count", "groups", "timeSeries"} {
			if _, ok := result["reads"].(map[string]any)[key]; !ok {
				continue
			}
			clock(result["reads"].(map[string]any)[key].(map[string]any), "readAt")
		}
		raw, _ := json.Marshal(result)
		response.Body.Reset()
		response.Body.Write(append(raw, '\n'))
	}
	return response
}
func TestRecordAggregatesHTTPWireFixtures(t *testing.T) {
	cases := []planningWireCase{
		{Name: "empty", Method: "POST", Path: "/record-aggregates", Status: 400},
		{Name: "unknown", Method: "POST", Path: "/record-aggregates", Body: `{"unexpected":true}`, Status: 400},
		{Name: "trailing", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody + ` {}`, Status: 400},
		{Name: "missing grouping", Method: "POST", Path: "/record-aggregates", Body: `{}`, Status: 400},
		{Name: "source unavailable", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, Status: 502},
		{Name: "exact size", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, PadTo: 16384, Status: 502},
		{Name: "oversize", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, PadTo: 16385, Status: 400},
		{Name: "success-observed", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, Status: 200},
		{Name: "success-no-data", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, Status: 200},
		{Name: "success-changed", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, Status: 200},
		{Name: "success-rich-series", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesRichHTTPBody, Status: 200},
		{Name: "late-groups-unavailable", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesHTTPBody, Status: 502},
		{Name: "late-series-unavailable", Method: "POST", Path: "/record-aggregates", Body: recordAggregatesRichHTTPBody, Status: 502},
		{Name: "method", Method: "GET", Path: "/record-aggregates", Status: 405},
	}
	const filename = "../../testdata/http/record-aggregates.json"
	if os.Getenv("GOO_UPDATE_RECORD_AGGREGATES_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := recordAggregatesWireResponse(t, cases[i])
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
	if frozen.Version != 1 || len(frozen.Cases) != 14 {
		t.Fatal("record aggregates contract coverage changed")
	}
	for _, c := range frozen.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := recordAggregatesWireResponse(t, c)
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

func TestRecordAggregatesUsesSharedCancellationAdmission(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	cancel := httptest.NewRecorder()
	budget.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	request := httptest.NewRequest("POST", "/record-aggregates", strings.NewReader(recordAggregatesHTTPBody))
	request.Header.Set("X-Goo-Query-ID", id)
	response := httptest.NewRecorder()
	budget.ServeHTTP(response, request)
	if cancel.Code != 202 || response.Code != 409 || called {
		t.Fatal("cancelled record aggregates admitted", cancel.Code, response.Code, called)
	}
}
func TestRecordAggregatesHTTPRejectsUnsupportedModesBeforeSource(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	reads := 0
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("unexpected")), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	for _, extra := range []string{`,"offset":1`, `,"groupTimeSeries":true`, `,"summary":"other"`} {
		body := strings.TrimSuffix(recordAggregatesHTTPBody, "}") + extra + "}"
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/record-aggregates", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal("unsupported counted mode", extra, w.Code, w.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("invalid counted mode read source", reads)
	}
}

func TestRecordAggregatesInflightCancellationStopsSource(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	started := make(chan struct{}, 1)
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		started <- struct{}{}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	budget := newQueryBudget(mux, 1)
	id := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	req := httptest.NewRequest("POST", "/record-aggregates", strings.NewReader(recordAggregatesHTTPBody))
	req.Header.Set("X-Goo-Query-ID", id)
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { w := httptest.NewRecorder(); budget.ServeHTTP(w, req); done <- w }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("aggregate source did not start")
	}
	// Admission is shared with /explore, rather than an independent new pool.
	busy := httptest.NewRecorder()
	budget.ServeHTTP(busy, httptest.NewRequest("POST", "/explore", strings.NewReader(`{}`)))
	if busy.Code != 429 {
		t.Fatal("aggregate did not occupy shared query budget", busy.Code)
	}
	cancelled := httptest.NewRecorder()
	budget.ServeHTTP(cancelled, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	if cancelled.Code != 202 {
		t.Fatal(cancelled.Code, cancelled.Body.String())
	}
	select {
	case result := <-done:
		if result.Code != 502 {
			t.Fatal("cancel became successful data", result.Code, result.Body.String())
		}
	case <-time.After(time.Second):
		t.Fatal("source cancellation did not finish")
	}
}
