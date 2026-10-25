package main

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenericStartupWithoutSampleFiles(t *testing.T) {
	if os.Getenv("GOO_GENERIC_STARTUP_CHILD") == "1" {
		server, e := configuredHTTPServer([]string{"--profile", "generic", "--investigations-dir", filepath.Join(os.TempDir(), "goo-uncreated-generic-test-store")})
		if e != nil {
			t.Fatal(e)
		}
		for _, path := range []string{"/capabilities", "/operation-packages", "/operations", "/counter-rates", "/histograms", "/http-package"} {
			w := httptest.NewRecorder()
			server.Handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			expected := 503
			if path == "/capabilities" {
				expected = 200
			}
			if w.Code != expected {
				t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
			}
			if expected == 503 && !strings.Contains(w.Body.String(), `"code":"capability_unconfigured"`) {
				t.Fatal(w.Body.String())
			}
		}
		// Pure dashboard planning remains available without operation declarations.
		w := httptest.NewRecorder()
		server.Handler.ServeHTTP(w, httptest.NewRequest("POST", "/dashboards/plan", strings.NewReader(`{"scope":{"environment":"prod","fromMs":1000,"toMs":61000},"service":"ordinary-python","signals":["spans"],"fields":[],"instruments":[]}`)))
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		// Existing saved-capture route remains registered rather than unavailable-family.
		w = httptest.NewRecorder()
		server.Handler.ServeHTTP(w, httptest.NewRequest("GET", "/investigations/not-an-id", nil))
		if w.Code != 404 {
			t.Fatal(w.Code)
		}
		return
	}
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(executable, "-test.run=^TestGenericStartupWithoutSampleFiles$")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "GOO_GENERIC_STARTUP_CHILD=1")
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("empty cwd startup: %v\n%s", e, out)
	}
}
func TestStartupProfilesFailClosedExplicitConfiguration(t *testing.T) {
	for _, args := range [][]string{nil, {"--profile", "future"}, {"--profile", "generic", "--binding", "missing.yaml"}, {"--profile", "generic", "--binding", ""}, {"--profile", "generic", "--binding", "missing", "--inventory", "missing", "--scope-name", "sdk", "--scope-version", "1"}, {"--profile", "generic", "--operation-package", "missing.yaml"}, {"--profile", "generic", "unexpected"}} {
		if _, e := configuredHTTPServer(args); e == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
func TestGenericExplicitOperationPackageRemainsAvailable(t *testing.T) {
	packagePath, e := filepath.Abs("../../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	server, e := configuredHTTPServer([]string{"--profile", "generic", "--operation-package", packagePath, "--investigations-dir", t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, httptest.NewRequest("GET", "/operation-packages", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "checkout") {
		t.Fatal(w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	server.Handler.ServeHTTP(w, httptest.NewRequest("GET", "/http-package", nil))
	if w.Code != 503 || !strings.Contains(w.Body.String(), "prepared-http") {
		t.Fatal(w.Code, w.Body.String())
	}
}
