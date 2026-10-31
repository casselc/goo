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

type viewMeasuresStep struct {
	Action  string `json:"action"`
	Reads   int    `json:"reads"`
	Value   string `json:"value"`
	Scale   string `json:"scale"`
	Blocked bool   `json:"blocked"`
}

var measureActions = []string{"staged", "read", "numeric", "duration", "invalid", "apply", "save", "reopen", "rerun"}

func checkViewMeasures(t *testing.T, c *validate.Checker, steps []viewMeasuresStep) bool {
	t.Helper()
	if len(steps) != len(measureActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != measureActions[i] {
			return false
		}
		measure, ok := map[string]int{"count": 0, "numeric.mean": 1, "duration.p95": 2}[s.Value]
		if !ok || measure != 0 && s.Scale != "max" {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.measure", measure, "obs.blocked", s.Blocked)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateViewMeasures(t *testing.T, steps []viewMeasuresStep) {
	t.Helper()
	spec, e := filepath.Abs("model/view_measures_276.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: view_measures_276\nconstants: {MAX_READS: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range measureActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and measure == {{.obs.measure}} and blocked == {{.obs.blocked}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkViewMeasures(t, c, steps) {
		t.Fatal("saved measure journey refused")
	}
	for mutant := 0; mutant < 6; mutant++ {
		bad := append([]viewMeasuresStep(nil), steps...)
		switch mutant {
		case 0:
			bad[2].Reads++
		case 1:
			bad[7].Reads++
		case 2:
			bad[4].Blocked = false
		case 3:
			bad[8].Value = "numeric.mean"
		case 4:
			bad[3].Scale = "population"
		case 5:
			bad[8].Reads--
		}
		if checkViewMeasures(t, c, bad) {
			t.Fatal("mutant passed", mutant)
		}
	}
	if checkViewMeasures(t, c, steps[:8]) {
		t.Fatal("unfinished prefix passed")
	}
}
func TestQuintViewMeasuresModelSmoke(t *testing.T) {
	steps := []viewMeasuresStep{}
	for i, a := range measureActions {
		s := viewMeasuresStep{Action: a, Scale: "max", Value: "duration.p95", Reads: 1, Blocked: i == 4}
		if i < 2 {
			s.Value = "count"
		}
		if i == 0 {
			s.Reads = 0
		}
		if i == 2 {
			s.Value = "numeric.mean"
		}
		if i == 8 {
			s.Reads = 2
		}
		steps = append(steps, s)
	}
	validateViewMeasures(t, steps)
}
func TestQuintViewMeasuresObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_VIEW_MEASURES_LEDGER")
	if file == "" {
		t.Skip("real observed browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []viewMeasuresStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateViewMeasures(t, steps)
}
