package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

// Freeze actual reset/gap frames, not synthetic event descriptions. A failing
// source ends the stream immediately without a timer or database dependency.
func TestAPITailSSEWireContract(t *testing.T) {
	client := &http.Client{Transport: nativeDependencyTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable")), Header: make(http.Header)}, nil
	})}
	mux := http.NewServeMux()
	registerTail(mux, exploration.Source{Endpoint: "http://source", Client: client})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/tail", strings.NewReader(`{"environment":"prod"}`)))
	want, err := os.ReadFile("../../testdata/http/tail.sse")
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || w.Body.String() != string(want) {
		t.Fatal(w.Code, w.Body.String())
	}
	for k, v := range map[string]string{"Content-Type": "text/event-stream", "Cache-Control": "no-store", "X-Accel-Buffering": "no"} {
		if w.Header().Get(k) != v {
			t.Fatal(k, w.Header().Get(k))
		}
	}
}
