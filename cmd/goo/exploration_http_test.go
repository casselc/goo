package main

import (
	"bytes"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestExplorationHTTPValidationAndSourceFailure(t *testing.T) {
	var reads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "count()") {
			fmt.Fprintln(w, `{"count":"0"}`)
		}
	}))
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()})
	call := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/explore", bytes.NewBufferString(body)))
		return w
	}
	for _, body := range []string{`{"signal":"spans"}`, `{"signal":"spans","environment":"prod","fromMs":1,"toMs":60001,"limit":100,"sql":"SELECT 1"}`, `{"signal":"spans","environment":"prod","fromMs":1,"toMs":60001,"limit":100,"where":{"field":{"kind":"column","key":"duration_ns"},"op":"gte","value":"1.5"}}`} {
		if w := call(body); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if reads.Load() != 0 {
		t.Fatal("invalid query reached source")
	}
	body := `{"signal":"spans","environment":"prod","fromMs":1,"toMs":60001,"limit":100,"offset":0}`
	if w := call(body); w.Code != 200 || !strings.Contains(w.Body.String(), `"state":"no_data"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	source.Close()
	if w := call(body); w.Code != 502 || strings.Contains(w.Body.String(), `"count":"0"`) {
		t.Fatal("failure became zero", w.Code, w.Body.String())
	}
}

func TestCohortComparisonHTTPRejectsDriftAndUnknownPayloads(t *testing.T) {
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Client: http.DefaultClient})
	cohort := `{"question":{"signal":"spans","environment":"local","fromMs":1,"toMs":60001,"limit":100,"groupBy":[{"kind":"attribute","key":"region"}]},"basis":"client-capture","count":"2","state":"observed","groups":[{"g0":"east","count":"2"}]}`
	valid := `{"baseline":` + cohort + `,"selection":` + cohort + `}`
	for _, tc := range []struct {
		body string
		code int
	}{{valid, 200}, {strings.Replace(valid, `"toMs":60001`, `"toMs":60002`, 1), 200}, {strings.Replace(valid, `"environment":"local"`, `"environment":"other"`, 1), 400}, {valid + ` {}`, 400}, {strings.Replace(valid, `"baseline":`, `"unknown":true,"baseline":`, 1), 400}, {strings.Replace(valid, `"count":"2"`, `"count":"02"`, 1), 400}} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/compare", strings.NewReader(tc.body)))
		if w.Code != tc.code {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("missing no-store")
		}
	}
}

func TestInventoryHTTPStrictScopeAndUnavailable(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	mux := http.NewServeMux()
	registerExploration(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()})
	call := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/inventory", strings.NewReader(body)))
		return w
	}
	for _, body := range []string{`{}`, `{"environment":"prod","fromMs":1,"toMs":60001,"limit":1}`, `{"environment":"prod","fromMs":1,"toMs":60001} {}`, `{"environment":"prod","fromMs":1,"toMs":99999999}`} {
		if w := call(body); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	valid := `{"environment":"prod","fromMs":1,"toMs":60001}`
	if w := call(valid); w.Code != 200 || !strings.Contains(w.Body.String(), `"state":"no_data"`) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Code, w.Body.String())
	}
	source.Close()
	if w := call(valid); w.Code != 502 || strings.Contains(w.Body.String(), `"count":"0"`) {
		t.Fatal("unavailable became no data", w.Code, w.Body.String())
	}
}
