package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func TestLiveCounterEndpointWindowCoverageAndFreeze(t *testing.T) {
	def := measurements.Definition{APIVersion: measurements.Version, Name: "requests", Title: "Requests", Kind: "counter", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 10000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "otel.http", Metric: "http.server.requests", Unit: "{request}"}}
	row := func(start, end int64, value string, instance string) json.RawMessage {
		b, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": "api", "service.instance.id": instance, "deployment.environment.name": "prod"}, "scope": "otel.http", "version": "", "schema_url": "", "name": "http.server.requests", "unit": "{request}", "attributes": map[string]string{}, "start_ns": fmt.Sprint(start), "end_ns": fmt.Sprint(end), "value": value, "temporality": 1, "monotonic": true})
		return b
	}
	q := measurements.Question{Definition: def, FromMS: 100000, ToMS: 102000}
	baselineResult, e := measurements.Evaluate(q, []json.RawMessage{row(100000000000, 102000000000, "6", "one")}, time.Unix(103, 0))
	if e != nil {
		t.Fatal(e)
	}
	baseline := &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Baseline", CapturedAt: "2026-10-20T00:00:00Z", Question: CaptureQuestion{Kind: "measurement", Measurement: &q}}, Measurement: baselineResult}
	before, _ := json.Marshal(baseline)
	var reads atomic.Int64
	var skipNext atomic.Bool
	raw := []json.RawMessage{row(198000000123, 200000000123, "10", "one")}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		io.Copy(io.Discard, r.Body)
		if skipNext.Swap(false) {
			return
		}
		for _, v := range raw {
			fmt.Fprintln(w, string(v))
		}
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL, Client: server.Client()}
	now := time.Unix(201, 0)
	snapshot, pair, info, e := source.LiveCounterComparison(context.Background(), baseline, now)
	if e != nil {
		t.Fatal(e)
	}
	if pair.State != "comparable" || pair.NativeWindows[0].Baseline.Rate != "3" || pair.NativeWindows[0].Selection.Rate != "5" || snapshot.Question.Measurement.FromNS != "198000000123" || snapshot.Question.Measurement.ToNS != "200000000123" || reads.Load() != 2 || info.LagNS != "999999877" {
		t.Fatal(snapshot, pair, info, reads.Load())
	}
	after, _ := json.Marshal(baseline)
	if string(before) != string(after) {
		t.Fatal("baseline changed")
	}
	// Latest point has a bad boundary; an older valid window cannot be substituted.
	raw = append(raw, row(200000000123, 201000000123, "3", "one"))
	_, pair, _, e = source.LiveCounterComparison(context.Background(), baseline, time.Unix(202, 0))
	if e != nil || pair.State == "comparable" || pair.NativeWindows[0].RateDifference != "" {
		t.Fatal("boundary interpolated", e, pair)
	}
	raw = []json.RawMessage{row(198000000123, 200000000123, "10", "two")}
	_, pair, _, e = source.LiveCounterComparison(context.Background(), baseline, now)
	if e != nil || pair.State == "comparable" {
		t.Fatal("instances silently matched", e, pair)
	}
	raw = []json.RawMessage{row(198000000123, 200000000123, "10", "one")}
	_, pair, info, e = source.LiveCounterComparison(context.Background(), baseline, time.Unix(220, 0))
	if e != nil || pair.State != "unavailable" || len(info.Reasons) != 1 || pair.NativeWindows[0].RateDifference != "" {
		t.Fatal("stale window asserted live", e, pair, info)
	}
	// Discovery saw no endpoint, but a complete export arrived before the second
	// read. Historical evidence is retained; the live claim still fails closed.
	raw = []json.RawMessage{row(199000000000, 201000000000, "10", "one")}
	skipNext.Store(true)
	raced, unavailable, absent, err := source.LiveCounterComparison(context.Background(), baseline, now)
	if err != nil || unavailable.State != "unavailable" || absent.EndpointNS != "" || absent.LagNS != "" || len(absent.Reasons) != 1 || absent.Reasons[0] != "no-observed-endpoint" || unavailable.NativeWindows[0].RateDifference != "" || raced.Measurement.Counters.Series[0].State != "ok" {
		t.Fatal("missing discovery endpoint asserted live", err, unavailable, absent)
	}
	// A newly observed cumulative decrease remains the endpoint; no older
	// healthy cumulative window may be selected instead.
	cumulative := func(a, b int64, value string) json.RawMessage {
		return json.RawMessage(strings.Replace(string(row(a, b, value, "one")), `"temporality":1`, `"temporality":2`, 1))
	}
	cq := q
	cr, err := measurements.Evaluate(cq, []json.RawMessage{cumulative(90000000000, 100000000000, "2"), cumulative(90000000000, 102000000000, "8")}, time.Unix(103, 0))
	if err != nil {
		t.Fatal(err)
	}
	cb := *baseline
	cb.Measurement = cr
	raw = []json.RawMessage{cumulative(190000000000, 198000000123, "100"), cumulative(190000000000, 200000000123, "90")}
	resetSnapshot, resetPair, _, err := source.LiveCounterComparison(context.Background(), &cb, now)
	if err != nil || resetPair.State == "comparable" || resetSnapshot.Question.Measurement.ToNS != "200000000123" {
		t.Fatal("reset fallback", err, resetPair)
	}
	if !strings.Contains(strings.Join(resetPair.NativeWindows[0].Selection.Reasons, ","), "counter-decrease") {
		t.Fatal(resetPair.NativeWindows)
	}
	server.Close()
	store := CaptureStore{Dir: t.TempDir()}
	frozen, e := store.FreezeSnapshot(snapshot)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := store.Load(frozen.ID)
	if e != nil || !reflect.DeepEqual(loaded.Measurement, snapshot.Measurement) {
		t.Fatal("freeze reread or changed exact counter snapshot", e)
	}
	if got := metrics.CompareCounterWindows(baseline.Measurement.Counters, loaded.Measurement.Counters); len(got) != 1 || got[0].RateDifference != "2" {
		t.Fatal(got)
	}
}

func TestNativeCounterBaselineValidationBeforeSource(t *testing.T) {
	for _, b := range []*SavedCapture{nil, {}, {Measurement: &measurements.Result{}, CaptureReceipt: CaptureReceipt{Question: CaptureQuestion{}}}} {
		_, _, _, e := (Source{Endpoint: "http://invalid"}).LiveCounterComparison(context.Background(), b, time.Now())
		if e == nil || strings.Contains(e.Error(), "http") {
			t.Fatal(e)
		}
	}
}
