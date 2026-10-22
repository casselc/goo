package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func telemetryWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerTelemetryPackages(mux, dir)
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
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	shouldPublish := c.Path == "/telemetry-packages" && c.Method == "POST" && w.Code == 201
	if shouldPublish && len(entries) != 1 {
		t.Fatal("publication missing")
	}
	if !shouldPublish && len(entries) != 0 {
		t.Fatal("planning or rejected input wrote files")
	}
	return w
}

func TestTelemetryHTTPWireFixtures(t *testing.T) {
	b, err := os.ReadFile("../../testdata/http/telemetry.json")
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
			w := telemetryWireResponse(t, c)
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
