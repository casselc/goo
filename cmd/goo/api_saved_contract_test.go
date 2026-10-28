package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
)

type apiObservation struct {
	Method   string          `json:"method"`
	Path     string          `json:"path"`
	Status   int             `json:"status"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response"`
}

// Execute a multi-owner user workflow and optionally export actual wire values
// for OpenAPI validation. Source-free operations must still work after failure.
func TestAPISavedEvidenceObservedContracts(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	root := t.TempDir()
	reads := 0
	offline := false
	client := &http.Client{Transport: nativeDependencyTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		if offline {
			t.Fatal("source-free operation attempted telemetry read")
		}
		sql, _ := io.ReadAll(r.Body)
		payload := ""
		switch {
		case strings.Contains(string(sql), "SELECT DISTINCT"):
		case strings.Contains(string(sql), "toString(count()) AS count"):
			payload = `{"count":"1"}`
		default:
			payload = `{"timestamp_ns":"1000000001","duration_ns":"0","ServiceName":"ordinary-api","TraceId":"11111111111111111111111111111111","SpanId":"2222222222222222","Attributes":{"region":"west"}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	mux := http.NewServeMux()
	capturehttp.Register(mux, exploration.Source{Endpoint: "http://source", Client: client}, root, resultSnapshots)
	registerWorkbench(mux, root, retentionConnection{}, resultSnapshots, nil)
	var observed []apiObservation
	call := func(method, path, pattern string, input any, want int) json.RawMessage {
		var body []byte
		if input != nil {
			body, _ = json.Marshal(input)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(string(body))))
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		observed = append(observed, apiObservation{method, pattern, w.Code, body, append(json.RawMessage(nil), w.Body.Bytes()...)})
		return w.Body.Bytes()
	}
	query := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
	raw := call("POST", "/captures", "/captures", map[string]any{"title": "API contract evidence", "question": exploration.CaptureQuestion{Kind: "explore", Explore: &query}}, 201)
	var baseline exploration.CapturePage
	if err := json.Unmarshal(raw, &baseline); err != nil {
		t.Fatal(err)
	}
	if baseline.Explore == nil || baseline.Explore.Count != "1" || reads == 0 {
		t.Fatal("missing source population")
	}
	raw = call("POST", "/captures/live", "/captures/live", map[string]string{"baseline": baseline.Receipt.ID}, 200)
	var draft struct{ Token string }
	json.Unmarshal(raw, &draft)
	if len(draft.Token) != 64 {
		t.Fatal("missing draft")
	}
	offline = true
	before := reads
	raw = call("POST", "/captures/live", "/captures/live", map[string]string{"freeze": draft.Token}, 201)
	var frozen exploration.CapturePage
	json.Unmarshal(raw, &frozen)
	again := call("POST", "/captures/live", "/captures/live", map[string]string{"freeze": draft.Token}, 201)
	var duplicate exploration.CapturePage
	json.Unmarshal(again, &duplicate)
	if duplicate.Receipt.ID != frozen.Receipt.ID {
		t.Fatal("freeze changed identity")
	}
	call("GET", "/captures/"+frozen.Receipt.ID, "/captures/{id}", nil, 200)
	call("GET", "/captures/"+frozen.Receipt.ID+"/receipt", "/captures/{id}/receipt", nil, 200)
	call("GET", "/captures", "/captures", nil, 200)
	doc := workbench.DocumentInput{Title: "API contract investigation", Cards: []workbench.Card{{Kind: "capture", Reference: frozen.Receipt.ID, Caption: "Preserved observation"}}}
	raw = call("POST", "/documents", "/documents", map[string]any{"document": doc}, 201)
	var saved workbench.Document
	json.Unmarshal(raw, &saved)
	call("GET", "/documents/"+saved.ID, "/documents/{id}", nil, 200)
	call("GET", "/documents", "/documents", nil, 200)
	call("GET", "/retained-evidence/status", "/retained-evidence/status", nil, 200)
	call("POST", "/retained-evidence", "/retained-evidence", map[string]any{}, 503)
	// Optional real native XTDB: selected records, exact reviews and read basis.
	if address := os.Getenv("GOO_XTDB_TEST_URL"); address != "" {
		mux = http.NewServeMux()
		capturehttp.Register(mux, exploration.Source{Endpoint: "http://source", Client: client}, root, resultSnapshots)
		registerWorkbench(mux, root, retentionConnection{URL: address}, resultSnapshots, nil)
		key := fmt.Sprintf("%064x", time.Now().UnixNano())
		raw = call("POST", "/retained-evidence", "/retained-evidence", map[string]any{"requestID": key, "captureID": frozen.Receipt.ID, "indexes": []int{0}, "tags": []string{"api:contract"}, "note": "Observed API contract"}, 201)
		var receipt retention.Receipt
		if err := json.Unmarshal(raw, &receipt); err != nil {
			t.Fatal(err)
		}
		call("GET", "/retained-evidence/"+receipt.Manifest.ID+"?awaitToken="+url.QueryEscape(receipt.AwaitToken), "/retained-evidence/{id}", nil, 200)

		call("GET", "/retained-evidence/catalog?tag=api:contract", "/retained-evidence/catalog", nil, 200)
		call("GET", "/retained-evidence/"+receipt.Manifest.ID+"/content?awaitToken="+url.QueryEscape(receipt.AwaitToken), "/retained-evidence/{id}/content", nil, 200)
		now := time.UnixMilli(60000)
		bank := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
		selector := exploration.TailSelector{Signal: "spans", Environment: "prod"}
		tracker := exploration.TailTracker{}
		rows, err := tracker.Observe(frozen.Explore.Rows)
		if err != nil {
			t.Fatal(err)
		}
		cap, err := bank.Freeze(exploration.TailMembershipBatch{Connection: key, Selector: selector, FromMS: 0, ToMS: 60000, ObservedAt: now, Rows: rows})
		if err != nil {
			t.Fatal(err)
		}
		tail, err := (retention.Store{URL: address}).PublishTail(context.Background(), bank, retention.TailIntent{Version: retention.TailEvidenceVersion, RequestID: fmt.Sprintf("%064x", time.Now().UnixNano()), Connection: key, Selector: selector, Refs: []exploration.TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}}, Tags: []string{"api:tail-contract"}, Note: "Exact observed contract"})
		if err != nil {
			t.Fatal(err)
		}
		call("GET", "/retained-evidence/"+tail.Manifest.ID+"/content?awaitToken="+url.QueryEscape(tail.AwaitToken), "/retained-evidence/{id}/content", nil, 200)
		call("GET", "/retained-evidence/catalog?origin=tail-selection&tag=api:tail-contract", "/retained-evidence/catalog", nil, 200)
		call("GET", "/retained-evidence/"+tail.Manifest.ID, "/retained-evidence/{id}", nil, 404)
		correction := map[string]any{"requestID": fmt.Sprintf("%064x", time.Now().UnixNano()), "awaitToken": receipt.AwaitToken, "target": workbench.MeaningTarget{Kind: "evidence", ID: receipt.Manifest.ID}, "parent": receipt.Annotation.ID, "tags": []string{"api:contract"}, "note": "Corrected interpretation", "state": "active"}
		raw = call("POST", "/retained-evidence/"+receipt.Manifest.ID+"/annotations", "/retained-evidence/{id}/annotations", correction, 201)
		var corrected retention.CorrectionReceipt
		json.Unmarshal(raw, &corrected)
		basis := "?awaitToken=" + url.QueryEscape(corrected.AwaitToken)
		call("GET", "/retained-evidence/"+receipt.Manifest.ID+"/annotations"+basis, "/retained-evidence/{id}/annotations", nil, 200)
		call("GET", "/retained-evidence/"+receipt.Manifest.ID+"/annotations/"+receipt.Annotation.ID+basis, "/retained-evidence/{id}/annotations/{revision}", nil, 200)
		pinned := workbench.DocumentInput{Title: "Pinned API review", Cards: []workbench.Card{{Kind: "retained", Reference: receipt.Manifest.ID, Review: receipt.Annotation.ID, ReadBasis: corrected.AwaitToken}}}
		raw = call("POST", "/documents", "/documents", map[string]any{"document": pinned}, 201)
		var pinnedDoc workbench.Document
		json.Unmarshal(raw, &pinnedDoc)
		mux = http.NewServeMux()
		registerWorkbench(mux, root, retentionConnection{}, resultSnapshots, nil)
		call("GET", "/documents/"+pinnedDoc.ID, "/documents/{id}", nil, 200)
	} else {
		t.Log("GOO_XTDB_TEST_URL absent: positive retention/correction/citation branch skipped")
	}
	if reads != before {
		t.Fatal("freeze/reopen/document/retention read telemetry")
	}
	if path := os.Getenv("GOO_API_OBSERVATIONS"); path != "" {
		b, err := json.MarshalIndent(observed, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d observed responses, source-free freeze/reopen/document/retention confirmed", len(observed))
}
