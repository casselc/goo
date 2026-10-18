package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"testing"
)

type recipeLedger struct {
	Reused                                        exploration.Query
	Q                                             exploration.Query
	Plan                                          exploration.RecipePlan
	Result                                        exploration.Result
	Capture                                       exploration.CapturePage
	Document                                      workbench.Document
	Before, AfterPreview, BeforeNotes, AfterNotes int
}

func TestQuintRecipesLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_RECIPES_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l recipeLedger
	if e = json.Unmarshal(data, &l); e != nil {
		t.Fatal(e)
	}
	replay := 0
	check := func(l recipeLedger) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("recipes"))
		expected, e := exploration.PlanRecipe(l.Plan.Recipe.ID, l.Q)
		good := e == nil && reflect.DeepEqual(expected, &l.Plan) && l.Before == l.AfterPreview
		rec.Record(context.Background(), "preview", nil, "obs.coherent", good)
		rec.Record(context.Background(), "execute", nil, "obs.coherent", reflect.DeepEqual(l.Result.Question, l.Plan.Question) && l.Result.Count != "0" && l.Result.Duration != nil)
		rec.Record(context.Background(), "capture", nil, "obs.coherent", l.Capture.Receipt.Question.Explore != nil && reflect.DeepEqual(*l.Capture.Receipt.Question.Explore, l.Plan.Question) && l.Capture.Explore != nil && reflect.DeepEqual(l.Capture.Explore.Question, l.Plan.Question))
		good = len(l.Document.Cards) == 2 && l.BeforeNotes == l.AfterNotes
		if len(l.Document.Cards) == 2 {
			good = good && l.Document.Cards[0].Role == "hypothesis" && l.Document.Cards[0].Reference == "" && l.Document.Cards[1].Reference == l.Capture.Receipt.ID
		}
		rec.Record(context.Background(), "document", nil, "obs.coherent", good)
		fresh := l.Reused
		good = fresh.ToMS >= l.Plan.Question.ToMS && fresh.ToMS-fresh.FromMS == l.Plan.Question.ToMS-l.Plan.Question.FromMS
		fresh.FromMS = l.Plan.Question.FromMS
		fresh.ToMS = l.Plan.Question.ToMS
		rec.Record(context.Background(), "reuse", nil, "obs.coherent", good && reflect.DeepEqual(fresh, l.Plan.Question))

		checker, e := validate.NewChecker("model/recipes.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		artifact(fmt.Sprintf("recipes-replay-%d.txt", replay), []byte(report.Summary()))
		replay++
		return report.OK()
	}
	if !check(l) {
		t.Fatal("live recipe ledger rejected")
	}
	for i := 0; i < 6; i++ {
		var bad recipeLedger
		json.Unmarshal(data, &bad)
		switch i {
		case 0:
			bad.Plan.Question.FromMS++
		case 1:
			bad.AfterPreview++
		case 2:
			bad.Result.Question.Environment = "different"
		case 3:
			bad.Document.Cards[1].Reference = "wrong"
		case 4:
			bad.Capture.Receipt.Question.Explore.ToMS++
		case 5:
			bad.Reused.Where = nil
		}
		if check(bad) {
			t.Fatal("negative accepted", i)
		}
		artifact(fmt.Sprintf("recipes-negative-%d.txt", i), []byte("rejected\n"))
	}
}
