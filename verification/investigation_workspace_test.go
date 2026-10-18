package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

func TestQuintInvestigationWorkspaceLedger(t *testing.T) {
	name := os.Getenv("GOO_INVESTIGATION_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type ledger struct {
		First, Second, Logs, Final, Fresh, Reopened workbench.Document
		Captures                                    []exploration.CapturePage
		Before, After                               int
	}
	var original ledger
	if e = json.Unmarshal(data, &original); e != nil {
		t.Fatal(e)
	}
	if len(original.Captures) != 4 {
		t.Fatal("require four actual captures")
	}
	check := func(mutation int) bool {
		var l ledger
		json.Unmarshal(data, &l)
		switch mutation {
		case 0:
			l.Second.Parent = "changed"
		case 1:
			l.Logs.Cards[1].Reference = "changed"
		case 2:
			l.Final.Question = "changed"
		case 3:
			l.Fresh.Cards[2].Reference = "changed"
		case 4:
			l.Reopened.Finding = "changed"
		case 5:
			l.After++
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, v bool) { rec.Record(context.Background(), a, nil, "obs.coherent", v) }
		record("begin", l.First.DocumentInput.Validate() == nil && l.First.Question != "" && len(l.First.Cards) == 0)
		child := func(d, p workbench.Document, n int) bool {
			return d.Parent == p.ID && d.Root == l.First.ID && d.Question == l.First.Question && len(d.Cards) == n && reflect.DeepEqual(d.Cards[:len(p.Cards)], p.Cards)
		}
		record("spans", child(l.Second, l.First, 1) && l.Second.Cards[0].Reference == l.Captures[0].Receipt.ID && l.Captures[0].Explore != nil && l.Captures[0].Explore.Question.Signal == "spans")
		record("logs", child(l.Logs, l.Second, 2) && l.Logs.Cards[1].Reference == l.Captures[1].Receipt.ID && l.Captures[1].Explore != nil && l.Captures[1].Explore.Question.Signal == "logs")
		record("metrics", child(l.Final, l.Logs, 3) && l.Final.Finding != "" && l.Final.OpenQuestions != "" && l.Final.Cards[2].Reference == l.Captures[2].Receipt.ID && l.Captures[2].Measurement != nil && measurements.ValidateResult(l.Captures[2].Measurement) == nil)
		a, b := l.Captures[2].Measurement, l.Captures[3].Measurement
		record("reuse", child(l.Fresh, l.Final, 4) && l.Fresh.Cards[3].Reference == l.Captures[3].Receipt.ID && a != nil && b != nil && reflect.DeepEqual(a.Question.Definition, b.Question.Definition) && b.Question.ToMS >= a.Question.ToMS && measurements.ValidateResult(b) == nil)
		record("reopen", reflect.DeepEqual(l.Final, l.Reopened) && l.Before == l.After)
		c, e := validate.NewChecker("model/investigation_workspace.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("actual investigation ledger rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutated ledger accepted", i)
		}
	}
}
