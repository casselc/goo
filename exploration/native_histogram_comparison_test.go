package exploration

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func nativeHistogramPairFixture(t *testing.T) (*measurements.Result, *measurements.Result, metrics.WindowComparison) {
	t.Helper()
	raw, err := os.ReadFile("../config/measurements/checkout-duration.yaml")
	if err != nil {
		t.Fatal(err)
	}
	d, err := measurements.Read(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	d.MaxGapMS = 3000
	makeResult := func(from int64, n int, count string) *measurements.Result {
		q := measurements.Question{Definition: d, FromMS: from, ToMS: from + int64(n)*2000}
		rows := []json.RawMessage{}
		for i := 0; i < n; i++ {
			start := q.FromMS + int64(i)*2000
			row, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": d.Source.Services[0], "service.instance.id": "one", "deployment.environment.name": d.Environment}, "scope": d.Source.Scope, "version": d.Source.Version, "schema_url": d.Source.SchemaURL, "attributes": d.Source.Attributes, "name": d.Source.Metric, "unit": d.Source.Unit, "start_ns": strconv.FormatInt(start*1000000, 10), "end_ns": strconv.FormatInt((start+2000)*1000000, 10), "temporality": 1, "count": count, "bounds": []float64{.1, .5}, "bucket_counts": []string{"1", "1", strconv.Itoa(mustInt(t, count) - 2)}})
			rows = append(rows, row)
		}
		r, e := measurements.Evaluate(q, rows, time.Unix(from/1000+200, 123))
		if e != nil {
			t.Fatal(e)
		}
		return r
	}
	a, b := makeResult(100000, 1, "3"), makeResult(200000, 2, "4")
	w := metrics.CompareHistogramWindows(a.Histograms, b.Histograms)
	if len(w) != 1 || w[0].State != "comparable" {
		t.Fatal(w)
	}
	return a, b, w[0]
}
func mustInt(t *testing.T, s string) int {
	t.Helper()
	n, e := strconv.Atoi(s)
	if e != nil {
		t.Fatal(e)
	}
	return n
}
func TestNativeHistogramComparisonIndependentHeldOperands(t *testing.T) {
	a, b, w := nativeHistogramPairFixture(t)
	before, _ := json.Marshal([]*measurements.Result{a, b})
	p, err := NativeHistogramComparison(a, b, w)
	if err != nil {
		t.Fatal(err)
	}
	if p.State != "observed" || p.Baseline.Histogram.Total != "3" || p.Selection.Histogram.Total != "8" || p.Baseline.CapturedAt == p.Selection.CapturedAt || p.Baseline.Scope.ID == p.Selection.Scope.ID {
		t.Fatal("lost independent populations/clocks")
	}
	for _, side := range []*ViewDataset{p.Baseline, p.Selection} {
		if side.Origin != "browser-held" || side.Completeness != nil || side.Quality.Extent != "unknown" {
			t.Fatal("pair gained source proof")
		}
	}
	after, _ := json.Marshal([]*measurements.Result{a, b})
	if string(before) != string(after) {
		t.Fatal("mutated retained operands")
	}
	if path := os.Getenv("GOO_NATIVE_HISTOGRAM_PAIR_FIXTURE"); path != "" {
		encoded, _ := json.Marshal(p)
		if err = os.WriteFile(path, encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
func TestNativeHistogramComparisonPreservesWithholding(t *testing.T) {
	changes := []func(*measurements.Result, *measurements.Result, *metrics.WindowComparison){
		func(a, b *measurements.Result, w *metrics.WindowComparison) {
			w.State = "unavailable"
			w.Reasons = []string{"another-member-withholds-comparison"}
		},
		func(a, b *measurements.Result, w *metrics.WindowComparison) { w.Reasons = []string{"gap"} },
		func(a, b *measurements.Result, w *metrics.WindowComparison) { w.Reasons = nil },
		func(a, b *measurements.Result, w *metrics.WindowComparison) { b.Question.Definition.MaxGapMS++ },
		func(a, b *measurements.Result, w *metrics.WindowComparison) {
			w.Selection.State = "no_data"
			w.Selection.Reasons = []string{"empty-distribution"}
			w.Selection.Distribution.Count = "0"
			w.Selection.Distribution.BucketCounts = []string{"0", "0", "0"}
		},
		func(a, b *measurements.Result, w *metrics.WindowComparison) {
			w.Selection.Distribution.Bounds = []float64{.2, .8}
		},
		func(a, b *measurements.Result, w *metrics.WindowComparison) {
			b.Histograms.Series[0].Temporality = "cumulative"
			w.Selection.Temporality = "cumulative"
		},
	}
	for i, change := range changes {
		a, b, w := nativeHistogramPairFixture(t)
		change(a, b, &w)
		if p, err := NativeHistogramComparison(a, b, w); p != nil || err == nil {
			t.Fatalf("invalid pair %d admitted", i)
		}
	}
}

func TestNativeHistogramComparisonEvidenceSourceMetadata(t *testing.T) {
	a, b, _ := nativeHistogramPairFixture(t)
	capture := func(r *measurements.Result, id string) *SavedCapture {
		return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, ID: id, Title: "Native histogram comparison", CapturedAt: "2026-10-31T00:00:00Z", Question: CaptureQuestion{Kind: "measurement", Measurement: &r.Question}}, Measurement: r}
	}
	before, _ := json.Marshal([]*measurements.Result{a, b})
	evidence, err := CompareEvidence(capture(a, strings.Repeat("a", 64)), capture(b, strings.Repeat("b", 64)))
	if err != nil {
		t.Fatal(err)
	}
	if evidence.State != "comparable" || evidence.Baseline.NativeHistogramSource == nil || evidence.Selection.NativeHistogramSource == nil || len(evidence.NativeWindows) != 1 {
		t.Fatal("missing native source context")
	}
	source := evidence.Baseline.NativeHistogramSource
	if source.MeasurementBasis != a.Basis || source.HistogramBasis != a.Histograms.Basis || len(source.Streams) != 1 {
		t.Fatal("lost captured source identity")
	}
	encoded, _ := json.Marshal(source)
	for _, forbidden := range []string{`"rows"`, `"intervals"`, `"buckets"`, `"distribution"`} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatal("source context expanded observations")
		}
	}
	after, _ := json.Marshal([]*measurements.Result{a, b})
	if string(before) != string(after) {
		t.Fatal("metadata projection changed source")
	}
	if path := os.Getenv("GOO_NATIVE_HISTOGRAM_EVIDENCE_FIXTURE"); path != "" {
		encoded, _ := json.Marshal(evidence)
		if err = os.WriteFile(path, encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
