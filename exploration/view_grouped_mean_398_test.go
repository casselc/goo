package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func groupedMeanFixture398(t *testing.T) *Result {
	t.Helper()
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 4500, Limit: 3, GroupBy: []Field{{Kind: "attribute", Key: "route"}}, BucketMS: 1000, GroupTimeSeries: true, Summary: "span-duration"}
	groups := []struct {
		value     *string
		from      string
		durations []uint64
	}{{nil, "1000", []uint64{0, 1, 1}}, {viewText(""), "2000", []uint64{0}}, {viewText("雪"), "4000", []uint64{7, 8}}}
	cells, intervals, aggregate := []json.RawMessage{}, []json.RawMessage{}, []json.RawMessage{}
	all := []uint64{}
	for _, group := range groups {
		d := ReferenceDurationSummary(group.durations)
		row := map[string]any{"g0": group.value, "from_ms": group.from, "count": d.Count, "sum_ns": d.SumNS, "min_ns": d.MinNS, "max_ns": d.MaxNS, "quantiles_ns": []string{d.P50NS, d.P95NS, d.P99NS}}
		raw, _ := json.Marshal(row)
		cells = append(cells, raw)
		delete(row, "g0")
		raw, _ = json.Marshal(row)
		intervals = append(intervals, raw)
		raw, _ = json.Marshal(map[string]any{"g0": group.value, "count": d.Count, "duration": d})
		aggregate = append(aggregate, raw)
		all = append(all, group.durations...)
	}
	grouped, e := DecodeGroupedTimeSeries(q, cells)
	if e != nil {
		t.Fatal(e)
	}
	overall, e := DecodeTimeSeries(q, intervals)
	if e != nil {
		t.Fatal(e)
	}
	d := ReferenceDurationSummary(all)
	return &Result{Question: q, Basis: strings.Repeat("a", 64), CapturedAt: "2026-11-10T00:00:00Z", Count: d.Count, Duration: &d, State: "observed", GroupedTimeSeries: grouped, TimeSeries: overall, Groups: aggregate}
}

func TestGroupedMean398ExactSupportAndZero(t *testing.T) {
	r := groupedMeanFixture398(t)
	if !GroupedMeanDurationTimelineQuery(r.Question) || GroupedCountTimelineQuery(r.Question) {
		t.Fatal("mean/count admission conflated")
	}
	d := groupedTimeViewDataset(r, true)
	if d == nil || len(d.Fields) != 6 || len(d.Rows) != 12 || d.Completeness == nil || d.Quality.Sampling != "unknown" {
		t.Fatal("grouped mean unavailable", d)
	}
	if d.Fields[4].Key != "duration.sum" || d.Fields[5].Key != "duration.mean" || d.Fields[5].Unit != "ns" || d.Fields[5].Aggregation.Method != "exact-rational" {
		t.Fatal("exact mean field contract changed")
	}
	positive, empty := 0, 0
	for _, row := range d.Rows {
		if *row.Values["fromMs"] == "4000" && *row.Values["toMs"] != "4500" {
			t.Fatal("final interval widened")
		}
		if row.State == "no_data" {
			empty++
			if *row.Values["count"] != "0" || row.Values["duration.mean"] != nil || row.Values["duration.sum"] != nil || row.Cells["duration.mean"].State != "no_data" || row.Cells["duration.mean"].Support.Valid != "0" {
				t.Fatal("empty latency fabricated", row)
			}
			continue
		}
		positive++
		want := map[string]string{"1000": "2/3", "2000": "0", "4000": "15/2"}[*row.Values["fromMs"]]
		if *row.Values["duration.mean"] != want || row.Cells["duration.mean"].State != "derived" || row.Cells["duration.sum"].Support.Valid != *row.Values["count"] || row.Cells["duration.mean"].Support.Total != *row.Values["count"] {
			t.Fatal("exact support or mean changed", row)
		}
	}
	if positive != 3 || empty != 9 {
		t.Fatal("dense matrix changed", positive, empty)
	}
	before, _ := json.Marshal(d)
	r.GroupedTimeSeries[0].Points[0].Duration.SumNS = "999"
	after, _ := json.Marshal(d)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("projection retained caller state")
	}
}

func TestGroupedMean398ChangedAndInvalidSupport(t *testing.T) {
	for _, change := range []func(*Result){func(r *Result) { r.State = "changed_between_reads" }, func(r *Result) { r.Duration.SumNS = "18" }, func(r *Result) { r.Count = "7"; r.Duration.Count = "7" }} {
		r := groupedMeanFixture398(t)
		change(r)
		d := groupedTimeViewDataset(r, true)
		if d == nil || d.Completeness != nil || d.Quality.Extent != "unknown" {
			t.Fatal("changed read gained certainty", d)
		}
		unknown := 0
		for _, row := range d.Rows {
			if row.State == "unknown" {
				unknown++
				if row.Values["count"] != nil || row.Values["duration.mean"] != nil || row.Values["duration.sum"] != nil || row.Cells["duration.mean"].State != "unknown" || row.Cells["duration.mean"].Support != nil {
					t.Fatal("unknown absence became duration", row)
				}
			}
		}
		if unknown != 9 {
			t.Fatal("unknown absences changed", unknown)
		}
	}
	for _, change := range []func(*Result){func(r *Result) { r.Duration = nil }, func(r *Result) { r.GroupedTimeSeries[0].Points[0].Duration.Count = "1" }, func(r *Result) { r.GroupedTimeSeries[0].Points[0].Duration.SumNS = "99" }, func(r *Result) { r.TimeSeries[0].Duration = nil }, func(r *Result) {
		r.Groups[0] = json.RawMessage(`{"g0":null,"count":"3","duration":{"state":"observed","count":"2","sumNs":"2","minNs":"0","maxNs":"1","p50Ns":"1","p95Ns":"1","p99Ns":"1"}}`)
	}} {
		r := groupedMeanFixture398(t)
		change(r)
		if groupedTimeViewDataset(r, true) != nil {
			t.Fatal("invalid raw duration support admitted")
		}
	}
	r := groupedMeanFixture398(t)
	r.State = "no_data"
	r.Count = "0"
	d0 := ReferenceDurationSummary(nil)
	r.Duration = &d0
	r.Groups = nil
	r.GroupedTimeSeries = nil
	r.TimeSeries, _ = DecodeTimeSeries(r.Question, nil)
	d := groupedTimeViewDataset(r, true)
	if d == nil || len(d.Rows) != 0 || d.Completeness == nil {
		t.Fatal("no-data fabricated a category")
	}
}
