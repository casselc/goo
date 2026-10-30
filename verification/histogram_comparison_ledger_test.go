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

type histogramComparisonStep struct {
	Action string   `json:"action"`
	Reads  int      `json:"reads"`
	Plans  int      `json:"plans"`
	Total  []int    `json:"total"`
	Valid  []int    `json:"valid"`
	Counts []int    `json:"counts"`
	Shares []string `json:"shares"`
	Scale  string   `json:"scale"`
	Edge   string   `json:"edge"`
}

// This deliberately small model projects observed display/effect conservation.
// It does not inspect or prove private completeness witnesses or host bindings.
func newHistogramComparisonChecker(t *testing.T) *validate.Checker {
	t.Helper()
	spec, err := filepath.Abs("model/histogram_comparison_270.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: histogram_comparison_270\nconstants: {CAPTURED_TOTAL: '4'}\ninvariants: [safety]\nactions:\n", spec)
	for _, action := range []string{"source", "population", "valid", "table", "brush", "open"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and plans == {{.obs.plans}} and scale == {{.obs.scale}} and baselineShare == {{.obs.baselineShare}} and selectionShare == {{.obs.selectionShare}}'\n", action)
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

func checkHistogramComparisonLedger(t *testing.T, checker *validate.Checker, steps []histogramComparisonStep) bool {
	t.Helper()
	names := []string{"source", "population", "valid", "table", "brush", "open"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, step := range steps {
		if step.Action != names[i] {
			return false
		}
		if len(step.Total) != 2 || len(step.Valid) != 2 || len(step.Counts) != 2 || len(step.Shares) != 2 || step.Total[0] != 4 || step.Total[1] != 4 || step.Valid[0] != 2 || step.Valid[1] != 3 || step.Counts[0] != 1 || step.Counts[1] != 2 || step.Edge != "9223372036854775808" {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.plans", step.Plans, "obs.scale", step.Scale, "obs.baselineShare", step.Shares[0], "obs.selectionShare", step.Shares[1])
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateHistogramComparisonMutants(t *testing.T, steps []histogramComparisonStep) {
	t.Helper()
	checker := newHistogramComparisonChecker(t)
	if !checkHistogramComparisonLedger(t, checker, steps) {
		t.Fatal("actual histogram-comparison trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]histogramComparisonStep(nil), steps...)
		switch mutant {
		case 0:
			altered[1].Edge = "9223372036854775807"
		case 1:
			altered[2].Scale = "population-share"
		case 2:
			altered[3].Reads++
		case 3:
			altered[4].Plans++
		case 4:
			altered[5].Reads++
		case 5:
			altered[5].Reads--
		case 6:
			altered[0].Plans++
		}
		if checkHistogramComparisonLedger(t, checker, altered) {
			t.Fatal("incorrect histogram-comparison trace accepted", mutant)
		}
	}
	if checkHistogramComparisonLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix counted as completed journey")
	}
}

func TestQuintHistogramComparisonObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_HISTOGRAM_COMPARISON_LEDGER")
	if file == "" {
		t.Skip("actual browser histogram-comparison ledger required; set GOO_HISTOGRAM_COMPARISON_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []histogramComparisonStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateHistogramComparisonMutants(t, steps)
}
