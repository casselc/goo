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

type wholeCompositionStep struct {
	Action        string `json:"action"`
	Reads         int    `json:"reads"`
	Composed      bool   `json:"composed"`
	AggregateHeld bool   `json:"aggregateHeld"`
	Blocked       bool   `json:"blocked"`
	Preserved     bool   `json:"preserved"`
}

var wholeCompositionActions = []string{"read", "blocked", "apply", "page", "inspect", "save", "reopen", "rerun"}

func checkWholeComposition(t *testing.T, c *validate.Checker, steps []wholeCompositionStep) bool {
	t.Helper()
	if len(steps) != len(wholeCompositionActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != wholeCompositionActions[i] || !s.Preserved || !s.AggregateHeld || i == 1 && !s.Blocked {
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
func validateWholeComposition(t *testing.T, steps []wholeCompositionStep) {
	t.Helper()
	spec, e := filepath.Abs("model/whole_composition_280.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: whole_composition_280\nconstants: {MAX_READS: '3'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range wholeCompositionActions {
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
	if !checkWholeComposition(t, c, steps) {
		t.Fatal("view composition journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]wholeCompositionStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[4].Composed = false
		case 2:
			bad[4].AggregateHeld = false
		case 3:
			bad[6].Reads++
		case 4:
			bad[7].Reads--
		case 5:
			bad[1].Blocked = false
		}
		if checkWholeComposition(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkWholeComposition(t, c, steps[:7]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintWholeCompositionModelSmoke(t *testing.T) {
	steps := []wholeCompositionStep{}
	for i, a := range wholeCompositionActions {
		reads := 1
		if i >= 3 {
			reads = 2
		}
		if i == 7 {
			reads = 3
		}
		steps = append(steps, wholeCompositionStep{Action: a, Reads: reads, Composed: i >= 2, AggregateHeld: true, Blocked: i == 1, Preserved: true})
	}
	validateWholeComposition(t, steps)
}
func TestQuintWholeCompositionObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_WHOLE_COMPOSITION_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []wholeCompositionStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateWholeComposition(t, steps)
}
