package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintServiceNativeLedger(t *testing.T) {
	name := os.Getenv("GOO_SERVICE_NATIVE_LEDGER")
	if name == "" {
		t.Skip("real-source opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type row struct {
		NativeKind, Instance, Value, State, SeriesState, FromNs, ToNs string
		Question                                                      workbench.Step
	}
	check := func(mutation int) bool {
		var l struct {
			Plan                exploration.DashboardPlan
			Native              []measurements.Result
			Rows, GaugeRows     []row
			Before, AfterBrowse int
			Focused             workbench.Step
			Preview             struct{ Question measurements.Question }
			Selected            measurements.Result
			Capture, Reopened   exploration.CapturePage
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Plan.Scope.Environment = "other"
		case 1:
			l.Rows[0].Value = "5 {request}/s"
		case 2:
			for i := range l.Rows {
				if l.Rows[i].Instance == "worker-01" {
					l.Rows[i].Value = "10 {request}/s"
				}
			}
		case 3:
			l.AfterBrowse++
		case 4:
			l.Preview.Question.Definition.Source.ExactStream.Resource["service.namespace"] = "other"
		case 5:
			l.Selected.Counters.Series[0].Stream.Resource["service.instance.id"] = "other"
		case 6:
			l.Capture.Measurement.Question.Definition.Source.ExactStream = nil
		case 7:
			l.Reopened.Receipt.ID = "other"
		case 8:
			l.GaugeRows[0].Value = "0"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		nativeWidgets := 0
		scopeOK := true
		for _, w := range l.Plan.Widgets {
			if w.Measurement != nil {
				nativeWidgets++
				q := w.Measurement
				scopeOK = scopeOK && q.Definition.Environment == l.Plan.Scope.Environment && q.Definition.Source.Namespace.Equal(l.Plan.Scope.Namespace) && q.FromMS == l.Plan.Scope.FromMS && q.ToMS == l.Plan.Scope.ToMS
			}
		}
		record("automatic", nativeWidgets == 2 && scopeOK)
		bad, rate, hist := false, false, false
		for _, r := range l.Rows {
			if r.Instance == "worker-00" {
				bad = r.Value == "Unavailable" && r.State == "unknown"
			}
			if r.Instance == "worker-01" {
				rate = r.Value == "5 {request}/s" && r.SeriesState == "unknown"
			}
			if r.NativeKind == "histogram" {
				hist = r.Value == "4 observations · P95 bucket range: (0.5, +∞) s · no interpolation"
			}
		}
		record("values", bad && rate && hist)
		record("browse", l.Before == l.AfterBrowse)
		q := l.Preview.Question
		selector := q.Definition.Source.ExactStream
		record("preview", selector != nil && selector.Resource["service.namespace"] == "payments" && selector.Resource["service.instance.id"] == "worker-01" && q.FromMS == l.Plan.Scope.FromMS && q.ToMS == l.Plan.Scope.ToMS)
		record("evaluate", len(l.Selected.Counters.Series) == 1 && l.Selected.Counters.Series[0].Stream.Resource["service.instance.id"] == "worker-01" && reflect.DeepEqual(l.Selected.Question, q))
		record("capture", reflect.DeepEqual(l.Capture.Receipt.Question.Measurement, &q) && reflect.DeepEqual(l.Capture.Measurement.Question, q))
		record("reopen", reflect.DeepEqual(l.Capture, l.Reopened))
		record("gauge", len(l.GaugeRows) == 1 && l.GaugeRows[0].Value == "0.25 1" && l.GaugeRows[0].State == "observed")
		checker, e := validate.NewChecker("model/service_native.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual native-service ledger rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("corrupt native service trace accepted", i)
		}
	}
}
