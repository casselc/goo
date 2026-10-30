package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type statusFractionStep struct {
	Action      string `json:"action"`
	Reads       int    `json:"reads"`
	Population  int    `json:"population"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Percent     int    `json:"percent"`
	Staged      bool   `json:"staged"`
}

// This deliberately small model projects observed display/effect conservation.
// It does not inspect or prove private completeness witnesses or host bindings.
func newStatusFractionsChecker(t *testing.T) *validate.Checker {
	t.Helper()
	spec, err := filepath.Abs("model/status_fractions_268.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: status_fractions_268\nconstants: {INITIAL_POPULATION: '103', GROUP_POPULATION: '100'}\ninvariants: [safety]\nactions:\n", spec)
	for _, action := range []string{"source", "find", "sort", "table", "rerun", "stale", "stage", "execute"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and population == {{.obs.population}} and numerator == {{.obs.numerator}} and denominator == {{.obs.denominator}} and percent == {{.obs.percent}} and staged == {{.obs.staged}}'\n", action)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	return checker
}

func checkStatusFractionLedger(t *testing.T, checker *validate.Checker, steps []statusFractionStep) bool {
	t.Helper()
	names := []string{"source", "find", "sort", "table", "rerun", "stale", "stage", "execute"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, step := range steps {
		if step.Action != names[i] {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.population", step.Population, "obs.numerator", step.Numerator, "obs.denominator", step.Denominator, "obs.percent", step.Percent, "obs.staged", step.Staged)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateStatusFractionMutants(t *testing.T, steps []statusFractionStep) {
	t.Helper()
	checker := newStatusFractionsChecker(t)
	if !checkStatusFractionLedger(t, checker, steps) {
		t.Fatal("actual status-fraction trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]statusFractionStep(nil), steps...)
		switch mutant {
		case 0:
			altered[1].Denominator = 103
		case 1:
			altered[1].Percent = 100
		case 2:
			altered[2].Reads++
		case 3:
			altered[4].Reads--
		case 4:
			altered[5].Staged = true
		case 5:
			altered[3].Numerator = 0
		case 6:
			altered[7].Population = 103
		}
		if checkStatusFractionLedger(t, checker, altered) {
			t.Fatal("incorrect status-fraction trace accepted", mutant)
		}
	}
	if checkStatusFractionLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix counted as completed journey")
	}
}

func TestQuintStatusFractionsObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_STATUS_FRACTION_LEDGER")
	if file == "" {
		t.Skip("actual browser status-fraction ledger required; set GOO_STATUS_FRACTION_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []statusFractionStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateStatusFractionMutants(t, steps)
}
