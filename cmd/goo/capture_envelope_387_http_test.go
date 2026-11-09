package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
)

func TestCaptureEnvelope387HTTPRefusesBeforeSource(t *testing.T) {
	var reads atomic.Int32
	backend := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reads.Add(1) }))
	defer backend.Close()
	mux := http.NewServeMux()
	capturehttp.Register(mux, exploration.Source{Endpoint: backend.URL}, t.TempDir(), &exploration.ResultSnapshots{})
	base := `{"kind":"explore","explore":{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":100}}`
	questions := []string{"null", "{}", `{"kind":"explore"}`, strings.Replace(base, `"explore":{`, `"explore":null,"ignored":{`, 1)}
	for _, key := range []string{"traceSearch", "explanation", "measurement", "comparison", "inventory"} {
		questions = append(questions, strings.TrimSuffix(base, "}")+`,"`+key+`":null}`, strings.TrimSuffix(base, "}")+`,"`+key+`":{}}`)
	}
	for _, question := range questions {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/captures", strings.NewReader(`{"title":"Invalid mixed capture","question":`+question+`}`)))
		if w.Code != 400 || reads.Load() != 0 {
			t.Fatalf("question=%s status=%d reads=%d body=%s", question, w.Code, reads.Load(), w.Body.String())
		}
	}
}
