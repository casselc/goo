package measurements

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/casselc/goo/metrics"
)

func TestHistogramWindowViewHeldWholeWindow(t *testing.T) {
	r := histogramTimeViewFixture(t)
	w := metrics.CompareHistogramWindows(r.Histograms, r.Histograms)[0]
	before, _ := json.Marshal(r)
	d, err := HistogramWindowView(r, w.Stream, w.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	if d.Origin != "browser-held" || d.Quality.Extent != "unknown" || d.Completeness != nil || d.Histogram.Total != "9" || d.SourceBasis != r.Histograms.Basis || d.Scope.ID != r.Basis+":native-window:0" || d.CapturedAt != r.ReadAt {
		t.Fatalf("lost independent held window metadata: %+v", d)
	}
	var selector HistogramWindowViewSelector
	if json.Unmarshal(d.Scope.Selector, &selector) != nil || selector.Kind != "measurement-window" || selector.FromNS != "100000000000" || selector.ToNS != "110000000000" {
		t.Fatal("invalid full-window scope")
	}
	if *d.Rows[0].Values["from"] != "-inf" || *d.Rows[len(d.Rows)-1].Values["to"] != "+inf" {
		t.Fatal("lost infinite geometry")
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("projection mutated retained measurement")
	}
	// The whole window is ten seconds, while each export has a three-second gap
	// ceiling. Full-window projection must not apply the per-export duration cap.
}
func TestHistogramWindowViewRefusesInvalidSummaryAndMembership(t *testing.T) {
	changes := []func(*Result, *metrics.WindowSummary, *metrics.Stream){
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { s.FromNS = "100000000001" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { s.State = "unavailable" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { s.Reasons = nil },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { s.Reasons = []string{"gap"} },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { s.Temporality = "cumulative" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { s.Distribution.Count = "10" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { r.Histograms.Series[0].State = "unknown" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) {
			r.Histograms.Series = append(r.Histograms.Series, r.Histograms.Series[0])
		},
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { r.Histograms.MaxGapNS = "1" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { r.ReadAt = "2026-02-31T00:00:00Z" },
		func(r *Result, s *metrics.WindowSummary, _ *metrics.Stream) { r.Basis = "bad" },
		func(r *Result, s *metrics.WindowSummary, stream *metrics.Stream) { stream.Name = "foreign" },
	}
	for i, change := range changes {
		r := histogramTimeViewFixture(t)
		w := metrics.CompareHistogramWindows(r.Histograms, r.Histograms)[0]
		change(r, w.Baseline, &w.Stream)
		if d, err := HistogramWindowView(r, w.Stream, w.Baseline); err == nil || d != nil {
			t.Fatalf("invalid case %d admitted", i)
		}
	}
}

func TestHistogramSourceIsCompactDetachedMetadata(t *testing.T) {
	r := histogramTimeViewFixture(t)
	before, _ := json.Marshal(r)
	source := HistogramSource(r)
	if source == nil || source.MeasurementBasis != r.Basis || source.HistogramBasis != r.Histograms.Basis || len(source.Streams) != 1 {
		t.Fatal("lost source context")
	}
	source.Streams[0].Stream.Resource["service.name"] = "mutated"
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("source metadata aliased retained resource attributes")
	}
	source = HistogramSource(r)
	wire, _ := json.Marshal(source)
	for _, forbidden := range []string{`"rows"`, `"intervals"`, `"buckets"`, `"distribution"`, `"bucketCounts"`} {
		if bytes.Contains(wire, []byte(forbidden)) {
			t.Fatalf("expanded source observations: %s", forbidden)
		}
	}
	r.Histograms.FromNS = "1"
	if HistogramSource(r) != nil {
		t.Fatal("mismatched source policy admitted")
	}
}
