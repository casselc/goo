package measurements

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
)

// Compare the new source admission with the unchanged pure kernels over the same
// raw streams. Matching effective labels never become stream identity.
func TestTypedCounterHistogramKernelParity320(t *testing.T) {
	for _, kind := range []string{"counter", "histogram"} {
		for _, selector := range selectors319() {
			t.Run(kind+selector.Policy+selector.Scope.Mode, func(t *testing.T) {
				d := gauge319(selector)
				d.Kind = kind
				d.Source.Metric = "typed-" + kind
				if kind == "histogram" {
					d.Source.Unit = "ms"
					width := int64(2000)
					d.HistogramBucketMS = &width
				}
				q := Question{Definition: d, FromMS: 100000, ToMS: 104000}
				plan, err := Compile(q)
				if err != nil {
					t.Fatal(err)
				}
				if plan.Parameters["p0"] != "40000" || plan.Parameters["p1"] != "104000" || !strings.Contains(plan.SQL, "TimeUnix<=fromUnixTimestamp64Milli") {
					t.Fatal("predecessor lookback/end-inclusive source bounds changed")
				}
				selected, _ := d.EnvironmentSelection()
				var rawRows []json.RawMessage
				var counters []metrics.CounterPoint
				var histograms []metrics.HistogramPoint
				streams := 0
				for _, base := range rows319(d) {
					if !selected.Matches(base.Resource) {
						continue
					}
					streams++
					for index, end := range []int64{100000, 102000, 104000} {
						row := base
						row.StartNS = ns(50000)
						row.EndNS = ns(end)
						row.Temporality = 2
						row.Monotonic = true
						row.Attributes = map[string]string{}
						stream := metrics.Stream{Resource: row.Resource, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}
						if kind == "counter" {
							row.Value = []string{"10", "18", "3"}[index]
							counters = append(counters, metrics.CounterPoint{Stream: stream, Temporality: "cumulative", Monotonic: true, StartNS: row.StartNS, EndNS: row.EndNS, Value: row.Value})
						} else {
							row.Count = []string{"2", "5", "1"}[index]
							row.Bounds = []float64{10, 20}
							row.BucketCounts = [][]string{{"1", "1", "0"}, {"2", "2", "1"}, {"1", "0", "0"}}[index]
							histograms = append(histograms, metrics.HistogramPoint{Stream: stream, Temporality: "cumulative", StartNS: row.StartNS, EndNS: row.EndNS, Bounds: row.Bounds, BucketCounts: row.BucketCounts, Count: row.Count})
						}
						rawRows = append(rawRows, raw(row))
					}
				}
				actual, err := Evaluate(q, rawRows, time.UnixMilli(105000))
				if err != nil {
					t.Fatal(err)
				}
				if err := ValidateResult(actual); err != nil {
					t.Fatal(err)
				}
				if kind == "counter" {
					expected, err := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: ns(q.FromMS), ToNS: ns(q.ToMS), MaxGapNS: ns(d.MaxGapMS), Points: counters})
					if err != nil || !reflect.DeepEqual(actual.Counters, expected) || len(actual.Counters.Series) != streams {
						t.Fatal("typed source altered counter kernel/stream identity", err)
					}
					for _, series := range actual.Counters.Series {
						if len(series.Intervals) != 2 || series.Intervals[0].Delta != "8" || series.Intervals[1].Reason != "counter-decrease" || series.Intervals[1].Delta != "" {
							t.Fatal("counter decrease/rate support was invented")
						}
					}
				} else {
					expected, err := metrics.EvaluateHistograms(metrics.HistogramRequest{FromNS: ns(q.FromMS), ToNS: ns(q.ToMS), MaxGapNS: ns(d.MaxGapMS), Points: histograms})
					if err != nil || !reflect.DeepEqual(actual.Histograms, expected) || len(actual.Histograms.Series) != streams {
						t.Fatal("typed source altered histogram kernel/stream identity", err)
					}
					bucketed, err := metrics.BucketHistograms(expected, ns(*d.HistogramBucketMS))
					if err != nil || !reflect.DeepEqual(actual.HistogramTime, bucketed) {
						t.Fatal("typed source altered whole-export bucket support")
					}
					for _, series := range actual.Histograms.Series {
						if len(series.Intervals) != 2 || series.Intervals[0].Distribution.Count != "3" || series.Intervals[1].Reason != "histogram-decrease" || series.Intervals[1].Distribution != nil {
							t.Fatal("histogram reset distribution was invented")
						}
					}
				}
				bad := rawRows[1]
				var fields map[string]json.RawMessage
				json.Unmarshal(bad, &fields)
				fields["temporality"] = json.RawMessage("3")
				invalid, _ := json.Marshal(fields)
				if _, err := Evaluate(q, []json.RawMessage{rawRows[0], invalid}, time.UnixMilli(105000)); err == nil {
					t.Fatal("unsupported temporality admitted")
				}
			})
		}
	}
}
func TestTypedCounterDomainAndHistogramUnitAdmission320(t *testing.T) {
	d := gauge319(selectors319()[1])
	d.Kind = "counter"
	d.CounterDomain = metrics.DecimalCounterDomain
	if d.Validate() != nil {
		t.Fatal("existing optional decimal counter policy lost")
	}
	d.CounterDomain = "invented"
	if d.Validate() == nil {
		t.Fatal("unknown counter domain admitted")
	}
	d = gauge319(selectors319()[1])
	d.Kind = "histogram"
	if _, err := Compile(Question{Definition: d, FromMS: 100000, ToMS: 104000}); err == nil {
		t.Fatal("typed non-duration histogram reached SQL")
	}
	d.Source.Unit = "ms"
	d.Grain = "consumer-group"
	if d.Validate() == nil {
		t.Fatal("typed consumer-group histogram admitted")
	}
}
