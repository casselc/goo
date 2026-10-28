package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Materialize the final image's local COPY inputs, rather than accidentally
// validating startup against the complete developer checkout. The scanner must
// find the operation packages' referenced source declarations in that image.
func TestRuntimeImageLayoutStartsLabWithoutCheckout(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dockerfile, err := os.ReadFile(filepath.Join(repository, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	layout := t.TempDir()
	runtime := false
	for line := range strings.SplitSeq(string(dockerfile), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == "FROM" {
			runtime = fields[3] == "runtime"
		}
		if !runtime || len(fields) == 0 || fields[0] != "COPY" || strings.HasPrefix(fields[1], "--from=") {
			continue
		}
		if len(fields) != 3 {
			t.Fatalf("update runtime layout fixture for COPY instruction: %s", line)
		}
		if err := os.CopyFS(filepath.Join(layout, fields[2]), os.DirFS(filepath.Join(repository, fields[1]))); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(layout)
	server, err := configuredHTTPServer([]string{
		"-operation-package", "config/operations/checkout.yaml",
		"-operation-package", "config/operations/reports.yaml",
		"-binding", "config/http-package.yaml",
		"-inventory", "config/http-inventory.json",
		"-scope-name", "goo.http-sample",
		"-scope-version", "0.1.0",
		"-investigations-dir", filepath.Join(layout, "investigations"),
	})
	if err != nil {
		t.Fatalf("lab startup from runtime image layout: %v", err)
	}
	for _, path := range []string{"/api/investigation-workflows", "/api/measurement-packages"} {
		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("runtime layout %s: %d %s", path, response.Code, response.Body.String())
		}
		var catalog []json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &catalog); err != nil || len(catalog) == 0 {
			t.Fatalf("runtime layout %s has no valid builtins: %v %s", path, err, response.Body.String())
		}
	}
}
