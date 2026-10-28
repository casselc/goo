package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestConfiguredServersOwnResultAndExplanationCapabilities(t *testing.T) {
	var reads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		text := string(data)
		switch {
		case strings.Contains(text, "GROUP BY"):
			side := "baseline"
			for _, values := range r.URL.Query() {
				for _, value := range values {
					if value == "selected" {
						side = "selected"
					}
				}
			}
			json.NewEncoder(w).Encode(map[string]string{"g0": side, "count": "1"})
		case strings.Contains(text, "arrayJoin"):
			// Empty attribute discovery is valid; fixed columns remain available.
		case strings.Contains(text, "count()"):
			fmt.Fprintln(w, `{"count":"1"}`)
		default:
			fmt.Fprintln(w, `{"timestamp_ns":"1500000000","duration_ns":"10","service":"baseline","trace_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","span_id":"aaaaaaaaaaaaaaaa","name":"work"}`)
		}
	}))
	defer source.Close()
	t.Setenv("CLICKHOUSE_URL", source.URL)
	// A malformed optional connection fails locally. The owning server should
	// reject the deliberately out-of-range selection before attempting recovery;
	// a foreign/expired token must preserve the existing ambiguous-recovery path.
	t.Setenv("GOO_XTDB_URL", "://invalid")
	dir := filepath.Join(t.TempDir(), "shared-durable-captures")
	newServer := func() *http.Server {
		server, err := configuredHTTPServer([]string{"--profile", "generic", "--investigations-dir", dir})
		if err != nil {
			t.Fatal(err)
		}
		return server
	}
	a, b := newServer(), newServer()
	call := func(server *http.Server, method, path string, body any) *httptest.ResponseRecorder {
		var raw []byte
		if body != nil {
			var err error
			raw, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		w := httptest.NewRecorder()
		server.Handler.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(raw)))
		return w
	}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "baseline"}}
	records := call(a, "POST", "/explore", q)
	var result exploration.Result
	if records.Code != 200 || json.Unmarshal(records.Body.Bytes(), &result) != nil || result.ResultToken == "" {
		t.Fatal("Explore did not hold its page", records.Code, records.Body.String())
	}
	for _, tc := range []struct {
		server *http.Server
		want   int
	}{{a, 400}, {b, 409}} {
		w := call(tc.server, "POST", "/captures", map[string]any{"title": "Wrong token family", "resultToken": result.ResultToken})
		if w.Code != tc.want {
			t.Fatal("capture loader crossed server boundary", w.Code, w.Body.String())
		}
	}
	retentionIntent := map[string]any{"requestID": strings.Repeat("a", 64), "resultToken": result.ResultToken, "indexes": []int{1}, "tags": []string{}, "note": "out of range"}
	if w := call(a, "POST", "/retained-evidence", retentionIntent); w.Code != 400 {
		t.Fatal("retention did not see its own held page", w.Code, w.Body.String())
	}
	if w := call(b, "POST", "/retained-evidence", retentionIntent); w.Code != 502 || !strings.Contains(w.Body.String(), "retry unchanged") {
		t.Fatal("foreign token bypassed existing recovery semantics", w.Code, w.Body.String())
	}
	selected := q
	selected.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "selected"}
	explained := call(a, "POST", "/population/explain", exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: q, Selection: selected, Fields: []exploration.Field{{Kind: "column", Key: "service"}}})
	var explanation exploration.ExplanationResult
	if explained.Code != 200 || json.Unmarshal(explained.Body.Bytes(), &explanation) != nil || explanation.ResultToken == "" {
		t.Fatal("explanation was not held", explained.Code, explained.Body.String())
	}
	before := reads.Load()
	promotion := map[string]any{"title": "Observed explanation", "resultToken": explanation.ResultToken}
	if w := call(b, "POST", "/captures", promotion); w.Code != 409 {
		t.Fatal("foreign explanation promoted", w.Code, w.Body.String())
	}
	frozen := call(a, "POST", "/captures", promotion)
	var capture struct{ Receipt struct{ ID string } }
	if frozen.Code != 201 || json.Unmarshal(frozen.Body.Bytes(), &capture) != nil || capture.Receipt.ID == "" {
		t.Fatal("own explanation could not promote", frozen.Code, frozen.Body.String())
	}
	// Shared storage intentionally shares durable captures, not live capabilities.
	for _, server := range []*http.Server{b, newServer()} {
		if w := call(server, "POST", "/captures", promotion); w.Code != 409 {
			t.Fatal("new/other server retained a foreign capability", w.Code, w.Body.String())
		}
		if w := call(server, "GET", "/captures/"+capture.Receipt.ID, nil); w.Code != 200 {
			t.Fatal("server isolation hid a shared durable capture", w.Code, w.Body.String())
		}
	}
	if w := call(a, "POST", "/captures", promotion); w.Code != 201 || w.Body.String() != frozen.Body.String() {
		t.Fatal("same-server promotion lost idempotence", w.Code)
	}
	if reads.Load() != before {
		t.Fatal("load, promotion or recovery reread telemetry")
	}
}
