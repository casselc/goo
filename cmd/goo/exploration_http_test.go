package main

import (
	"bytes"
	"fmt"
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
	t.Setenv("CLICKHOUSE_URL", source.URL)
	mux := http.NewServeMux()
	registerExploration(mux, source.Client())
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
