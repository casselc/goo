package measurements

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/scopedview"
)

func histogramTimeViewFixture(t *testing.T) *Result {
	t.Helper()
	d := definition(t, "checkout-duration")
	width := int64(4000)
	d.HistogramBucketMS = &width
	d.MaxGapMS = 3000
	q := Question{Definition: d, FromMS: 100000, ToMS: 110000}
	counts := [][]string{{"1", "1", "1"}, {"1", "1", "2"}, {"0", "0", "0"}, {"0", "0", "0"}, {"1", "0", "1"}}
	totals := []string{"3", "4", "0", "0", "2"}
	rows := []json.RawMessage{}
	for i, c := range counts {
		row := gaugeRow(d, "one", int64(102000+i*2000), "")
		row.StartNS = ns(int64(100000 + i*2000))
		row.Temporality = 1
		row.Count = totals[i]
		row.Bounds = []float64{.1, .5}
		row.BucketCounts = c
		rows = append(rows, raw(row))
	}
	r, err := Evaluate(q, rows, time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestHistogramTimeViewExactWholeExports(t *testing.T) {
	r := histogramTimeViewFixture(t)
	before, _ := json.Marshal(r)
	d, err := HistogramTimeView(r, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if d.Histogram.Total != "7" || d.SourceBasis != r.HistogramTime.Basis || d.SourceBasis == r.Histograms.Basis || d.Scope.ID != r.Basis+":native-time:0:0" {
		t.Fatalf("lost derived bucket basis/count: %+v", d)
	}
	var selector HistogramTimeViewSelector
	if err = json.Unmarshal(d.Scope.Selector, &selector); err != nil || selector.Kind != "measurement-time-bucket" || selector.WidthNS != "4000000000" || selector.BucketIndex != 0 || selector.FromNS != "100000000000" || selector.ToNS != "104000000000" || !reflect.DeepEqual(selector.Question, r.Question) {
		t.Fatalf("invalid bucket selector %+v %v", selector, err)
	}
	if *d.Rows[0].Values["from"] != "-inf" || *d.Rows[len(d.Rows)-1].Values["to"] != "+inf" || d.Completeness == nil {
		t.Fatal("lost exact native histogram geometry")
	}
	// The four-second bucket is complete, though each export is bounded by the
	// three-second gap policy. A raw-export duration cap cannot constrain it.
	if _, err = HistogramView(r, 0, 0); err != nil {
		t.Fatal(err)
	}
	empty, err := HistogramTimeView(r, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Histogram.Total != "0" || !reflect.DeepEqual(r.HistogramTime.Series[0].Buckets[1].Reasons, []string{"empty-distribution"}) {
		t.Fatal("known empty bucket incorrectly withheld")
	}
	for _, row := range empty.Rows {
		if row.State != "no_data" {
			t.Fatal("invented positive count")
		}
	}
	final, err := HistogramTimeView(r, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if json.Unmarshal(final.Scope.Selector, &selector) != nil || selector.FromNS != "108000000000" || selector.ToNS != "110000000000" || selector.WidthNS != "4000000000" || final.Histogram.Total != "2" {
		t.Fatal("short final bucket lost anchored width or exact boundaries")
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("projection mutated retained source")
	}
	r.Count = "999"
	if _, err = HistogramTimeView(r, 0, 0); err != nil {
		t.Fatal("used export count as observation denominator", err)
	}
}
func TestHistogramTimeViewRefusesMalformedOrUnknown(t *testing.T) {
	for name, mutate := range map[string]func(*Result){
		"parent-basis":      func(r *Result) { r.Basis = "bad" },
		"clock":             func(r *Result) { r.ReadAt = "2026-02-31T00:00:00Z" },
		"source-chain":      func(r *Result) { r.HistogramTime.SourceBasis = strings.Repeat("f", 64) },
		"derived-basis":     func(r *Result) { r.HistogramTime.Basis = "bad" },
		"window":            func(r *Result) { r.HistogramTime.ToNS = "110000000001" },
		"width":             func(r *Result) { r.HistogramTime.WidthNS = "4000000001" },
		"missing-width":     func(r *Result) { r.Question.Definition.HistogramBucketMS = nil },
		"source-policy":     func(r *Result) { r.Histograms.MaxGapNS = "1" },
		"stream":            func(r *Result) { r.HistogramTime.Series[0].Stream.Name = "other" },
		"temporality":       func(r *Result) { r.HistogramTime.Series[0].Temporality = "cumulative" },
		"duplicate-member":  func(r *Result) { r.Histograms.Series = append(r.Histograms.Series, r.Histograms.Series[0]) },
		"duplicate-derived": func(r *Result) { r.HistogramTime.Series = append(r.HistogramTime.Series, r.HistogramTime.Series[0]) },
		"missing-member":    func(r *Result) { r.Histograms.Series = []metrics.HistogramSeries{} },
		"missing-bucket":    func(r *Result) { r.HistogramTime.Series[0].Buckets = r.HistogramTime.Series[0].Buckets[:2] },
		"unanchored":        func(r *Result) { r.HistogramTime.Series[0].Buckets[0].FromNS = "100000000001" },
		"unknown":           func(r *Result) { r.HistogramTime.Series[0].Buckets[0].State = "unavailable" },
		"known-reason":      func(r *Result) { r.HistogramTime.Series[0].Buckets[0].Reasons = []string{"boundary-crossing"} },
		"null-reasons":      func(r *Result) { r.HistogramTime.Series[0].Buckets[0].Reasons = nil },
		"state-count":       func(r *Result) { r.HistogramTime.Series[0].Buckets[0].State = "no_data" },
		"distribution":      func(r *Result) { r.HistogramTime.Series[0].Buckets[0].Distribution = nil },
		"unit":              func(r *Result) { r.HistogramTime.Series[0].Buckets[0].Distribution.Unit = "ms" },
		"sum":               func(r *Result) { r.HistogramTime.Series[0].Buckets[0].Distribution.Count = "8" },
	} {
		t.Run(name, func(t *testing.T) {
			r := histogramTimeViewFixture(t)
			mutate(r)
			if _, err := HistogramTimeView(r, 0, 0); err == nil {
				t.Fatal("accepted malformed bucket projection")
			}
		})
	}
	for _, reasons := range [][]string{nil, {}, {"empty-distribution", "empty-distribution"}, {"empty-distribution", "missing-distribution"}} {
		r := histogramTimeViewFixture(t)
		r.HistogramTime.Series[0].Buckets[1].Reasons = reasons
		if _, err := HistogramTimeView(r, 0, 1); err == nil {
			t.Fatal("accepted malformed known-empty diagnostics", reasons)
		}
	}
	for _, pair := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 3}} {
		if _, err := HistogramTimeView(histogramTimeViewFixture(t), pair[0], pair[1]); err == nil {
			t.Fatal("accepted invalid bucket membership", pair)
		}
	}
}
func TestHistogramTimeViewSchemaFixtures(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_HISTOGRAM_TIME_VIEWS")
	if path == "" {
		return
	}
	r := histogramTimeViewFixture(t)
	views := map[string]*scopedview.Dataset{}
	for name, index := range map[string]int{"observed": 0, "empty": 1, "shortFinal": 2} {
		d, err := HistogramTimeView(r, 0, index)
		if err != nil {
			t.Fatal(err)
		}
		views[name] = d
	}
	r.Question.Definition.Grain = "consumer-group"
	d, err := HistogramTimeView(r, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	views["consumerGroup"] = d
	encoded, err := json.Marshal(views)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatal(err)
	}
}
