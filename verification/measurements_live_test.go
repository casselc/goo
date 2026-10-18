package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

type measurementLedger struct {
	Q, Reused                                       measurements.Question
	Results                                         []*measurements.Result
	Capture, Reopened, Refresh                      exploration.CapturePage
	Document                                        workbench.Document
	Before, AfterPreview, BeforeReopen, AfterReopen int
}

func TestQuintMeasurementsLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_MEASUREMENTS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l measurementLedger
	if e = json.Unmarshal(data, &l); e != nil {
		t.Fatal(e)
	}
	replay := 0
	check := func(l measurementLedger) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("measurements"))
		record := func(action string, good bool) { rec.Record(context.Background(), action, nil, "obs.coherent", good) }
		_, e := measurements.Compile(l.Q)
		record("preview", e == nil && l.Before == l.AfterPreview)
		good := len(l.Results) >= 4
		families := map[string]bool{}
		for _, r := range l.Results {
			good = good && measurements.ValidateResult(r) == nil && r.Count != "0"
			families[r.Question.Definition.Kind] = true
		}
		record("execute", good && len(families) == 4)
		record("capture", l.Capture.Measurement != nil && measurements.ValidateResult(l.Capture.Measurement) == nil && l.Capture.Receipt.Question.Kind == "measurement" && reflect.DeepEqual(l.Capture.Measurement.Question, l.Q))
		record("reopen", reflect.DeepEqual(l.Capture, l.Reopened) && l.BeforeReopen == l.AfterReopen)
		record("refresh", l.Refresh.Measurement != nil && l.Refresh.Receipt.Parent == l.Capture.Receipt.ID && reflect.DeepEqual(l.Refresh.Measurement.Question, l.Q) && measurements.ValidateResult(l.Refresh.Measurement) == nil)
		record("document", len(l.Document.Cards) == 1 && l.Document.Cards[0].Reference == l.Capture.Receipt.ID)
		reuse := l.Reused
		fresh := reuse.ToMS >= l.Q.ToMS && reuse.ToMS-reuse.FromMS == l.Q.ToMS-l.Q.FromMS
		reuse.FromMS = l.Q.FromMS
		reuse.ToMS = l.Q.ToMS
		record("reuse", fresh && reflect.DeepEqual(reuse, l.Q))
		checker, e := validate.NewChecker("model/measurements.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		artifact(fmt.Sprintf("measurements-replay-%d.txt", replay), []byte(report.Summary()))
		replay++
		return report.OK()
	}
	if !check(l) {
		t.Fatal("live measurement ledger rejected")
	}
	for i := 0; i < 7; i++ {
		var bad measurementLedger
		json.Unmarshal(data, &bad)
		switch i {
		case 0:
			bad.AfterPreview++
		case 1:
			bad.Results[0].Question.Definition.Environment = "other"
		case 2:
			bad.Capture.Measurement.Question.FromMS++
		case 3:
			bad.AfterReopen++
		case 4:
			bad.Refresh.Receipt.Parent = "wrong"
		case 5:
			bad.Document.Cards[0].Reference = "wrong"
		case 6:
			bad.Reused.Definition.Source.Services = []string{"wrong"}
		}
		if check(bad) {
			t.Fatal("negative accepted", i)
		}
	}
}
