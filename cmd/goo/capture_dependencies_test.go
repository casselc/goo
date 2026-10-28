package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestCaptureRegistrationsIsolateSourcesDraftsAndFrozenEvidence(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
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
		backend            *httptest.Server
		dir                string
		reads, clientCalls *atomic.Int32
		count              string
		baseline           exploration.CapturePage
		token              string
	}
	var regs []registration
	for n := 1; n <= 2; n++ {
		name := fmt.Sprintf("source-%d", n)
		count := fmt.Sprint(n)
		reads, clientCalls := &atomic.Int32{}, &atomic.Int32{}
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reads.Add(1)
			u, p, ok := r.BasicAuth()
			if !ok || u != name || p != "local-only" || r.URL.Query().Get("readonly") != "1" {
				t.Error("capture source auth/readonly changed")
			}
			sql, _ := io.ReadAll(r.Body)
			switch {
			case strings.Contains(string(sql), "SELECT DISTINCT"):
			case strings.HasPrefix(string(sql), "SELECT toString(count()) AS count"):
				fmt.Fprintf(w, "{\"count\":%q}\n", count)
			default:
				for j := 0; j < n; j++ {
					fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"duration_ns\":\"0\",\"marker\":%q}\n", 1000000001+j, name)
				}
			}
		}))
		defer backend.Close()
		client := backend.Client()
		client.Transport = nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
			clientCalls.Add(1)
			return http.DefaultTransport.RoundTrip(r)
		})
		s := exploration.Source{Endpoint: backend.URL, User: name, Password: "local-only", Client: client}
		dir := t.TempDir()
		mux := http.NewServeMux()
		registerCaptures(mux, s, dir, resultSnapshots)
		s.Endpoint = trap.URL
		regs = append(regs, registration{mux: mux, backend: backend, dir: dir, reads: reads, clientCalls: clientCalls, count: count})
	}
	call := func(mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(b))))
		return w
	}
	query := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
	question := exploration.CaptureQuestion{Kind: "explore", Explore: &query}
	payload := map[string]any{"title": "Injected baseline", "question": question}
	for i := range regs {
		reg := &regs[i]
		other := regs[1-i].reads.Load()
		w := call(reg.mux, "POST", "/captures", payload)
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &reg.baseline) != nil || reg.baseline.Explore == nil || reg.baseline.Explore.Count != reg.count {
			t.Fatal("capture source isolation", w.Code, w.Body.String())
		}
		before := reg.reads.Load()
		changed := query
		changed.Environment = "other"
		w = call(reg.mux, "POST", "/captures", map[string]any{"title": "Changed", "parent": reg.baseline.Receipt.ID, "question": exploration.CaptureQuestion{Kind: "explore", Explore: &changed}})
		if w.Code != 409 || reg.reads.Load() != before {
			t.Fatal("refresh mismatch reached source", w.Code)
		}
		b, _ := json.Marshal(payload)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		w = httptest.NewRecorder()
		reg.mux.ServeHTTP(w, httptest.NewRequest("POST", "/captures", strings.NewReader(string(b))).WithContext(ctx))
		if w.Code != 502 || reg.reads.Load() != before {
			t.Fatal("cancelled capture reached source", w.Code)
		}
		w = call(reg.mux, "POST", "/captures/live", map[string]string{"baseline": reg.baseline.Receipt.ID})
		var live struct {
			Token      string
			Comparison exploration.EvidenceComparison
		}
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &live) != nil || len(live.Token) != 64 || live.Comparison.Selection.Count != reg.count {
			t.Fatal("live source isolation", w.Code, w.Body.String())
		}
		reg.token = live.Token
		if reg.clientCalls.Load() == 0 || regs[1-i].reads.Load() != other {
			t.Fatal("client bypass/cross-registration reads")
		}
	}
	for i := range regs {
		regs[i].backend.Close()
	}
	for i, reg := range regs {
		before := reg.reads.Load()
		if w := call(reg.mux, "POST", "/captures/live", map[string]string{"freeze": regs[1-i].token}); w.Code != 410 {
			t.Fatal("foreign draft capability accepted", w.Code)
		}
		var frozen exploration.CapturePage
		w := call(reg.mux, "POST", "/captures/live", map[string]string{"freeze": reg.token})
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &frozen) != nil || frozen.Explore == nil || frozen.Explore.Count != reg.count {
			t.Fatal("source-free freeze", w.Code, w.Body.String())
		}
		var again exploration.CapturePage
		w = call(reg.mux, "POST", "/captures/live", map[string]string{"freeze": reg.token})
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &again) != nil || again.Receipt.ID != frozen.Receipt.ID {
			t.Fatal("freeze idempotence")
		}
		reopened := http.NewServeMux()
		registerCaptures(reopened, exploration.Source{Endpoint: trap.URL, Client: reg.backend.Client()}, reg.dir, resultSnapshots)
		for _, id := range []string{reg.baseline.Receipt.ID, frozen.Receipt.ID} {
			w = call(reopened, "GET", "/captures/"+id, nil)
			if w.Code != 200 {
				t.Fatal("source-free reopen", w.Code, w.Body.String())
			}
		}
		if w = call(reopened, "POST", "/captures/live", map[string]string{"freeze": reg.token}); w.Code != 410 {
			t.Fatal("draft survived owner replacement", w.Code)
		}
		if reg.reads.Load() != before {
			t.Fatal("freeze/reopen read source")
		}
	}
	if forbidden.Load() != 0 {
		t.Fatal("ambient source consulted")
	}
}
