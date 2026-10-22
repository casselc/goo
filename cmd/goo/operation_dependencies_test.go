package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/investigations"
	"github.com/casselc/goo/prebuilt"
)

func TestOperationAndInvestigationRegistrationsUseInjectedConnections(t *testing.T) {
	plan, e := annotations.LoadOperationPackage("../../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	plans := map[string]*annotations.OperationPlan{"checkout": plan}
	ordered := []*annotations.OperationPlan{plan}
	b, e := os.ReadFile("../../config/http-package.yaml")
	if e != nil {
		t.Fatal(e)
	}
	binding, e := prebuilt.ReadBinding(bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	b, e = os.ReadFile("../../config/http-inventory.json")
	if e != nil {
		t.Fatal(e)
	}
	inventory, e := prebuilt.ReadInventory(bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	expanded, e := prebuilt.Expand(inventory, binding)
	if e != nil {
		t.Fatal(e)
	}
	httpSource := prebuilt.HTTPSource{ScopeName: "fixture.http", ScopeVersion: "1", SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
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
		mux          *http.ServeMux
		backend      *httptest.Server
		dir          string
		reads, calls *atomic.Int32
		capture      investigations.Capture
	}
	var regs []registration
	for _, name := range []string{"one", "two"} {
		reads, calls := &atomic.Int32{}, &atomic.Int32{}
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reads.Add(1)
			u, p, ok := r.BasicAuth()
			if !ok || u != name || p != "local-only" || r.URL.Query().Get("readonly") != "1" {
				t.Error("operation auth/readonly changed")
			}
			io.Copy(io.Discard, r.Body)
		}))
		defer backend.Close()
		client := backend.Client()
		client.Transport = nativeDependencyTransport(func(r *http.Request) (*http.Response, error) { calls.Add(1); return http.DefaultTransport.RoundTrip(r) })
		c := clickHouseConnection{Endpoint: backend.URL, User: name, Password: "local-only", Client: client}
		dir := t.TempDir()
		mux := http.NewServeMux()
		registerOperationReads(mux, plans, ordered, expanded, httpSource, c)
		registerInvestigations(mux, plans, c, dir)
		c.Endpoint = trap.URL
		regs = append(regs, registration{mux: mux, backend: backend, dir: dir, reads: reads, calls: calls})
	}
	call := func(mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(b)))
		return w
	}
	q := investigations.Question{Package: "checkout", Revision: plan.Revision, FromMS: 60000, ToMS: 120000}
	for i := range regs {
		reg := &regs[i]
		other := regs[1-i].reads.Load()
		before := reg.reads.Load()
		if w := call(reg.mux, "GET", "/operation-packages", nil); w.Code != 200 || reg.reads.Load() != before {
			t.Fatal("declarations read source", w.Code)
		}
		for _, path := range []string{"/http-package", "/http-package?minutes=5", "/operations", "/operations?package=checkout", "/counter-rates?package=checkout&fromMs=60000&toMs=120000", "/histograms?package=checkout&fromMs=60000&toMs=120000"} {
			before = reg.reads.Load()
			w := call(reg.mux, "GET", path, nil)
			if w.Code != 200 || reg.reads.Load() <= before || w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("injected operation route", path, w.Code, w.Body.String())
			}
			before = reg.reads.Load()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w = httptest.NewRecorder()
			reg.mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil).WithContext(ctx))
			if w.Code != 502 || reg.reads.Load() != before {
				t.Fatal("cancelled operation read source", path, w.Code)
			}
		}
		w := call(reg.mux, "POST", "/investigations", map[string]any{"question": q})
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &reg.capture) != nil || reg.capture.Question != q {
			t.Fatal("injected investigation", w.Code, w.Body.String())
		}
		before = reg.reads.Load()
		changed := q
		changed.Revision = "unavailable"
		if w = call(reg.mux, "POST", "/investigations", map[string]any{"question": changed}); w.Code != 409 || reg.reads.Load() != before {
			t.Fatal("revision mismatch read source", w.Code)
		}
		changed = q
		changed.ToMS++
		if w = call(reg.mux, "POST", "/investigations", map[string]any{"question": changed, "parent": reg.capture.ID}); w.Code != 409 || reg.reads.Load() != before {
			t.Fatal("refresh mismatch read source", w.Code)
		}
		if reg.calls.Load() == 0 || regs[1-i].reads.Load() != other {
			t.Fatal("client bypass/cross-registration read")
		}
	}
	for i := range regs {
		regs[i].backend.Close()
	}
	for _, reg := range regs {
		before := reg.reads.Load()
		fresh := http.NewServeMux()
		registerInvestigations(fresh, nil, clickHouseConnection{Endpoint: trap.URL, Client: reg.backend.Client()}, reg.dir)
		w := call(fresh, "GET", "/investigations/"+reg.capture.ID, nil)
		var got investigations.Capture
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil || got.ID != reg.capture.ID || reg.reads.Load() != before {
			t.Fatal("offline reopen changed evidence", w.Code)
		}
	}
	if forbidden.Load() != 0 {
		t.Fatal("ambient source consulted")
	}
}

func TestSharedClickHouseStartupSnapshotFeedsTypedAdapters(t *testing.T) {
	t.Setenv("CLICKHOUSE_URL", "http://configured.invalid")
	t.Setenv("CLICKHOUSE_USER", "fixture")
	t.Setenv("CLICKHOUSE_PASSWORD", "local-only")
	client := &http.Client{}
	c := clickHouseConnectionFromEnvironment(client)
	t.Setenv("CLICKHOUSE_URL", "http://other.invalid")
	n, r := c.measurementSource(), c.explorationSource()
	if c.Endpoint != "http://configured.invalid" || c.User != "fixture" || c.Password != "local-only" || n.Endpoint != c.Endpoint || r.Endpoint != c.Endpoint || n.User != c.User || r.User != c.User || n.Password != c.Password || r.Password != c.Password || n.Client != client || r.Client != client {
		t.Fatal("startup adapter snapshot drift")
	}
}
