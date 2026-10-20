package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

func TestQuintFreshComparisonLive(t *testing.T) {
	file := os.Getenv("GOO_FRESH_COMPARISON_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action string
				Reads  int
			}
			Cases []struct {
				Original, Saved   workbench.Document
				Baseline, Capture exploration.CapturePage
				Receipt           workbench.ComparisonReceipt
			}
			After int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[0].Reads++
		case 1:
			l.Cases[0].Saved.Cards[0].Step.Scope.FromMS++
		case 2:
			l.Cases[0].Receipt.Result.Baseline.Receipt.ID = l.Cases[0].Capture.Receipt.ID
		case 3:
			l.Cases[0].Capture.Receipt.Question.Explore.Environment = "wrong"
		case 4:
			l.Cases[1].Capture.Receipt.Question.Measurement.Definition.Source.Unit = "wrong"
		case 5:
			l.After++
		}
		coherent := l.After == 2 && len(l.Cases) == 2
		for _, c := range l.Cases {
			a, b := c.Baseline.Receipt.Question, c.Capture.Receipt.Question
			coherent = coherent && reflect.DeepEqual(c.Saved.Cards[0], c.Original.Cards[0]) && c.Receipt.Result.Baseline.Receipt.ID == c.Baseline.Receipt.ID && c.Receipt.Result.Selection.Receipt.ID == c.Capture.Receipt.ID && c.Receipt.Result.State == "comparable" && c.Saved.Cards[1].Reference == c.Receipt.ID
			if a.Explore != nil {
				x, y := *a.Explore, *b.Explore
				coherent = coherent && (x.ToMS-x.FromMS) == (y.ToMS-y.FromMS) && y.FromMS != x.FromMS
				x.FromMS, x.ToMS = 0, 0
				y.FromMS, y.ToMS = 0, 0
				coherent = coherent && reflect.DeepEqual(x, y)
			} else {
				coherent = coherent && reflect.DeepEqual(a.Measurement.Definition, b.Measurement.Definition) && (a.Measurement.ToMS-a.Measurement.FromMS) == (b.Measurement.ToMS-b.Measurement.FromMS)
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/fresh_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
