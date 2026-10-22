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

func explorationJSONWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	reads := 0
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		deadline, ok := r.Context().Deadline()
		want := map[string]time.Duration{"/inventory": 15 * time.Second, "/discovery": 30 * time.Second, "/explore": 55 * time.Second}[c.Path]
		remaining := time.Until(deadline)
		if !ok || remaining > want || remaining < want-5*time.Second {
			t.Fatal("source deadline changed", c.Path, remaining)
		}
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("fixture unavailable")), Header: make(http.Header)}, nil
	})}
	registerExploration(mux, exploration.Source{Endpoint: "http://source", Client: client})
	body := c.Body + strings.Repeat("x", c.FillBytes) + c.Suffix
	if c.PadTo > 0 {
		if c.PadTo < len(body) {
			t.Fatal("invalid fixture padding")
		}
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	r := httptest.NewRequest(c.Method, c.Path, strings.NewReader(body))
	if c.ContentType != "" {
		r.Header.Set("Content-Type", c.ContentType)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	wantReads := 0
	if c.Status == 502 {
		wantReads = 1
	}
	if reads != wantReads {
		t.Fatal("unexpected source execution", reads, wantReads)
	}
	return w
}

func TestExplorationJSONHTTPWireFixtures(t *testing.T) {
	b, err := os.ReadFile("../../testdata/http/exploration-json.json")
	if err != nil {
		t.Fatal(err)
	}
	var f planningWireFixtures
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Version != 1 || len(f.Cases) != 68 {
		t.Fatal("coverage changed", len(f.Cases))
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := explorationJSONWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatalf("wire changed: status %d body %s", w.Code, w.Body.String())
			}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				if w.Header().Get(k) != c.Headers[k] {
					t.Fatalf("header %s changed: %q", k, w.Header().Get(k))
				}
			}
		})
	}
}
