package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestKeptObservationHTTPSourceFreeContracts(t *testing.T) {
	now := time.UnixMilli(200000).UTC()
	bank := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
	tracker := exploration.TailTracker{}
	rows, err := tracker.Observe([]json.RawMessage{json.RawMessage(`{"timestamp_ns":"150000000000","Body":"original","ResourceAttributes":{"deployment.environment.name":"prod"}}`)})
	if err != nil {
		t.Fatal(err)
	}
	selector := exploration.TailSelector{Signal: "logs", Environment: "prod"}
	cap, err := bank.Freeze(exploration.TailMembershipBatch{Connection: strings.Repeat("a", 64), Selector: selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows})
	if err != nil {
		t.Fatal(err)
	}
	in := exploration.KeepTailIntent{Version: exploration.KeepTailVersion, RequestID: strings.Repeat("b", 64), Connection: strings.Repeat("a", 64), Selector: selector, Refs: []exploration.TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}}}
	root := t.TempDir()
	mux := http.NewServeMux()
	registerKeptObservations(mux, root, bank, nil)
	call := func(method, path, body string, want int) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Header().Get("Cache-Control"), "no-store") {
			t.Fatal("missing no-store")
		}
		return w.Body.Bytes()
	}
	raw, _ := json.Marshal(in)
	first := call("POST", "/kept-observations/tail", string(raw), 201)
	now = now.Add(3 * time.Minute)
	mux = http.NewServeMux()
	registerKeptObservations(mux, root, nil, nil)
	if string(call("POST", "/kept-observations/tail", string(raw), 201)) != string(first) {
		t.Fatal("restart retry changed receipt")
	}
	if string(call("GET", "/kept-observations/"+in.RequestID, "", 200)) != string(first) {
		t.Fatal("exact reopen changed receipt")
	}
	in.Selector.Environment = "other"
	changed, _ := json.Marshal(in)
	call("POST", "/kept-observations/tail", string(changed), 409)
	in.RequestID = strings.Repeat("c", 64)
	changed, _ = json.Marshal(in)
	call("POST", "/kept-observations/tail", string(changed), 410)
	call("POST", "/kept-observations/tail", strings.TrimSuffix(string(raw), "}")+`,"rows":[]}`, 400)
	call("POST", "/kept-observations/tail", string(raw)+` {}`, 400)
	call("POST", "/kept-observations/tail?refresh=true", string(raw), 400)
	call("GET", "/kept-observations/"+strings.Repeat("d", 64), "", 404)
}

func TestKeptRecordHTTPReferenceOnlyAndRestart(t *testing.T) {
	bank := new(exploration.ResultSnapshots)
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 100000, ToMS: 160000, Limit: 100, Offset: 100}
	token, e := bank.Freeze(&exploration.Result{Question: q, Rows: []json.RawMessage{json.RawMessage(`{"timestamp_ns":"150000000000","Body":"original","ResourceAttributes":{"deployment.environment.name":"prod"}}`)}, CapturedAt: time.UnixMilli(200000).UTC().Format(time.RFC3339Nano), Basis: strings.Repeat("a", 64), State: "observed"})
	if e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	mux := http.NewServeMux()
	registerKeptObservations(mux, root, nil, bank)
	in := exploration.KeepRecordsIntent{Version: exploration.KeepRecordsVersion, RequestID: strings.Repeat("b", 64), ResultToken: token, Indexes: []int{0}}
	raw, _ := json.Marshal(in)
	call := func(method, path, body string, want int) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	first := call("POST", "/kept-observations/records", string(raw), 201)
	if strings.Contains(string(first), token) {
		t.Fatal("live token leaked")
	}
	mux = http.NewServeMux()
	registerKeptObservations(mux, root, nil, nil)
	if string(call("POST", "/kept-observations/records", string(raw), 201)) != string(first) {
		t.Fatal("restart changed winner")
	}
	if string(call("GET", "/kept-observations/"+in.RequestID, "", 200)) != string(first) {
		t.Fatal("reopen changed receipt")
	}
	call("POST", "/kept-observations/records", strings.TrimSuffix(string(raw), "}")+`,"rows":[]}`, 400)
	in.RequestID = strings.Repeat("c", 64)
	raw, _ = json.Marshal(in)
	call("POST", "/kept-observations/records", string(raw), 410)
	in.CaptureID = strings.Repeat("a", 64)
	raw, _ = json.Marshal(in)
	call("POST", "/kept-observations/records", string(raw), 400)
}

func TestKeptDiscoveryHTTPAndExplicitRepair(t *testing.T) {
	bank := new(exploration.ResultSnapshots)
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 100000, ToMS: 160000, Limit: 100}
	token, e := bank.Freeze(&exploration.Result{Question: q, Rows: []json.RawMessage{json.RawMessage(`{"timestamp_ns":"150000000000","Body":"original","ResourceAttributes":{"deployment.environment.name":"prod"}}`)}, CapturedAt: time.UnixMilli(200000).UTC().Format(time.RFC3339Nano), Basis: strings.Repeat("a", 64), State: "observed"})
	if e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	mux := http.NewServeMux()
	registerKeptObservations(mux, root, nil, bank)
	call := func(method, path, body string, want int) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Header().Get("Cache-Control"), "no-store") {
			t.Fatal("missing no-store")
		}
		return w.Body.Bytes()
	}
	in := exploration.KeepRecordsIntent{Version: exploration.KeepRecordsVersion, RequestID: strings.Repeat("b", 64), ResultToken: token, Indexes: []int{0}}
	raw, _ := json.Marshal(in)
	call("POST", "/kept-observations/records", string(raw), 201)
	var page exploration.KeptSearchPage
	json.Unmarshal(call("GET", "/kept-observations/search?signal=logs", "", 200), &page)
	if len(page.Items) != 1 || page.Items[0].ID != in.RequestID {
		t.Fatal(page)
	}
	call("GET", "/kept-observations/search?signal=metrics", "", 400)
	call("GET", "/kept-observations/search?signal=%ZZ", "", 400)
	call("GET", "/kept-observations/search?signal=logs&signal=spans", "", 400)
	call("GET", "/kept-observations/search?cursor=invalid", "", 409)
	os.RemoveAll(filepath.Join(root, "kept-observations", ".catalog-v1"))
	call("GET", "/kept-observations/search", "", 503)
	call("GET", "/kept-observations/"+in.RequestID, "", 200)
	call("POST", "/kept-observations/repair", `{}`, 200)
	call("GET", "/kept-observations/search", "", 200)
	call("POST", "/kept-observations/repair", `{"rows":[]}`, 400)
}
