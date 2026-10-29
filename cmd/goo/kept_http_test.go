package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"net/http"
	"net/http/httptest"
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
	registerKeptObservations(mux, root, bank)
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
	registerKeptObservations(mux, root, nil)
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
