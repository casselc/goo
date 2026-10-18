package measurements

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiscoveryExactOptionalIdentity(t *testing.T) {
	q := DiscoveryQuestion{Environment: "local", Service: "independent", FromMS: 1000, ToMS: 2000}
	rows := []json.RawMessage{json.RawMessage(`{"kind":"counter","scope":"ordinary","version":"","schemaURL":"","metric":"jobs","unit":"{job}","temporality":1,"monotonic":true,"dimensions":["outcome"]}`)}
	r, e := discoverRows(q, rows)
	if e != nil || len(r.Instruments) != 1 || !r.Instruments[0].Supported {
		t.Fatalf("%+v %v", r, e)
	}
	d := r.Instruments[0].Definition
	p, e := Compile(Question{Definition: *d, FromMS: 1000, ToMS: 2000})
	if e != nil {
		t.Fatal(e)
	}
	if p.Parameters["p5"] != "" || p.Parameters["p6"] != "" {
		t.Fatal("invented optional metadata", p)
	}
	rows = append(rows, json.RawMessage(strings.ReplaceAll(string(rows[0]), `"version":""`, `"version":"2"`)))
	r, e = discoverRows(q, rows)
	if e != nil || len(r.Instruments) != 2 {
		t.Fatal("merged distinct identities", r, e)
	}
}
func TestDiscoveryBoundedTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("param_service") != "x' OR 1=1" || r.URL.Query().Get("readonly") != "1" {
			t.Error("typed parameters missing")
		}
		for i := 0; i <= RowLimit; i++ {
			w.Write([]byte("{}\n"))
		}
	}))
	defer server.Close()
	_, e := (Source{Endpoint: server.URL}).Discover(context.Background(), DiscoveryQuestion{Environment: "local", Service: "x' OR 1=1", FromMS: 1000, ToMS: 2000})
	if e == nil {
		t.Fatal("partial catalog accepted")
	}
}
