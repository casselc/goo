package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func tracePage403(t *testing.T) exploration.TraceSearchPageRequest {
	t.Helper()
	var request exploration.TraceSearchRequest
	if err := json.Unmarshal([]byte(traceSearchHTTPBody), &request); err != nil {
		t.Fatal(err)
	}
	position, err := exploration.NextTraceSearchPosition(request, "11111111111111111111111111111111")
	if err != nil {
		t.Fatal(err)
	}
	return exploration.TraceSearchPageRequest{Request: request, Position: *position}
}

func TestTraceSearchPage403HTTPValidationBeforeSource(t *testing.T) {
	page := tracePage403(t)
	raw, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, method, body string
		status             int
	}{
		{"valid source failure", "POST", string(raw), 502},
		{"exact envelope bound", "POST", string(raw) + strings.Repeat(" ", 66560-len(raw)), 502},
		{"oversize", "POST", string(raw) + strings.Repeat(" ", 66561-len(raw)), 400},
		{"missing request", "POST", `{"position":{}}`, 400},
		{"null request", "POST", `{"request":null,"position":{}}`, 400},
		{"missing position", "POST", `{"request":` + traceSearchHTTPBody + `}`, 400},
		{"null position", "POST", `{"request":` + traceSearchHTTPBody + `,"position":null}`, 400},
		{"unknown envelope", "POST", strings.TrimSuffix(string(raw), "}") + `,"unexpected":true}`, 400},
		{"trailing value", "POST", string(raw) + ` {}`, 400},
		{"method", "GET", string(raw), 405},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reads := 0
			source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
				reads++
				return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable")), Header: make(http.Header)}, nil
			})}}
			mux := http.NewServeMux()
			registerExploration(mux, source, &exploration.ResultSnapshots{}, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(c.method, "/trace-search/page", strings.NewReader(c.body)))
			wantReads := 0
			if c.status == 502 {
				wantReads = 1
			}
			if response.Code != c.status || reads != wantReads {
				t.Fatalf("status=%d reads=%d body=%s", response.Code, reads, response.Body.String())
			}
		})
	}
	for _, change := range []func(*exploration.TraceSearchPageRequest){
		func(p *exploration.TraceSearchPageRequest) { p.Request.Limit++ },
		func(p *exploration.TraceSearchPageRequest) { p.Request.Environment = "other" },
		func(p *exploration.TraceSearchPageRequest) { p.Request.ToMS++ },
	} {
		bad := page
		change(&bad)
		body, _ := json.Marshal(bad)
		w := traceSearchWireResponse(t, planningWireCase{Method: "POST", Path: "/trace-search/page", Body: string(body), Status: 400})
		if w.Code != 400 {
			t.Fatal("scope position mismatch accepted", w.Code, w.Body.String())
		}
	}
}

func TestTraceSearchPage403CancellationUsesSharedAdmission(t *testing.T) {
	called := false
	budget := newQueryBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), 1)
	id := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	cancel := httptest.NewRecorder()
	budget.ServeHTTP(cancel, httptest.NewRequest("POST", "/queries/"+id+"/cancel", nil))
	raw, _ := json.Marshal(tracePage403(t))
	request := httptest.NewRequest("POST", "/trace-search/page", strings.NewReader(string(raw)))
	request.Header.Set("X-Goo-Query-ID", id)
	response := httptest.NewRecorder()
	budget.ServeHTTP(response, request)
	if cancel.Code != 202 || response.Code != 409 || called {
		t.Fatal("cancelled continuation admitted", cancel.Code, response.Code, called)
	}
}

func TestTraceSearchPage403EmptyPageKeepsFullCountAndPosition(t *testing.T) {
	page := tracePage403(t)
	raw, _ := json.Marshal(page)
	reads := 0
	source := exploration.Source{Endpoint: "http://source", Client: &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		if r.URL.Query().Get("readonly") != "1" {
			t.Error("continuation is not readonly")
		}
		if r.URL.Query().Get("param_trace_after") != "11111111111111111111111111111111" {
			t.Error("continuation cursor parameter missing")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"matched_traces":"2","remaining_traces":"0","page_present":0,"trace_id":"","c0":"","w0":["","","","","","","",[]]}` + "\n")), Header: make(http.Header)}, nil
	})}}
	mux := http.NewServeMux()
	registerExploration(mux, source, &exploration.ResultSnapshots{}, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/trace-search/page", strings.NewReader(string(raw))))
	var result exploration.TraceSearchResult
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || reads != 1 || result.State != "no_data" || result.MatchedTraces != "2" || result.RemainingTraces != "0" || result.HasMore || result.NextPosition != nil || result.Position == nil || *result.Position != page.Position || len(result.Items) != 0 {
		t.Fatal("empty continuation lost population or identity", w.Code, w.Body.String(), reads)
	}
}
