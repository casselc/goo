package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func durationViewFixture() *Result {
	q := basic()
	q.GroupBy = []Field{{Kind: "column", Key: "duration_bucket"}}
	return &Result{Question: q, Count: "99", State: "observed", Basis: strings.Repeat("a", 64), CapturedAt: "2026-10-30T01:00:00Z", Groups: []json.RawMessage{json.RawMessage(`{"g0":"[0,1) ms","count":"2"}`), json.RawMessage(`{"g0":"[1000,+inf) ms","count":"3"}`)}}
}
func TestDurationHistogramViewExactSameReadPartition(t *testing.T) {
	r := durationViewFixture()
	d := durationHistogramViewDataset(r, true)
	if d == nil || d.Histogram.Kind != "span-duration" || d.Histogram.Total != "5" || d.Histogram.Unit != "ns" || len(d.Rows) != 8 || *d.Rows[7].Values["to"] != "+inf" || *d.Rows[1].Values["count"] != "0" || d.Rows[1].State != "no_data" || d.Completeness == nil {
		t.Fatal("duration geometry or same-read denominator lost", d)
	}
	raw, e := json.Marshal(d)
	if e != nil {
		t.Fatal(e)
	}
	var wire map[string]any
	json.Unmarshal(raw, &wire)
	metadata := wire["histogram"].(map[string]any)
	if _, ok := metadata["binding"]; ok {
		t.Fatal("fake numeric binding")
	}
	if _, ok := metadata["support"]; ok {
		t.Fatal("fake parsing support")
	}
	fixtures := map[string]*ViewDataset{"observed": d}
	r.Question.Summary = "span-duration"
	r.Question.SpanStatus = true
	r.Question.BucketMS = 1000
	if durationHistogramViewDataset(r, true) == nil {
		t.Fatal("unrelated measure/time context blocked same-read bins")
	}
	r = durationViewFixture()
	r.State = "no_data"
	r.Count = "0"
	r.Groups = nil
	d = durationHistogramViewDataset(r, true)
	if d == nil || d.Histogram.Total != "0" || len(d.Rows) != 8 {
		t.Fatal("exhaustive empty distribution lost")
	}
	fixtures["empty"] = d
	if path := os.Getenv("GOO_DURATION_HISTOGRAM_FIXTURES"); path != "" {
		b, e := json.Marshal(fixtures)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
func TestDurationHistogramViewRefusesUnprovedOrMalformedGroups(t *testing.T) {
	if durationHistogramViewDataset(durationViewFixture(), false) != nil {
		t.Fatal("nonexhaustive filled zeroes")
	}
	for _, mutate := range []func(*Result){
		func(r *Result) { r.State = "changed_between_reads" },
		func(r *Result) { r.State = "no_data" },
		func(r *Result) { r.Groups = nil }, func(r *Result) { r.Question.Signal = "logs" }, func(r *Result) { r.CapturedAt = "" }, func(r *Result) { r.Question.GroupTimeSeries = true }, func(r *Result) {
			r.Question.GroupBy = append(r.Question.GroupBy, Field{Kind: "column", Key: "service"})
		}, func(r *Result) { r.Groups = append(r.Groups, r.Groups[0]) }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":null,"count":"1"}`)} }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"other","count":"1"}`)} }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"[0,1) ms","count":"01"}`)} }, func(r *Result) {
			r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"[0,1) ms","count":"18446744073709551615"}`), json.RawMessage(`{"g0":"[1,5) ms","count":"1"}`)}
		},
	} {
		r := durationViewFixture()
		mutate(r)
		if durationHistogramViewDataset(r, true) != nil {
			t.Fatal("unsupported duration histogram accepted", r)
		}
	}
	r := durationViewFixture()
	r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"[0,1) ms","count":"18446744073709551615"}`)}
	if d := durationHistogramViewDataset(r, true); d == nil || d.Histogram.Total != "18446744073709551615" {
		t.Fatal("uint64 precision lost")
	}
}

func TestDurationHistogramSourceUsesExistingGroupRead(t *testing.T) {
	q := basic()
	q.GroupBy = []Field{{Kind: "column", Key: "duration_bucket"}}
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		switch {
		case strings.Contains(sql, "GROUP BY"):
			fmt.Fprintln(w, `{"g0":"[0,1) ms","count":"1"}`)
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"1","duration_bucket":"[0,1) ms"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			return
		default:
			fmt.Fprintln(w, `{"count":"1"}`)
		}
	}))
	defer server.Close()
	r, e := (&Source{Endpoint: server.URL}).Fetch(context.Background(), q)
	if e != nil || reads != 4 || r.HistogramDataset == nil || r.HistogramDataset.Histogram.Total != "1" || r.HistogramDataset.Completeness == nil {
		t.Fatal("duration projection lost existing group read", reads, e, r)
	}
}
