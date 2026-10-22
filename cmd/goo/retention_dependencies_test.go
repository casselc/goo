package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
)

func TestRetentionStartupAndIndependentAvailability(t *testing.T) {
	t.Setenv("GOO_XTDB_URL", "")
	empty := retentionConnectionFromEnvironment()
	t.Setenv("GOO_XTDB_URL", "postgres://fixture@127.0.0.1:1/xtdb?sslmode=disable")
	configured := retentionConnectionFromEnvironment()
	t.Setenv("GOO_XTDB_URL", "")
	if empty.URL != "" || configured.URL == "" {
		t.Fatal("startup snapshot changed")
	}
	for _, c := range []struct {
		connection retentionConnection
		configured bool
	}{{empty, false}, {configured, true}} {
		mux := http.NewServeMux()
		registerWorkbench(mux, t.TempDir(), c.connection)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", "/retained-evidence/status", nil))
		want := fmt.Sprintf("{\"configured\":%t}\n", c.configured)
		if w.Code != 200 || w.Body.String() != want || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("optional status changed", w.Code, w.Body.String())
		}
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/retained-evidence", strings.NewReader(`{}`)))
		expected := 503
		if c.configured {
			expected = 400
		}
		if w.Code != expected {
			t.Fatal("availability/validation order", w.Code)
		}
	}
}

func TestXTDBHTTPInjectedCitationBasisAndOfflineDocument(t *testing.T) {
	address := os.Getenv("GOO_XTDB_TEST_URL")
	if address == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	wire, e := os.ReadFile("../../testdata/storage/exploration-capture.json")
	if e != nil {
		t.Fatal(e)
	}
	var capture exploration.SavedCapture
	if e = json.Unmarshal(wire, &capture); e != nil {
		t.Fatal(e)
	}
	unique := fmt.Sprintf("%x", sha256.Sum256([]byte(t.Name()+time.Now().String())))
	capture.ID = unique
	// The retained-record adapter consumes the exporter's decimal-string clock shape.
	capture.Explore.Rows[0] = json.RawMessage(`{"ServiceName":"api","timestamp_ns":"1000000001","duration_ns":"1000000","SpanId":"0000000000000001","TraceId":"00000000000000000000000000000001"}`)
	connection := retentionConnection{URL: address}
	manifest, e := retention.Select(&capture, []int{0})
	if e != nil {
		t.Fatal(e)
	}
	initial, e := connection.store("").Publish(ctx, manifest, []string{"review:wiring"}, "Original pinned review", unique)
	if e != nil {
		t.Fatal(e)
	}
	input := initial.Annotation.MeaningInput
	input.Parent = initial.Annotation.ID
	input.Note = "Later corrected review"
	corrected, e := connection.store(initial.AwaitToken).CorrectOnce(ctx, manifest.ID, input, fmt.Sprintf("%x", sha256.Sum256([]byte(unique+"correction"))))
	if e != nil {
		t.Fatal(e)
	}
	if initial.AwaitToken == "" || corrected.AwaitToken == "" {
		t.Fatal("missing committed read basis")
	}
	t.Setenv("GOO_XTDB_URL", "")
	root := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, root, connection)
	connection.URL = "postgres://fixture@127.0.0.1:1/xtdb?sslmode=disable"
	call := func(mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(b)).WithContext(ctx))
		return w
	}
	cards := []workbench.Card{{Kind: "retained", Reference: manifest.ID, Review: initial.Annotation.ID, ReadBasis: corrected.AwaitToken}, {Kind: "retained", Reference: manifest.ID, Review: corrected.ID, ReadBasis: corrected.AwaitToken}}
	w := call(mux, "POST", "/documents", map[string]any{"document": workbench.DocumentInput{Title: "Pinned configuration investigation", Cards: cards}})
	var doc workbench.Document
	if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &doc) != nil || len(doc.Cards) != 2 || doc.Cards[0].Review != cards[0].Review || doc.Cards[0].ReadBasis != corrected.AwaitToken {
		t.Fatal("injected citation verification", w.Code, w.Body.String())
	}
	// Distinct request tokens must not become mutable shared store state.
	var wg sync.WaitGroup
	for _, token := range []string{corrected.AwaitToken, "not-a-valid-await-token"} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			path := "/retained-evidence/" + manifest.ID + "/annotations/" + initial.Annotation.ID + "?awaitToken=" + url.QueryEscape(token)
			r := call(mux, "GET", path, nil)
			if token == corrected.AwaitToken {
				var review workbench.MeaningRevision
				if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &review) != nil || review.Note != "Original pinned review" {
					t.Error("valid pinned basis", r.Code, r.Body.String())
				}
			} else if r.Code != 404 {
				t.Error("invalid request basis accepted", r.Code)
			}
		}(token)
	}
	wg.Wait()
	// Documents reopen from their own files even without optional-store configuration.
	offline := http.NewServeMux()
	registerWorkbench(offline, root, retentionConnection{})
	w = call(offline, "GET", "/documents/"+doc.ID, nil)
	var reopened workbench.Document
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &reopened) != nil || reopened.ID != doc.ID || reopened.Cards[1].Review != corrected.ID || reopened.Cards[1].ReadBasis != corrected.AwaitToken {
		t.Fatal("offline pinned document reopen", w.Code, w.Body.String())
	}
	w = call(offline, "POST", "/documents", map[string]any{"document": workbench.DocumentInput{Title: "Unverified citation", Cards: cards}})
	if w.Code != 400 {
		t.Fatal("unconfigured citation published", w.Code)
	}
}
