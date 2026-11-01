package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func cliArtifactFile(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../../config/views/service-requests.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "view.yaml")
	if err = os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return file
}
func TestViewArtifactCLIValidationOfflineAndBudgets(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	file := cliArtifactFile(t)
	var out bytes.Buffer
	if err := run([]string{"validate-view", "-file", file}, &out); err != nil {
		t.Fatal(err)
	}
	var receipt viewArtifactCLIReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Version != "goo.view-validation/1" || receipt.Name != "service-requests" || receipt.View == nil || receipt.ID != "" {
		t.Fatal("invalid validation receipt", out.String())
	}
	raw, _ := os.ReadFile(file)
	atLimit := append(raw, bytes.Repeat([]byte(" "), workbench.ViewArtifactBytes-len(raw))...)
	if err := os.WriteFile(file, atLimit, 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"validate-view", "-file", file}, io.Discard); err != nil {
		t.Fatal("exact file budget refused", err)
	}
	for _, body := range [][]byte{append(atLimit, ' '), []byte("version: bad\n"), []byte("version: goo.view-artifact/1\nversion: duplicate\n")} {
		if err := os.WriteFile(file, body, 0600); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		if err := run([]string{"import-view", "-file", file, "-url", "http://source.invalid"}, &out); err == nil || out.Len() != 0 {
			t.Fatal("invalid file imported", err, out.String())
		}
	}
	if reads != 0 {
		t.Fatal("local validation attempted network", reads)
	}
	for _, args := range [][]string{{"validate-view"}, {"validate-view", "-file", file, "extra"}, {"import-view", "-file", file}, {"import-view", "-url", "http://localhost"}, {"validate-view", "-file", t.TempDir()}} {
		if err := run(args, io.Discard); err == nil {
			t.Fatal("invalid CLI invocation accepted", args)
		}
	}
}
func TestViewArtifactCLIDestination(t *testing.T) {
	for base, want := range map[string]string{"https://example.test": "https://example.test/api/exploration-views", "https://example.test/goo/": "https://example.test/goo/api/exploration-views", "https://example.test/team%2Fgoo": "https://example.test/team%2Fgoo/api/exploration-views",
		"https://example.test/team%2F":  "https://example.test/team%2F/api/exploration-views",
		"https://example.test/team%2F/": "https://example.test/team%2F/api/exploration-views", "https://example.test/team%20one": "https://example.test/team%20one/api/exploration-views", "https://example.test/api": "https://example.test/api/api/exploration-views"} {
		got, err := viewArtifactDestination(base)
		if err != nil || got != want {
			t.Fatalf("%q: %q %v", base, got, err)
		}
	}
	for _, base := range []string{"", "localhost:8080", "ftp://host", "http://user:secret@host", "https://host?x=1", "https://host?", "https://host#x", "https://host#", "http:///missing"} {
		if _, err := viewArtifactDestination(base); err == nil {
			t.Fatal("invalid destination accepted", base)
		}
	}
}
func TestViewArtifactCLIImportsExistingAPIExactlyOnce(t *testing.T) {
	dir := t.TempDir()
	api := http.NewServeMux()
	registerWorkbench(api, dir, retentionConnection{}, &exploration.ResultSnapshots{}, nil)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/team/api/exploration-views" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		http.StripPrefix("/team/api", api).ServeHTTP(w, r)
	}))
	defer server.Close()
	var out bytes.Buffer
	if err := run([]string{"import-view", "-file", cliArtifactFile(t), "-url", server.URL + "/team"}, &out); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("import made extra requests", calls)
	}
	var receipt viewArtifactCLIReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Version != "goo.view-import/1" || receipt.Namespace != "platform" || receipt.Name != "service-requests" || receipt.Revision != "v1" || !workbench.ID.MatchString(receipt.ID) || receipt.View != nil {
		t.Fatal("invalid import receipt", out.String())
	}
	saved, err := (workbench.Store{Dir: filepath.Join(dir, "workbench")}).LoadExplorationView(receipt.ID)
	if err != nil || saved.Title != "Service span volume" {
		t.Fatal("ordinary API persistence failed", err)
	}
}
func TestViewArtifactCLIRejectsUntrustworthyReceiptsWithoutRetry(t *testing.T) {
	artifact, err := readViewArtifactFile(cliArtifactFile(t))
	if err != nil {
		t.Fatal(err)
	}
	saved, err := (workbench.Store{Dir: t.TempDir()}).CreateExplorationView(artifact.Artifact.View)
	if err != nil {
		t.Fatal(err)
	}
	good, _ := json.Marshal(saved)
	for name, tc := range map[string]struct {
		status      int
		media, body string
	}{
		"changed view":  {201, "application/json", strings.Replace(string(good), "Service span volume", "Different question", 1)},
		"wrong ID":      {201, "application/json", strings.Replace(string(good), saved.ID, strings.Repeat("a", 64), 1)},
		"unknown field": {201, "application/json", string(good[:len(good)-1]) + `,"unknown":true}`},
		"trailing":      {201, "application/json", string(good) + ` {}`}, "null": {201, "application/json", "null"},
		"oversize":     {201, "application/json", string(good) + strings.Repeat(" ", workbench.MaxBytes)},
		"content type": {201, "text/html", string(good)}, "error": {503, "application/json", `{"error":"unavailable"}`},
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", tc.media)
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			}))
			defer server.Close()
			var out bytes.Buffer
			err := run([]string{"import-view", "-file", cliArtifactFile(t), "-url", server.URL}, &out)
			if err == nil || !strings.Contains(err.Error(), "may have saved") || out.Len() != 0 || calls != 1 {
				t.Fatal("bad receipt accepted/retried", err, out.String(), calls)
			}
		})
	}
	t.Run("redirect", func(t *testing.T) {
		calls := 0
		targetCalls := 0
		target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { targetCalls++ }))
		defer target.Close()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
		}))
		defer server.Close()
		err := run([]string{"import-view", "-file", cliArtifactFile(t), "-url", server.URL}, io.Discard)
		if err == nil || calls != 1 || targetCalls != 0 {
			t.Fatal("redirect followed", err, calls, targetCalls)
		}
	})
}
func TestViewArtifactCLIContextDeadlineAndNonReplayableBody(t *testing.T) {
	artifact, err := readViewArtifactFile(cliArtifactFile(t))
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err = submitViewArtifact(ctx, &http.Client{}, server.URL, artifact.Artifact.View); err == nil {
		t.Fatal("deadline ignored")
	}
	if calls.Load() != 1 {
		t.Fatal("deadline request retried", calls.Load())
	}
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		if r.GetBody != nil {
			t.Error("POST body is replayable")
		}
		return nil, io.ErrUnexpectedEOF
	})}
	if _, err = submitViewArtifact(context.Background(), client, "http://source.invalid", artifact.Artifact.View); err == nil {
		t.Fatal("transport error ignored")
	}
}
