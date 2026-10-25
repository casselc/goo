package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
)

type runtimeWireCase struct {
	planningWireCase
	Profile    string `json:"profile,omitempty"`
	Prepared   bool   `json:"prepared,omitempty"`
	Operations bool   `json:"operations,omitempty"`
	Query      string `json:"query,omitempty"`
}
type runtimeWireFixtures struct {
	Version int               `json:"version"`
	Cases   []runtimeWireCase `json:"cases"`
}

func runtimeWireResponse(t *testing.T, c runtimeWireCase) *httptest.ResponseRecorder {
	t.Helper()
	var handler http.Handler
	if c.Path == "/capabilities" {
		mux := http.NewServeMux()
		registerRuntimeCapabilities(mux, c.Profile, c.Prepared, c.Operations)
		handler = mux
	} else {
		server, err := configuredHTTPServer([]string{"-profile", "generic", "-investigations-dir", t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		handler = server.Handler
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(c.Method, c.Path+c.Query, nil))
	return w
}
func TestRuntimeCapabilitiesHTTPWireFixtures(t *testing.T) {
	var sourceReads atomic.Int32
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceReads.Add(1)
		http.Error(w, "unexpected source read", 500)
	}))
	defer source.Close()
	t.Setenv("CLICKHOUSE_URL", source.URL)
	t.Cleanup(func() {
		if sourceReads.Load() != 0 {
			t.Fatal("runtime metadata/refusal read source", sourceReads.Load())
		}
	})

	cases := []runtimeWireCase{}
	for _, config := range []struct {
		name, profile        string
		prepared, operations bool
	}{
		{"generic-unconfigured", "generic", false, false}, {"generic-configured", "generic", true, true},
		{"generic-prepared-only", "generic", true, false}, {"generic-operations-only", "generic", false, true}, {"lab-configured", "lab", true, true},
	} {
		cases = append(cases, runtimeWireCase{planningWireCase: planningWireCase{Name: config.name, Method: "GET", Path: "/capabilities", Status: 200}, Profile: config.profile, Prepared: config.prepared, Operations: config.operations})
	}
	cases = append(cases, runtimeWireCase{planningWireCase: planningWireCase{Name: "capabilities-POST-refused", Method: "POST", Path: "/capabilities", Status: 405}, Profile: "generic"})
	cases = append(cases, runtimeWireCase{planningWireCase: planningWireCase{Name: "capabilities-query-refused", Method: "GET", Path: "/capabilities", Status: 400}, Profile: "generic", Query: "?foreign=1"})
	for _, path := range []string{"/http-package", "/operation-packages", "/operations", "/counter-rates", "/histograms"} {
		cases = append(cases, runtimeWireCase{planningWireCase: planningWireCase{Name: "unconfigured" + path, Method: "GET", Path: path, Status: 503}})
	}
	const filename = "../../testdata/http/runtime-capabilities.json"
	if os.Getenv("GOO_UPDATE_RUNTIME_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := runtimeWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		out, _ := json.MarshalIndent(runtimeWireFixtures{Version: 1, Cases: cases}, "", "  ")
		out = append(out, '\n')
		if err := os.WriteFile(filename, out, 0644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var fixtures runtimeWireFixtures
	if err = json.Unmarshal(raw, &fixtures); err != nil || fixtures.Version != 1 || len(fixtures.Cases) != len(cases) {
		t.Fatal("invalid fixtures", err)
	}
	for _, c := range fixtures.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := runtimeWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("header changed", k)
				}
			}
		})
	}
}
