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

const traceSearchHTTPBody = `{"version":"goo/trace-search/v1","environment":"prod","fromMs":1000,"toMs":61000,"conditions":[{"field":{"kind":"column","key":"service"},"op":"eq","value":"db"}],"limit":20}`

func traceSearchWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	resultSnapshots := &exploration.ResultSnapshots{}
	t.Helper()
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		deadline, ok := r.Context().Deadline()
		remaining := time.Until(deadline)
		if !ok || remaining > 55*time.Second || remaining < 50*time.Second {
			t.Fatal("trace search source deadline", remaining)
		}
		if r.URL.Query().Get("readonly") != "1" {
			t.Fatal("trace search source not readonly")
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
	if reads != want {
		t.Fatal("unexpected source effects", reads, want)
	}
	return response
}
func TestTraceSearchHTTPWireFixtures(t *testing.T) {
	cases := []planningWireCase{
		{Name: "empty", Method: "POST", Path: "/trace-search", Status: 400},
		{Name: "unknown", Method: "POST", Path: "/trace-search", Body: `{"unexpected":true}`, Status: 400},
		{Name: "trailing", Method: "POST", Path: "/trace-search", Body: traceSearchHTTPBody + ` {}`, Status: 400},
		{Name: "missing version", Method: "POST", Path: "/trace-search", Body: `{}`, Status: 400},
		{Name: "source unavailable", Method: "POST", Path: "/trace-search", Body: traceSearchHTTPBody, Status: 502},
		{Name: "exact size", Method: "POST", Path: "/trace-search", Body: traceSearchHTTPBody, PadTo: 65536, Status: 502},
		{Name: "oversize", Method: "POST", Path: "/trace-search", Body: traceSearchHTTPBody, PadTo: 65537, Status: 400},
		{Name: "method", Method: "GET", Path: "/trace-search", Status: 405},
	}
	const filename = "../../testdata/http/trace-search.json"
	if os.Getenv("GOO_UPDATE_TRACE_SEARCH_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := traceSearchWireResponse(t, cases[i])
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
	if frozen.Version != 1 || len(frozen.Cases) != 8 {
		t.Fatal("trace search contract coverage changed")
	}
	for _, c := range frozen.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := traceSearchWireResponse(t, c)
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
func TestTraceSearchHTTPSuccessIsSingleReadWithoutRetention(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	reads := 0
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	request := httptest.NewRequest("POST", "/trace-search", strings.NewReader(traceSearchHTTPBody))
	request.Header.Set("Content-Type", "text/plain")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	var result exploration.TraceSearchResult
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil || result.State != "no_data" || result.MatchedTraces != "0" || result.Items == nil || reads != 1 {
		t.Fatal(response.Code, response.Body.String(), reads)
	}
	var object map[string]json.RawMessage
	_ = json.Unmarshal(response.Body.Bytes(), &object)
	if _, ok := object["resultToken"]; ok {
		t.Fatal("trace search unexpectedly issued retention membership")
	}
}
func TestTraceSearchUsesSharedCancellationAdmission(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	cancel := httptest.NewRecorder()
	budget.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	request := httptest.NewRequest("POST", "/trace-search", strings.NewReader(traceSearchHTTPBody))
	request.Header.Set("X-Goo-Query-ID", id)
	response := httptest.NewRecorder()
	budget.ServeHTTP(response, request)
	if cancel.Code != 202 || response.Code != 409 || called {
		t.Fatal("cancelled trace search admitted", cancel.Code, response.Code, called)
	}
}

func TestTraceSearchHTTPPredicateBudgetRefusedBeforeSource(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	var request exploration.TraceSearchRequest
	if err := json.Unmarshal([]byte(traceSearchHTTPBody), &request); err != nil {
		t.Fatal(err)
	}
	group := exploration.Predicate{}
	for i := 0; i < 31; i++ {
		binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "amount"}, Type: "decimal64", Scale: 9, Unit: "USD", Parse: "canonical-decimal"}
		group.All = append(group.All, exploration.Predicate{Numeric: &binding, Op: "valid"})
	}
	request.Conditions = []exploration.Predicate{group, group}
	if err := request.Validate(); err != nil {
		t.Fatal("maximum-node fixture must pass structural validation", err)
	}
	plan, err := exploration.CompileTraceSearch(request)
	if err != nil || len(plan.SQL) > 256<<10 {
		t.Fatal("maximum-node fixture violated compiled bound", len(plan.SQL), err)
	}
	request.Conditions = append(request.Conditions, exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "db"})
	raw, err := json.Marshal(request)
	if err != nil || len(raw) > 65536 {
		t.Fatal("invalid request fixture", err, len(raw))
	}
	called := false
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, resultSnapshots, nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest("POST", "/trace-search", strings.NewReader(string(raw))))
	if response.Code != 400 || called || !strings.Contains(response.Body.String(), "exceeds 64 predicate nodes") {
		t.Fatal("oversize predicate tree reached source", response.Code, called)
	}
}
