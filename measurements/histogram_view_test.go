package measurements

import (
	"encoding/json"
	"github.com/casselc/goo/scopedview"
	"math"
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func histogramViewFixture(t *testing.T) *Result {
	t.Helper()
	d := definition(t, "checkout-duration")
	q := Question{Definition: d, FromMS: 100000, ToMS: 104000}
	row := gaugeRow(d, "one", 101000, "")
	row.StartNS = ns(100000)
	row.Temporality = 1
	row.Count = "3"
	row.Bounds = []float64{0, 0.1, 1}
	row.BucketCounts = []string{"0", "1", "0", "2"}
	r, err := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestHistogramViewExactLazyProjection(t *testing.T) {
	r := histogramViewFixture(t)
	before, _ := json.Marshal(r)
	d, err := HistogramView(r, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Histograms.Series[0].State != "unknown" {
		t.Fatal("fixture must contain unrelated gap")
	}
	if d.Scope.Signal != "metrics" || d.Histogram.Total != "3" || d.Histogram.Grain != "observations" || d.Histogram.Closure != "left-open-right-closed" || d.SourceBasis != r.Histograms.Basis || len(d.Rows) != 4 {
		t.Fatalf("bad native metadata: %+v", d)
	}
	if *d.Rows[0].Values["from"] != "-inf" || *d.Rows[3].Values["to"] != "+inf" || *d.Rows[1].Values["to"] != new(big.Rat).SetFloat64(.1).RatString() {
		t.Fatal("lost exact binary64 bounds or open tails")
	}
	if d.Rows[0].State != "no_data" || d.Completeness == nil || d.Completeness.ScopeID != d.Scope.ID {
		t.Fatal("lost knownzero/fullinterval coverage")
	}
	wire, _ := json.Marshal(d)
	if strings.Contains(string(wire), `"support"`) || strings.Contains(string(wire), `"binding"`) {
		t.Fatal("invented numeric parse metadata")
	}
	var sel HistogramViewSelector
	if json.Unmarshal(d.Scope.Selector, &sel) != nil || !reflect.DeepEqual(sel.Question, r.Question) || !sameExactStream(sel.Stream, r.Histograms.Series[0].Stream) {
		t.Fatal("lost exact measurement scope")
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("projection mutated measurement")
	}
	r.Count = "999"
	if _, err = HistogramView(r, 0, 0); err != nil {
		t.Fatal("raw export count incorrectly used as observation total", err)
	}
}
func TestHistogramViewRefusesMalformedOrUnknown(t *testing.T) {
	for name, mutate := range map[string]func(*Result){
		"parent-basis": func(r *Result) { r.Basis = "bad" },
		"readclock":    func(r *Result) { r.ReadAt = "bad" },
		"policy":       func(r *Result) { r.Histograms.MaxGapNS = "1" },
		"exactbinding": func(r *Result) {
			stream := r.Histograms.Series[0].Stream
			stream.Resource = map[string]string{}
			for k, v := range r.Histograms.Series[0].Stream.Resource {
				stream.Resource[k] = v
			}
			stream.Resource["service.instance.id"] = "other"
			r.Question.Definition.Source.ExactStream = &stream
		},
		"wrongwindow":  func(r *Result) { r.Histograms.ToNS = "104000000001" },
		"unknown":      func(r *Result) { r.Histograms.Series[0].Intervals[0].State = "unknown" },
		"reason":       func(r *Result) { r.Histograms.Series[0].Intervals[0].Reason = "reset" },
		"distribution": func(r *Result) { r.Histograms.Series[0].Intervals[0].Distribution = nil },
		"binding":      func(r *Result) { r.Histograms.Series[0].Stream.Name = "other" },
		"temporality":  func(r *Result) { r.Histograms.Series[0].Temporality = "unknown" },
		"unit":         func(r *Result) { r.Histograms.Series[0].Intervals[0].Distribution.Unit = "ms" },
		"sum":          func(r *Result) { r.Histograms.Series[0].Intervals[0].Distribution.Count = "4" },
		"overflow": func(r *Result) {
			r.Histograms.Series[0].Intervals[0].Distribution.BucketCounts = []string{"18446744073709551615", "1", "0", "0"}
		},
		"nan":          func(r *Result) { r.Histograms.Series[0].Intervals[0].Distribution.Bounds[1] = math.NaN() },
		"negative":     func(r *Result) { r.Histograms.Series[0].Intervals[0].Distribution.Bounds[0] = -1 },
		"noncanonical": func(r *Result) { r.Histograms.Series[0].Intervals[0].Distribution.BucketCounts[0] = "00" },
	} {
		t.Run(name, func(t *testing.T) {
			r := histogramViewFixture(t)
			mutate(r)
			if _, e := HistogramView(r, 0, 0); e == nil {
				t.Fatal("accepted invalid projection")
			}
		})
	}
	r := histogramViewFixture(t)
	for _, pair := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 100}} {
		if _, e := HistogramView(r, pair[0], pair[1]); e == nil {
			t.Fatal("accepted foreign membership", pair)
		}
	}
}
func TestHistogramViewZeroAndExtremeBounds(t *testing.T) {
	r := histogramViewFixture(t)
	p := r.Histograms.Series[0].Intervals[0].Distribution
	p.Bounds = []float64{0, math.SmallestNonzeroFloat64, math.MaxFloat64}
	p.BucketCounts = []string{"0", "0", "0", "0"}
	p.Count = "0"
	d, e := HistogramView(r, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	for _, row := range d.Rows {
		if row.State != "no_data" {
			t.Fatal("knownempty observationstate")
		}
	}
	if *d.Rows[1].Values["to"] != new(big.Rat).SetFloat64(math.SmallestNonzeroFloat64).RatString() || *d.Rows[2].Values["to"] != new(big.Rat).SetFloat64(math.MaxFloat64).RatString() {
		t.Fatal("rounded extreme finite bounds")
	}
}

func TestHistogramViewSchemaFixtures(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_HISTOGRAM_VIEWS")
	if path == "" {
		return
	}
	r := histogramViewFixture(t)
	observed, err := HistogramView(r, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	p := r.Histograms.Series[0].Intervals[0].Distribution
	p.Count = "0"
	p.BucketCounts = []string{"0", "0", "0", "0"}
	empty, err := HistogramView(r, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(map[string]*scopedview.Dataset{"observed": observed, "empty": empty})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatal(err)
	}
}
