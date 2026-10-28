package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestLiveComparisonHTTPExactFreezeAndBoundedCapabilities(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	var reads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "count()") {
			fmt.Fprintln(w, `{"count":"105"}`)
		} else if !strings.Contains(string(b), "SELECT DISTINCT") {
			fmt.Fprintln(w, `{"timestamp_ns":"1000000001","duration_ns":"0"}`)
		}
	}))
	defer source.Close()
	mux := http.NewServeMux()
	registerCaptures(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, t.TempDir(), resultSnapshots)
	call := func(path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, bytes.NewReader(b)))
		return w
	}
	initial := call("/captures", map[string]any{"title": "Baseline", "question": map[string]any{"kind": "explore", "explore": map[string]any{"signal": "spans", "environment": "prod", "fromMs": 1000, "toMs": 61000, "limit": 100}}})
	var baseline exploration.CapturePage
	if initial.Code != 201 {
		t.Fatal(initial.Body.String())
	}
	json.Unmarshal(initial.Body.Bytes(), &baseline)
	for _, in := range []any{map[string]string{}, map[string]string{"baseline": baseline.Receipt.ID, "freeze": "x"}, map[string]string{"baseline": baseline.Receipt.ID, "extra": "x"}} {
		before := reads.Load()
		if w := call("/captures/live", in); w.Code != 400 || reads.Load() != before {
			t.Fatal("invalid request read source", w.Code)
		}
	}
	var first struct {
		Token      string
		Comparison exploration.EvidenceComparison
	}
	w := call("/captures/live", map[string]string{"baseline": baseline.Receipt.ID})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var unbucketed map[string]any
	json.Unmarshal(w.Body.Bytes(), &unbucketed)
	if _, ok := unbucketed["recordSeries"]; ok {
		t.Fatal("unrequested buckets added")
	}
	json.Unmarshal(w.Body.Bytes(), &first)
	if len(first.Token) != 64 || first.Comparison.Baseline.Receipt.ID != baseline.Receipt.ID {
		t.Fatal(first)
	}
	for i := 0; i < 15; i++ {
		if extra := call("/captures/live", map[string]string{"baseline": baseline.Receipt.ID}); extra.Code != 200 {
			t.Fatal("draft admission", i, extra.Code, extra.Body.String())
		}
	}
	if extra := call("/captures/live", map[string]string{"baseline": baseline.Receipt.ID}); extra.Code != 429 {
		t.Fatal("unbounded drafts", extra.Code)
	}
	oldToken := first.Token
	w = call("/captures/live", map[string]string{"baseline": baseline.Receipt.ID, "replace": oldToken})
	if w.Code != 200 {
		t.Fatal("replacement did not free budget", w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &first)
	if old := call("/captures/live", map[string]string{"freeze": oldToken}); old.Code != 410 {
		t.Fatal("replaced capability retained", old.Code)
	}
	before := reads.Load()
	source.Close()
	var frozen exploration.CapturePage
	w = call("/captures/live", map[string]string{"freeze": first.Token})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &frozen)
	if frozen.Explore.Count != first.Comparison.Selection.Count || frozen.Explore.CapturedAt != first.Comparison.Selection.ReadAt || reads.Load() != before {
		t.Fatal("freeze changed source basis")
	}
	w = call("/captures/live", map[string]string{"freeze": first.Token})
	var again exploration.CapturePage
	json.Unmarshal(w.Body.Bytes(), &again)
	if again.Receipt.ID != frozen.Receipt.ID {
		t.Fatal("freeze not idempotent")
	}
	if w = call("/captures/live", map[string]string{"freeze": "unknown"}); w.Code != 410 {
		t.Fatal(w.Code)
	}
}

func TestLiveComparisonRetainedSeriesFreezeWithoutReread(t *testing.T) {
	resultSnapshots := &exploration.ResultSnapshots{}
	var reads, seriesReads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		sql, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(string(sql), "GROUP BY from_ms"):
			seriesReads.Add(1)
			fmt.Fprintf(w, `{"from_ms":%q,"count":"5"}`+"\n", r.URL.Query().Get("param_from"))
		case strings.Contains(string(sql), "count()"):
			fmt.Fprintln(w, `{"count":"5"}`)
		case !strings.Contains(string(sql), "SELECT DISTINCT"):
			fmt.Fprintln(w, `{"timestamp_ns":"1000000001","duration_ns":"0"}`)
		}
	}))
	defer source.Close()
	mux := http.NewServeMux()
	registerCaptures(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, t.TempDir(), resultSnapshots)
	call := func(body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/captures/live", bytes.NewReader(b)))
		return w
	}
	b, _ := json.Marshal(map[string]any{"title": "Bucket baseline", "question": map[string]any{"kind": "explore", "explore": map[string]any{"signal": "spans", "environment": "prod", "fromMs": 1000, "toMs": 61000, "limit": 1, "bucketMs": 10000}}})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/captures", bytes.NewReader(b)))
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var baseline exploration.CapturePage
	json.Unmarshal(w.Body.Bytes(), &baseline)
	beforeSeries := seriesReads.Load()
	w = call(map[string]string{"baseline": baseline.Receipt.ID})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var draft struct {
		Token        string
		RecordSeries exploration.RecordSeriesPair
	}
	if e := json.Unmarshal(w.Body.Bytes(), &draft); e != nil {
		t.Fatal(e)
	}
	if seriesReads.Load() != beforeSeries+1 || len(draft.RecordSeries.Selection) != 6 || len(draft.RecordSeries.Baseline) != 6 || draft.RecordSeries.Selection[0].Count != "5" || draft.RecordSeries.Selection[1].State != "no_data" {
		t.Fatal("series not retained full aggregates", draft)
	}
	if !reflect.DeepEqual(draft.RecordSeries.Baseline, baseline.Explore.TimeSeries) {
		t.Fatal("baseline series changed")
	}
	if path := os.Getenv("GOO_LIVE_SERIES_RESPONSE"); path != "" {
		if e := os.WriteFile(path, w.Body.Bytes(), 0600); e != nil {
			t.Fatal(e)
		}
	}
	before := reads.Load()
	source.Close()
	w = call(map[string]string{"freeze": draft.Token})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var frozen exploration.CapturePage
	json.Unmarshal(w.Body.Bytes(), &frozen)
	if reads.Load() != before || !reflect.DeepEqual(draft.RecordSeries.Selection, frozen.Explore.TimeSeries) {
		t.Fatal("freeze changed displayed series")
	}
}
