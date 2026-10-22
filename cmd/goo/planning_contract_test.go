package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type planningWireCase struct {
	FillBytes   int               `json:"fillBytes,omitempty"`
	Suffix      string            `json:"suffix,omitempty"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Body        string            `json:"body"`
	PadTo       int               `json:"padTo,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
	Status      int               `json:"status"`
	Headers     map[string]string `json:"headers"`
	Response    string            `json:"response"`
}
type planningWireFixtures struct {
	Version int                `json:"version"`
	Cases   []planningWireCase `json:"cases"`
}

func planningWireResponse(t *testing.T, c planningWireCase) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	registerRecipes(mux)
	registerDashboards(mux, nil)
	body := c.Body + strings.Repeat("x", c.FillBytes) + c.Suffix
	if c.PadTo > 0 {
		if c.PadTo < len(body) {
			t.Fatal("invalid padding fixture")
		}
		body += strings.Repeat(" ", c.PadTo-len(body))
	}
	r := httptest.NewRequest(c.Method, c.Path, strings.NewReader(body))
	if c.ContentType != "" {
		r.Header.Set("Content-Type", c.ContentType)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}
func TestPlanningHTTPWireFixtures(t *testing.T) {
	b, e := os.ReadFile("../../testdata/http/planning.json")
	if e != nil {
		t.Fatal(e)
	}
	var f planningWireFixtures
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	if f.Version != 1 || len(f.Cases) != 32 {
		t.Fatal("coverage changed", len(f.Cases))
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := planningWireResponse(t, c)
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
