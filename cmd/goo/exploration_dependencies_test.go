package main

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

func TestExplorationStartupDefaultsAndSnapshot(t *testing.T) {
	t.Setenv("CLICKHOUSE_URL", "")
	t.Setenv("CLICKHOUSE_USER", "record-fixture")
	t.Setenv("CLICKHOUSE_PASSWORD", "local-only")
	client := &http.Client{}
	s := clickHouseConnectionFromEnvironment(client).explorationSource()
	t.Setenv("CLICKHOUSE_URL", "http://example.invalid")
	t.Setenv("CLICKHOUSE_USER", "other")
	t.Setenv("CLICKHOUSE_PASSWORD", "other")
	if s.Endpoint != "http://localhost:8123" || s.User != "record-fixture" || s.Password != "local-only" || s.Client != client {
		t.Fatal("startup source default/snapshot changed")
	}
}
func TestExplorationAndTailRegistrationsKeepInjectedSources(t *testing.T) {
	var forbidden atomic.Int32
	trap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forbidden.Add(1)
		http.Error(w, "unexpected ambient source", 500)
	}))
	defer trap.Close()
	t.Setenv("CLICKHOUSE_URL", trap.URL)
	t.Setenv("CLICKHOUSE_USER", "ambient")
	t.Setenv("CLICKHOUSE_PASSWORD", "ambient")
	type registration struct {
		mux                *http.ServeMux
		reads, clientCalls *atomic.Int32
	}
	var registrations []registration
	for _, name := range []string{"one", "two"} {
		reads, clientCalls := &atomic.Int32{}, &atomic.Int32{}
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reads.Add(1)
			u, p, ok := r.BasicAuth()
			if !ok || u != name || p != "local-only" || r.URL.Query().Get("readonly") != "1" {
				t.Error("injected auth/readonly changed")
			}
			sql, _ := io.ReadAll(r.Body)
			if strings.HasPrefix(string(sql), "SELECT toString(count()) AS count") {
				io.WriteString(w, "{\"count\":\"0\"}\n")
			}
		}))
		defer backend.Close()
		client := backend.Client()
		client.Transport = nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
			clientCalls.Add(1)
			return http.DefaultTransport.RoundTrip(r)
		})
		s := exploration.Source{Endpoint: backend.URL, User: name, Password: "local-only", Client: client}
		mux := http.NewServeMux()
		registerExploration(mux, s)
		s.Endpoint = trap.URL
		registrations = append(registrations, registration{mux, reads, clientCalls})
	}
	// Both routers remain registered while each is exercised, exposing singleton drift.
	for i, reg := range registrations {
		other := registrations[1-i].reads.Load()
		for _, route := range []struct{ path, body string }{{"/explore", `{"signal":"spans","environment":"prod","fromMs":1,"toMs":60001,"limit":100,"offset":0}`}, {"/inventory", `{"environment":"prod","fromMs":1,"toMs":60001}`}, {"/discovery", `{"environment":"prod","fromMs":1,"toMs":60001}`}} {
			before := reg.reads.Load()
			w := httptest.NewRecorder()
			reg.mux.ServeHTTP(w, httptest.NewRequest("POST", route.path, strings.NewReader(`{}`)))
			if w.Code != 400 || reg.reads.Load() != before {
				t.Fatal("invalid input reached source", route.path)
			}
			w = httptest.NewRecorder()
			reg.mux.ServeHTTP(w, httptest.NewRequest("POST", route.path, strings.NewReader(route.body)))
			if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || reg.reads.Load() <= before {
				t.Fatal("injected route failed", route.path, w.Code, w.Body.String())
			}
			before = reg.reads.Load()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w = httptest.NewRecorder()
			reg.mux.ServeHTTP(w, httptest.NewRequest("POST", route.path, strings.NewReader(route.body)).WithContext(ctx))
			if w.Code != 502 || reg.reads.Load() != before {
				t.Fatal("cancelled request reached backend", route.path, w.Code)
			}
		}
		if reg.clientCalls.Load() == 0 {
			t.Fatal("injected client unused")
		}
		server := httptest.NewServer(reg.mux)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "POST", server.URL+"/tail", strings.NewReader(`{"environment":"prod"}`))
		resp, e := server.Client().Do(req)
		if e != nil {
			cancel()
			server.Close()
			t.Fatal(e)
		}
		scanner := bufio.NewScanner(resp.Body)
		reset, batch := false, false
		for scanner.Scan() {
			line := scanner.Text()
			if line == "event: reset" {
				reset = true
			}
			if line == "event: batch" {
				batch = true
				break
			}
		}
		cancel()
		resp.Body.Close()
		server.Close()
		if !reset || !batch || resp.Header.Get("Content-Type") != "text/event-stream" || resp.Header.Get("Cache-Control") != "no-store" || resp.Header.Get("X-Accel-Buffering") != "no" {
			t.Fatal("injected tail framing changed", scanner.Err())
		}
		if registrations[1-i].reads.Load() != other {
			t.Fatal("cross-registration source drift")
		}
	}
	if forbidden.Load() != 0 {
		t.Fatal("request-time ambient configuration used")
	}
}
