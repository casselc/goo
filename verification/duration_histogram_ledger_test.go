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

type durationHistogramStep struct {
	Action   string `json:"action"`
	Reads    int    `json:"reads"`
	Pairs    int    `json:"pairs"`
	Total    int    `json:"total"`
	Range    int    `json:"range"`
	Upper    string `json:"upper"`
	Child    int    `json:"child"`
	Selected int    `json:"selected"`
	Rest     int    `json:"rest"`
}

// This deliberately small model projects observed display/effect conservation.
// It does not inspect or prove private completeness witnesses or host bindings.
func newDurationHistogramChecker(t *testing.T) *validate.Checker {
	t.Helper()
	spec, err := filepath.Abs("model/duration_histogram_271.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: duration_histogram_271\nconstants: {GROUP_TOTAL: '11'}\ninvariants: [safety]\nactions:\n", spec)
	for _, action := range []string{"source", "brush", "manual", "execute", "rerun", "stale", "rest", "compare"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pairs == {{.obs.pairs}}'\n", action)
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

func checkDurationHistogramLedger(t *testing.T, checker *validate.Checker, steps []durationHistogramStep) bool {
	t.Helper()
	names := []string{"source", "brush", "manual", "execute", "rerun", "stale", "rest", "compare"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, step := range steps {
		if step.Action != names[i] {
			return false
		}
		if step.Total != 11 || step.Range != 2 || step.Upper != "+inf" || step.Action == "execute" && step.Child != 1 || step.Action == "compare" && (step.Selected != 2 || step.Rest != 9) {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.pairs", step.Pairs)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateDurationHistogramMutants(t *testing.T, steps []durationHistogramStep) {
	t.Helper()
	checker := newDurationHistogramChecker(t)
	if !checkDurationHistogramLedger(t, checker, steps) {
		t.Fatal("actual duration-histogram trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]durationHistogramStep(nil), steps...)
		switch mutant {
		case 0:
			altered[1].Upper = "1000000000"
		case 1:
			altered[2].Reads++
		case 2:
			altered[0].Total = 1
		case 3:
			altered[3].Child = 2
		case 4:
			altered[5].Reads++
		case 5:
			altered[7].Reads--
		case 6:
			altered[7].Rest = 10
		}
		if checkDurationHistogramLedger(t, checker, altered) {
			t.Fatal("incorrect duration-histogram trace accepted", mutant)
		}
	}
	if checkDurationHistogramLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix counted as completed journey")
	}
}

func TestQuintDurationHistogramObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_DURATION_HISTOGRAM_LEDGER")
	if file == "" {
		t.Skip("actual browser duration-histogram ledger required; set GOO_DURATION_HISTOGRAM_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []durationHistogramStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateDurationHistogramMutants(t, steps)
}
