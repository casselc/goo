package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type histogramTimePoint struct {
	Key, FromNS, ToNS, State string
	Value                    *string
}
type histogramTimeView struct{ Points []histogramTimePoint }

func TestQuintHistogramTimeLive(t *testing.T) {
	file := os.Getenv("GOO_HISTOGRAM_TIME_LEDGER")
	if file == "" {
		t.Skip("real OTLP ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Result, Crossing              measurements.Result
			View, RawView, SavedView      histogramTimeView
			Before, After, RecipientReads int
			Capture, Reopened             exploration.CapturePage
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Result.HistogramTime.Series[0].Buckets[0].Distribution.Count = "8"
		case 1:
			l.Result.HistogramTime.Series[0].Buckets[1].Distribution = &metrics.Distribution{Count: "0"}
		case 2:
			l.Result.HistogramTime.WidthNS = "3000000000"
		case 3:
			v := "0"
			l.View.Points[1].Value = &v
		case 4:
			l.After++
		case 5:
			v := int64(3000)
			l.Capture.Receipt.Question.Measurement.Definition.HistogramBucketMS = &v
		case 6:
			l.RecipientReads++
		case 7:
			l.Reopened.Measurement.HistogramTime.SourceBasis = "corrupt"
		case 8:
			l.SavedView.Points[0].FromNS = "0"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, coherent bool) {
			rec.Record(context.Background(), action, nil, "obs.coherent", coherent)
		}
		r := &l.Result
		evaluated := measurements.ValidateResult(r) == nil && r.HistogramTime != nil && r.HistogramTime.SourceBasis == r.Histograms.Basis && r.Question.Definition.HistogramBucketMS != nil && *r.Question.Definition.HistogramBucketMS == 4000
		record("evaluate", evaluated)
		inspected := l.Before == l.After && len(l.RawView.Points) == 9 && len(l.View.Points) == 5
		if len(r.HistogramTime.Series) != 1 {
			return false
		}
		buckets := r.HistogramTime.Series[0].Buckets
		if len(buckets) != len(l.View.Points) {
			return false
		}
		for i, b := range buckets {
			p := l.View.Points[i]
			inspected = inspected && p.FromNS == b.FromNS && p.ToNS == b.ToNS
			if b.State == "unavailable" {
				inspected = inspected && p.State == "unknown" && p.Value == nil && b.Distribution == nil
			} else {
				inspected = inspected && p.State == "observed" && p.Value != nil && b.Distribution != nil
				if p.Value != nil && b.Distribution != nil {
					inspected = inspected && *p.Value == b.Distribution.Count
				}
			}
		}
		inspected = inspected && buckets[0].Distribution.Count == "7" && buckets[2].State == "no_data" && buckets[2].Distribution.Count == "0" && buckets[3].State == "unavailable"
		record("inspect", inspected)
		cross := l.Crossing.HistogramTime.Series[0].Buckets[0]
		crossing := measurements.ValidateResult(&l.Crossing) == nil && l.Crossing.HistogramTime.Basis != r.HistogramTime.Basis && l.Crossing.HistogramTime.WidthNS == "3000000000" && cross.State == "unavailable" && cross.Distribution == nil
		hasReason := false
		for _, reason := range cross.Reasons {
			hasReason = hasReason || reason == "boundary-crossing"
		}
		record("crossing", crossing && hasReason)
		record("capture", reflect.DeepEqual(l.Capture.Receipt.Question.Measurement, &r.Question) && reflect.DeepEqual(l.Capture.Measurement.HistogramTime, r.HistogramTime) && measurements.ValidateResult(l.Capture.Measurement) == nil)
		record("reopen", l.RecipientReads == 0 && reflect.DeepEqual(l.Capture, l.Reopened) && reflect.DeepEqual(l.View, l.SavedView))
		checker, e := validate.NewChecker("model/histogram_time.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("actual histogram time trace rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("corrupt histogram time trace accepted", i)
		}
	}
}
