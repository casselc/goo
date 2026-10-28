package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
)

func issuedTailHTTP(t *testing.T, bank *exploration.TailMembershipSnapshots, states ...string) retention.TailIntent {
	t.Helper()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"timestamp_ns":"%d","Body":"exact occurrence","value":9007199254740993}`+"\n", time.Now().Add(-time.Second).UnixNano())
	}))
	defer source.Close()
	mux := http.NewServeMux()
	registerTail(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, bank)
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", server.URL+"/tail", strings.NewReader(`{"environment":"prod"}`))
	res, e := server.Client().Do(req)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	var connection string
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var frame struct {
			Connection      string                               `json:"connection"`
			Rows            []exploration.TailRow                `json:"rows"`
			MembershipState string                               `json:"membershipState"`
			Membership      exploration.TailMembershipCapability `json:"membership"`
		}
		if e = json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &frame); e != nil {
			t.Fatal(e)
		}
		if frame.Connection != "" {
			connection = frame.Connection
			continue
		}
		wantState := "available"
		if len(states) > 0 {
			wantState = states[0]
		}
		if frame.MembershipState != wantState || len(frame.Rows) != 1 {
			t.Fatalf("unissued batch: %s", line)
		}
		return retention.TailIntent{Version: retention.TailEvidenceVersion, RequestID: strings.Repeat("d", 64), Connection: connection, Selector: exploration.TailSelector{Environment: "prod"}, Refs: []exploration.TailMembershipRef{{Token: frame.Membership.Token, Key: frame.Rows[0].Key}}, Tags: []string{"http:tail"}}
	}
	t.Fatal("no batch", scanner.Err())
	return retention.TailIntent{}
}
func TestTailHTTPIssuesServerOwnedMembership(t *testing.T) {
	bank := exploration.NewTailMembershipSnapshots(time.Now)
	in := issuedTailHTTP(t, bank)
	draft, e := bank.Resolve(in.Connection, in.Selector, in.Refs)
	if e != nil || len(draft.Members) != 1 || !strings.Contains(string(draft.Members[0].Record), "9007199254740993") {
		t.Fatal(draft, e)
	}
	if _, e = bank.Resolve(strings.Repeat("e", 64), in.Selector, in.Refs); e == nil {
		t.Fatal("rebound connection")
	}
}
func TestTailPublicationHTTPValidation(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	mux := http.NewServeMux()
	registerRetention(mux, t.TempDir(), retentionConnection{}, resultSnapshots, nil)
	for _, body := range []string{`{"rows":[]}`, `{}`, `{"version":"goo/tail-evidence/v1alpha1","refs":[]}`} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/retained-evidence/tail", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
func TestXTDBTailHTTPCommittedRecoveryWithoutBank(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	bank := exploration.NewTailMembershipSnapshots(time.Now)
	in := issuedTailHTTP(t, bank)
	in.RequestID = fmt.Sprintf("%064x", time.Now().UnixNano())
	raw, _ := json.Marshal(in)
	mux := http.NewServeMux()
	registerRetention(mux, t.TempDir(), retentionConnection{URL: url}, resultSnapshots, bank)
	publish := func(m *http.ServeMux, body []byte) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("POST", "/retained-evidence/tail", strings.NewReader(string(body))))
		return w
	}
	first := publish(mux, raw)
	if first.Code != 201 {
		t.Fatal(first.Code, first.Body.String())
	}
	restarted := http.NewServeMux()
	registerRetention(restarted, t.TempDir(), retentionConnection{URL: url}, resultSnapshots, nil)
	recovered := publish(restarted, raw)
	var a, b retention.TailReceipt
	json.Unmarshal(first.Body.Bytes(), &a)
	json.Unmarshal(recovered.Body.Bytes(), &b)
	if recovered.Code != 201 || a.Manifest.ID != b.Manifest.ID || a.Annotation.ID != b.Annotation.ID || !a.RetainedAt.Equal(b.RetainedAt) {
		t.Fatal(recovered.Code, recovered.Body.String())
	}
	in.Note = "changed intent"
	changed, _ := json.Marshal(in)
	if w := publish(restarted, changed); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	in.RequestID = fmt.Sprintf("%064x", time.Now().UnixNano())
	absent, _ := json.Marshal(in)
	if w := publish(restarted, absent); w.Code != 410 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestTailHTTPCapacityRefusesProofButKeepsObservations(t *testing.T) {
	now := time.Now().UTC()
	bank := exploration.NewTailMembershipSnapshots(time.Now)
	tracker := exploration.TailTracker{}
	rows, e := tracker.Observe([]json.RawMessage{json.RawMessage(fmt.Sprintf(`{"timestamp_ns":"%d","Body":"existing"}`, now.Add(-time.Second).UnixNano()))})
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < exploration.TailMembershipMaxBatches; i++ {
		_, e = bank.Freeze(exploration.TailMembershipBatch{Connection: fmt.Sprintf("%064x", i), Selector: exploration.TailSelector{Environment: "prod"}, FromMS: now.UnixMilli() - 60000, ToMS: now.UnixMilli(), ObservedAt: now, Rows: rows})
		if e != nil {
			t.Fatal(e)
		}
	}
	in := issuedTailHTTP(t, bank, "unavailable")
	if in.Refs[0].Token != "" {
		t.Fatal("capacity issued a false capability")
	}
}
