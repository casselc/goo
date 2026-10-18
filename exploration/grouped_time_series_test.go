package exploration

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func seriesFixture() (Query, []json.RawMessage) {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 4000, Limit: 1, Summary: "span-duration", BucketMS: 1000, GroupTimeSeries: true, GroupBy: []Field{{Kind: "attribute", Key: "http.route"}}}
	raw := []json.RawMessage{
		json.RawMessage(`{"g0":"/a","from_ms":"1000","count":"2","sum_ns":"30","min_ns":"10","max_ns":"20","quantiles_ns":["20","20","20"]}`),
		json.RawMessage(`{"g0":"/a","from_ms":"3000","count":"1","sum_ns":"40","min_ns":"40","max_ns":"40","quantiles_ns":["40","40","40"]}`),
		json.RawMessage(`{"g0":null,"from_ms":"2000","count":"1","sum_ns":"50","min_ns":"50","max_ns":"50","quantiles_ns":["50","50","50"]}`),
		json.RawMessage(`{"g0":"","from_ms":"1000","count":"1","sum_ns":"0","min_ns":"0","max_ns":"0","quantiles_ns":["0","0","0"]}`)}
	return q, raw
}
func TestSplitSeriesTypedMeasuresAndCompatibility(t *testing.T) {
	q, raw := seriesFixture()
	groups, e := DecodeGroupedTimeSeries(q, raw)
	if e != nil || len(groups) != 3 {
		t.Fatal(groups, e)
	}
	counts := map[string][]string{}
	for _, g := range groups {
		for _, p := range g.Points {
			counts[groupKey(g.Values)] = append(counts[groupKey(g.Values)], p.Count)
		}
	}
	if !reflect.DeepEqual(counts[`["/a"]`], []string{"2", "0", "1"}) || !reflect.DeepEqual(counts[`[null]`], []string{"0", "1", "0"}) || !reflect.DeepEqual(counts[`[""]`], []string{"1", "0", "0"}) {
		t.Fatal("lost missing/empty or interval meaning", counts)
	}
	typed, e := q.RecordMeasures()
	if e != nil || typed.Scope.Grain != "span-record" || len(typed.Measures) != 8 {
		t.Fatal(typed, e)
	}
	for _, m := range typed.Measures {
		if m.Operator == "quantile" && m.Merge != "raw-values-required" {
			t.Fatal("scalar percentile merge")
		}
		if m.Input != nil && (*m.Input != (Field{Kind: "column", Key: "duration_ns"}) || m.Unit != "ns") {
			t.Fatal("guessed metric units")
		}
	}
	plain := q
	plain.GroupTimeSeries = false
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.GroupedTimeSeries, "GROUP BY g0, from_ms") || !strings.Contains(p.GroupedTimeSeries, "LIMIT 1001") {
		t.Fatal(p, e)
	}
	old, _ := Compile(plain)
	p.GroupedTimeSeries = ""
	if !reflect.DeepEqual(p, old) {
		t.Fatal("split chart changed existing SQL plans")
	}
	wire, _ := json.Marshal(plain)
	if strings.Contains(string(wire), "groupTimeSeries") {
		t.Fatal("default serialized new field")
	}
	draft, e := DraftTelemetryPackage(q, "routes", "Routes", 1)
	if e != nil || !draft.Question(q.FromMS, q.ToMS).GroupTimeSeries {
		t.Fatal("package lost split setting", e)
	}
	scoped, _ := (CaptureQuestion{Kind: "explore", Explore: &q}).Scope()
	other, _ := (CaptureQuestion{Kind: "explore", Explore: &plain}).Scope()
	if !reflect.DeepEqual(scoped, other) {
		t.Fatal("chart changes population")
	}
}
func TestSplitSeriesRejectsPartialInvalidAndOversizedCells(t *testing.T) {
	q, raw := seriesFixture()
	tests := [][]json.RawMessage{append(append([]json.RawMessage{}, raw...), raw[0]), {json.RawMessage(`{"from_ms":"1000","count":"1"}`)}, {json.RawMessage(`{"g0":3,"from_ms":"1000","count":"1"}`)}, {json.RawMessage(`{"g0":"a","from_ms":"1001","count":"1"}`)}}
	for _, input := range tests {
		if _, e := DecodeGroupedTimeSeries(q, input); e == nil {
			t.Fatal("invalid split series accepted", input)
		}
	}
	q.Summary = ""
	rows := []json.RawMessage{}
	for i := 0; i < 9; i++ {
		b, _ := json.Marshal(map[string]string{"g0": string(rune('a' + i)), "from_ms": "1000", "count": "1"})
		rows = append(rows, b)
	}
	if _, e := DecodeGroupedTimeSeries(q, rows); e == nil {
		t.Fatal("silently trimmed series")
	}
	if _, e := DecodeGroupedTimeSeries(q, make([]json.RawMessage, 1001)); e == nil {
		t.Fatal("silently trimmed cells")
	}
	q.GroupBy = nil
	if e := q.Validate(); e == nil {
		t.Fatal("split chart without breakdown")
	}
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	q.BucketMS = 0
	if e := q.Validate(); e == nil {
		t.Fatal("split chart without interval")
	}
}
func TestSplitSeriesCrossReadDisagreement(t *testing.T) {
	q, raw := seriesFixture()
	groups, e := DecodeGroupedTimeSeries(q, raw)
	if e != nil {
		t.Fatal(e)
	}
	totalRaw := []json.RawMessage{json.RawMessage(`{"from_ms":"1000","count":"3","sum_ns":"30","min_ns":"0","max_ns":"20","quantiles_ns":["10","20","20"]}`), raw[2], raw[1]}
	points, e := DecodeTimeSeries(q, totalRaw)
	if e != nil {
		t.Fatal(e)
	}
	r := Result{Question: q, TimeSeries: points, GroupedTimeSeries: groups, Groups: []json.RawMessage{
		json.RawMessage(`{"g0":"/a","count":"3","duration":{"state":"observed","count":"3","sumNs":"70","minNs":"10","maxNs":"40","p50Ns":"20","p95Ns":"40","p99Ns":"40"}}`),
		json.RawMessage(`{"g0":null,"count":"1","duration":{"state":"observed","count":"1","sumNs":"50","minNs":"50","maxNs":"50","p50Ns":"50","p95Ns":"50","p99Ns":"50"}}`),
		json.RawMessage(`{"g0":"","count":"1","duration":{"state":"observed","count":"1","sumNs":"0","minNs":"0","maxNs":"0","p50Ns":"0","p95Ns":"0","p99Ns":"0"}}`)}}
	if !groupedSeriesAgree(&r) {
		t.Fatal("matching additive states disagree")
	}
	for i := range r.GroupedTimeSeries {
		g := &r.GroupedTimeSeries[i]
		if groupKey(g.Values) == `["/a"]` {
			g.Points[0].Duration.SumNS = "31"
		}
	}
	if groupedSeriesAgree(&r) {
		t.Fatal("same-count changed durations hidden")
	}
	r.GroupedTimeSeries[0].Points[0].ToMS++
	if ValidateGroupedTimeSeries(q, r.GroupedTimeSeries) == nil {
		t.Fatal("invalid boundary accepted")
	}
}
func TestSplitLogCountsDoNotAcquireSpanOrMetricSemantics(t *testing.T) {
	q, raw := seriesFixture()
	q.Signal = "logs"
	q.Summary = ""
	measures, e := q.RecordMeasures()
	if e != nil || len(measures.Measures) != 1 || measures.Scope.Grain != "log-record" || measures.Measures[0].Unit != "{record}" {
		t.Fatal(measures, e)
	}
	p, e := Compile(q)
	if e != nil || strings.Contains(p.GroupedTimeSeries, "Duration") || !strings.Contains(p.GroupedTimeSeries, "LogAttributes") {
		t.Fatal("wrong signal lowering", p, e)
	}
	groups, e := DecodeGroupedTimeSeries(q, raw)
	if e != nil {
		t.Fatal(e)
	}
	for _, g := range groups {
		for _, point := range g.Points {
			if point.Duration != nil || point.SpanStatus != nil {
				t.Fatal("invented log measures")
			}
		}
	}
}
func TestRecordMeasureLoweringRejectsUnsupportedBindings(t *testing.T) {
	q, _ := seriesFixture()
	measures := recordMeasures(q)
	measures[1].Unit = "s"
	if _, e := lowerRecordMeasures(measures); e == nil {
		t.Fatal("implicit unit conversion")
	}
	measures = recordMeasures(q)
	measures[1].Input = &Field{Kind: "attribute", Key: "custom.number"}
	if _, e := lowerRecordMeasures(measures); e == nil {
		t.Fatal("implicit string cast")
	}
	measures = recordMeasures(q)
	measures[5].Probability = "0.90"
	if _, e := lowerRecordMeasures(measures); e == nil {
		t.Fatal("unsupported percentile profile")
	}
	if _, e := lowerRecordMeasures([]RecordMeasure{{Operator: "arbitrary-sql"}}); e == nil {
		t.Fatal("unsupported operator")
	}
}
