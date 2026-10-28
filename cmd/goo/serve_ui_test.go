package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/queryhttp"
	uiassets "github.com/casselc/goo/ui"
)

func TestEmbeddedUIAndMetadataWithoutSource(t *testing.T) {
	for _, profile := range []string{"generic", "lab"} {
		t.Run(profile, func(t *testing.T) {
			mux := http.NewServeMux()
			if err := registerUIAPI(mux, profile); err != nil {
				t.Fatal(err)
			}
			handler, err := newUIHandler(mux)
			if err != nil {
				t.Fatal(err)
			}
			request := func(path string) *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
				return w
			}
			for path, contentType := range map[string]string{"/": "text/html", "/index.html": "text/html", "/ui-components": "text/html", "/app.mjs": "application/javascript", "/components.css": "text/css", "/api-docs": "text/html", "/openapi.json": "application/json"} {
				w := request(path)
				if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), contentType) || w.Header().Get("Cache-Control") != "no-store" {
					t.Fatal(path, w.Code, w.Header())
				}
			}
			// Every flat display dependency ships with the binary and its exact bytes.
			entries, _ := fs.ReadDir(uiassets.Assets, ".")
			for _, entry := range entries {
				name := entry.Name()
				if !strings.HasSuffix(name, ".mjs") && !strings.HasSuffix(name, ".css") {
					continue
				}
				w := request("/" + name)
				body, _ := uiassets.Assets.ReadFile(name)
				if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), body) {
					t.Fatal("asset missing", name)
				}
			}
			w := request("/api/ui-config")
			var config struct {
				Profile  string
				Examples struct{ State string }
			}
			if json.Unmarshal(w.Body.Bytes(), &config) != nil || config.Profile != profile {
				t.Fatal(w.Body.String())
			}
			expected := "unconfigured"
			if profile == "lab" {
				expected = "configured"
			}
			if config.Examples.State != expected {
				t.Fatal(config)
			}
			for _, path := range []string{"/api/example-questions", "/api/prebuilts"} {
				w := request(path)
				want := 503
				if profile == "lab" {
					want = 200
				}
				if w.Code != want || !json.Valid(w.Body.Bytes()) {
					t.Fatal(path, w.Code, w.Body.String())
				}
			}
			for _, path := range []string{"/api/ui-config?unexpected=1", "/api/health?unexpected=1"} {
				if w := request(path); w.Code != 400 {
					t.Fatal(path, w.Code)
				}
			}
			for _, path := range []string{"/api/no-such-route", "/no-such-route", "/assets.go", "/server.py", "/api/package", "/api/query", "/api/explain", "/components.html"} {
				w := request(path)
				if w.Code != 404 || strings.Contains(w.Body.String(), "app.mjs") {
					t.Fatal("unexpected fallback", path, w.Code)
				}
			}
			if w := request("/health"); w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"ok":true}` {
				t.Fatal(w.Code, w.Body.String())
			}
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("HEAD", "/app.mjs", nil))
			if w.Code != 200 || w.Body.Len() != 0 {
				t.Fatal("HEAD", w.Code, w.Body.Len())
			}
		})
	}
}

func TestBrowserAPIPreservesRequestAndStreamsCancellation(t *testing.T) {
	for _, prefix := range []string{"", "/api"} {
		t.Run(prefix, func(t *testing.T) {
			ready := make(chan struct{})
			finished := make(chan struct{})
			source := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/explore" || r.URL.RawQuery != "scope=a%2Fb" || r.Header.Get("X-Test") != "preserved" {
					t.Error("request changed", r.URL, r.Header)
				}
				var v struct {
					Value string `json:"value"`
				}
				if err := httpjson.Decode(w, r, &v, 1024, "one object"); err != nil {
					t.Error(err)
					return
				}
				if v.Value != "kept" {
					t.Error(v)
				}
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("X-Test", "stream")
				fmt.Fprint(w, "data: ready\n\n")
				w.(http.Flusher).Flush()
				close(ready)
				<-r.Context().Done()
				close(finished)
			})
			budget := queryhttp.NewBudget(source, 1, func(r *http.Request) bool { return r.URL.Path == "/explore" })
			handler, err := newUIHandler(budget)
			if err != nil {
				t.Fatal(err)
			}
			srv := httptest.NewServer(handler)
			defer srv.Close()
			srv.Client().Timeout = 2 * time.Second
			id := "12345678-1234-1234-1234-123456789abc"
			req, _ := http.NewRequest("POST", srv.URL+prefix+"/explore?scope=a%2Fb", strings.NewReader(`{"value":"kept"}`))
			req.Header.Set("X-Test", "preserved")
			req.Header.Set("X-Goo-Query-ID", id)
			response, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.Header.Get("X-Test") != "stream" {
				t.Fatal("response headers changed")
			}
			select {
			case <-ready:
			case <-time.After(time.Second):
				t.Fatal("stream buffered")
			}
			// The opposite prefix cancels the SAME in-flight budget/registry.
			other := "/api"
			if prefix == "/api" {
				other = ""
			}
			ack, err := srv.Client().Post(srv.URL+other+"/queries/"+id+"/cancel", "application/json", nil)
			if err != nil {
				t.Fatal(err)
			}
			ack.Body.Close()
			if ack.StatusCode != 202 {
				t.Fatal(ack.StatusCode)
			}
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("cancellation did not reach handler")
			}
			b, err := io.ReadAll(response.Body)
			if err != nil || !strings.Contains(string(b), "data: ready") {
				t.Fatal(string(b), err)
			}
		})
	}
}

func TestBrowserAPICancelsIncompleteIncomingBody(t *testing.T) {
	var effects atomic.Int32
	entered := make(chan struct{})
	released := make(chan struct{}, 2)
	api := queryhttp.NewBudget(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-entered:
		default:
			close(entered)
		}
		var input struct{}
		if err := httpjson.Decode(w, r, &input, 1024, "one object"); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		effects.Add(1)
		w.WriteHeader(204)
	}), 1, func(r *http.Request) bool { return r.URL.Path == "/explore" })
	handler, err := newUIHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.ServeHTTP(w, r)
		if r.URL.Path == "/explore" {
			released <- struct{}{}
		}
	}))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	defer srv.Close()
	srv.Client().Timeout = time.Second
	peer, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	id := "12345678-1234-1234-1234-123456789abc"
	fmt.Fprintf(peer, "POST /api/explore HTTP/1.1\r\nHost: test\r\nContent-Length: 100\r\nX-Goo-Query-ID: %s\r\n\r\n{}", id)
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("not entered")
	}
	ack, err := srv.Client().Post(srv.URL+"/api/queries/"+id+"/cancel", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	ack.Body.Close()
	if ack.StatusCode != 202 {
		t.Fatal(ack.StatusCode)
	}
	select {
	case <-released:
	case <-time.After(time.Second):
		t.Fatal("slot retained")
	}
	if effects.Load() != 0 {
		t.Fatal("incomplete body started effects")
	}
	response, err := srv.Client().Post(srv.URL+"/api/explore", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 204 || effects.Load() != 1 {
		t.Fatal(response.StatusCode, effects.Load())
	}
}

func TestConfiguredUIWithoutSourceCheckout(t *testing.T) {
	if os.Getenv("GOO_EMBEDDED_UI_CHILD") == "1" {
		server, err := configuredHTTPServer([]string{"--profile", "generic", "--investigations-dir", t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{"/", "/app.mjs", "/api-docs", "/openapi.json", "/api/ui-config", "/api/capabilities", "/api/health"} {
			w := httptest.NewRecorder()
			server.Handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code != 200 {
				t.Fatal(path, w.Code, w.Body.String())
			}
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "-test.run=^TestConfiguredUIWithoutSourceCheckout$")
	command.Dir = t.TempDir()
	command.Env = append(os.Environ(), "GOO_EMBEDDED_UI_CHILD=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("standalone embedded UI: %v\n%s", err, output)
	}
}
