package exploration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTailMultisetLateAndCanonicalIdentity(t *testing.T) {
	tracker := TailTracker{}
	a := json.RawMessage(`{"timestamp_ns":"100","Body":"same","Attributes":{}}`)
	reordered := json.RawMessage(`{"Body":"same","Attributes":{},"timestamp_ns":"100"}`)
	rows, e := tracker.Observe([]json.RawMessage{a, a})
	if e != nil || len(rows) != 2 || rows[0].Key == rows[1].Key {
		t.Fatal(rows, e)
	}
	rows, e = tracker.Observe([]json.RawMessage{reordered, a})
	if e != nil || len(rows) != 0 {
		t.Fatal(rows, e)
	}
	rows, e = tracker.Observe([]json.RawMessage{a, a, a, json.RawMessage(`{"timestamp_ns":"99","Body":"late"}`)})
	if e != nil || len(rows) != 2 || !rows[1].Late || rows[0].Late {
		t.Fatal(rows, e)
	}
	if _, e = tracker.Observe([]json.RawMessage{json.RawMessage(`{"Body":"missing clock"}`)}); e == nil {
		t.Fatal("missing clock accepted")
	}
}
func TestTailOnlyBoundedRowRead(t *testing.T) {
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		query = string(b)
		if r.URL.Query().Get("param_environment") != "prod" {
			t.Error("selector lost")
		}
		io.WriteString(w, "{\"timestamp_ns\":\"200\",\"Body\":\"newest\"}\n{\"timestamp_ns\":\"100\",\"Body\":\"oldest\"}\n")
	}))
	defer srv.Close()
	rows, limited, e := (Source{Endpoint: srv.URL, Client: srv.Client()}).PollTail(context.Background(), TailSelector{Environment: "prod"}, 60001)
	if e != nil || limited || len(rows) != 2 {
		t.Fatal(rows, limited, e)
	}
	if !strings.Contains(string(rows[0]), "oldest") || !strings.Contains(string(rows[1]), "newest") {
		t.Fatal("bounded rows must retain newest at the end", rows)
	}
	if !strings.Contains(query, "ORDER BY Timestamp DESC,") || !strings.Contains(query, "LIMIT 1001 OFFSET 0") || strings.Contains(query, "count()") || strings.Contains(query, "otel_traces") {
		t.Fatal(query)
	}
}
