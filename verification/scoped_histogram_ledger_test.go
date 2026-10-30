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

type scopedHistogramStep struct {
	Action   string `json:"action"`
	Reads    int    `json:"reads"`
	Plans    int    `json:"plans"`
	Total    int    `json:"total"`
	Valid    int    `json:"valid"`
	Selected int    `json:"selected"`
	Edge     string `json:"edge"`
}

// This deliberately small model projects observed display/effect conservation.
// It does not inspect or prove private completeness witnesses or host bindings.
func newScopedHistogramChecker(t *testing.T) *validate.Checker {
	t.Helper()
	spec, err := filepath.Abs("model/scoped_histogram_269.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: scoped_histogram_269\nconstants: {ORIGINAL_TOTAL: '18'}\ninvariants: [safety]\nactions:\n", spec)
	for _, action := range []string{"source", "brush", "table", "rerun", "stale", "choose", "execute"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and plans == {{.obs.plans}} and total == {{.obs.total}} and valid == {{.obs.valid}} and selected == {{.obs.selected}} and edge == {{.obs.edge}}'\n", action)
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

func checkScopedHistogramLedger(t *testing.T, checker *validate.Checker, steps []scopedHistogramStep) bool {
	t.Helper()
	names := []string{"source", "brush", "table", "rerun", "stale", "choose", "execute"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, step := range steps {
		if step.Action != names[i] {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.plans", step.Plans, "obs.total", step.Total, "obs.valid", step.Valid, "obs.selected", step.Selected, "obs.edge", step.Edge)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateScopedHistogramMutants(t *testing.T, steps []scopedHistogramStep) {
	t.Helper()
	checker := newScopedHistogramChecker(t)
	if !checkScopedHistogramLedger(t, checker, steps) {
		t.Fatal("actual scoped-histogram trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]scopedHistogramStep(nil), steps...)
		switch mutant {
		case 0:
			altered[1].Edge = "9223372036854775807"
		case 1:
			altered[2].Valid = 18
		case 2:
			altered[3].Reads--
		case 3:
			altered[4].Plans++
		case 4:
			altered[5].Reads++
		case 5:
			altered[6].Total = 13
		case 6:
			altered[0].Plans++
		}
		if checkScopedHistogramLedger(t, checker, altered) {
			t.Fatal("incorrect scoped-histogram trace accepted", mutant)
		}
	}
	if checkScopedHistogramLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix counted as completed journey")
	}
}

func TestQuintScopedHistogramObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_SCOPED_HISTOGRAM_LEDGER")
	if file == "" {
		t.Skip("actual browser scoped-histogram ledger required; set GOO_SCOPED_HISTOGRAM_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []scopedHistogramStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateScopedHistogramMutants(t, steps)
}
