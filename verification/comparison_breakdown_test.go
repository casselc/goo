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

func TestQuintBreakdownLive(t *testing.T) {
	file := os.Getenv("GOO_BREAKDOWN_LEDGER")
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
			Original, Saved workbench.Document
			Receipt         workbench.ComparisonReceipt
			Captures        []exploration.CapturePage
			Focused         workbench.Step
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[0].Reads++
		case 1:
			l.Focused.Scope.FromMS++
		case 2:
			l.Receipt.Result.Breakdown.Rows[0].BaselineCount = "999"
		case 3:
			l.Saved.Cards[0].Evidence = nil
		case 5:
			l.Focused.Record.Where.All[1].Value = "wrong"
		case 4:
			l.Receipt.Result.Baseline.Receipt.ID = l.Captures[1].Receipt.ID
		}
		a, b := l.Captures[0], l.Captures[1]
		ac := exploration.SavedCapture{CaptureReceipt: a.Receipt, Explore: a.Explore}
		bc := exploration.SavedCapture{CaptureReceipt: b.Receipt, Explore: b.Explore}
		expected, e := exploration.CompareEvidence(&ac, &bc)
		wanted := *b.Receipt.Question.Explore
		wanted.Where = &exploration.Predicate{All: []exploration.Predicate{*wanted.Where, {Field: &exploration.Field{Kind: "attribute", Key: "http.route"}, Op: "eq", Value: "/new"}}}
		coherent := reflect.DeepEqual(l.Focused.Record, &wanted) && e == nil && reflect.DeepEqual(expected, l.Receipt.Result) && reflect.DeepEqual(l.Original.Cards[0], l.Saved.Cards[0]) && l.Focused.Validate() == nil && l.Focused.Scope.FromMS == b.Receipt.Question.Explore.FromMS && len(l.Receipt.Result.Breakdown.Rows) == 4 && a.Explore.Count == "1200" && b.Explore.Count == "1300"
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/comparison_breakdown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
