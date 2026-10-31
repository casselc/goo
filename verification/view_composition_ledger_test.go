package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"testing"
)

type viewCompositionStep struct {
	Action        string `json:"action"`
	Reads         int    `json:"reads"`
	Composed      bool   `json:"composed"`
	AggregateHeld bool   `json:"aggregateHeld"`
	Blocked       bool   `json:"blocked"`
	Preserved     bool   `json:"preserved"`
}

var viewCompositionActions = []string{"read", "compose", "edit", "invalid", "page", "save", "reopen", "rerun"}

func checkViewComposition(t *testing.T, c *validate.Checker, steps []viewCompositionStep) bool {
	t.Helper()
	if len(steps) != len(viewCompositionActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != viewCompositionActions[i] || !s.Preserved || !s.AggregateHeld || i == 3 && !s.Blocked {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.composed", s.Composed, "obs.aggregateHeld", s.AggregateHeld)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateViewComposition(t *testing.T, steps []viewCompositionStep) {
	t.Helper()
	spec, e := filepath.Abs("model/view_composition_279.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: view_composition_279\nconstants: {MAX_READS: '3'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range viewCompositionActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and composed == {{.obs.composed}} and aggregateHeld == {{.obs.aggregateHeld}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkViewComposition(t, c, steps) {
		t.Fatal("view composition journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]viewCompositionStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[1].Composed = false
		case 2:
			bad[4].AggregateHeld = false
		case 3:
			bad[6].Reads++
		case 4:
			bad[7].Reads--
		case 5:
			bad[3].Blocked = false
		}
		if checkViewComposition(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkViewComposition(t, c, steps[:7]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintViewCompositionModelSmoke(t *testing.T) {
	steps := []viewCompositionStep{}
	for i, a := range viewCompositionActions {
		reads := 1
		if i >= 4 {
			reads = 2
		}
		if i == 7 {
			reads = 3
		}
		steps = append(steps, viewCompositionStep{Action: a, Reads: reads, Composed: i >= 1, AggregateHeld: true, Blocked: i == 3, Preserved: true})
	}
	validateViewComposition(t, steps)
}
func TestQuintViewCompositionObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_VIEW_COMPOSITION_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []viewCompositionStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateViewComposition(t, steps)
}
