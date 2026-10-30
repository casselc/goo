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

type distributionShareStep struct {
	Action        string `json:"action"`
	Reads         int    `json:"reads"`
	Denominator   int    `json:"denominator"`
	WorkerPercent int    `json:"workerPercent"`
	Staged        bool   `json:"staged"`
}

// This deliberately small model projects observed display/effect conservation.
// It does not inspect or prove private completeness witnesses or host bindings.
func newDistributionSharesChecker(t *testing.T) *validate.Checker {
	t.Helper()
	spec, err := filepath.Abs("model/distribution_shares_267.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: distribution_shares_267\nconstants: {ORIGINAL_TOTAL: '4'}\ninvariants: [safety]\nactions:\n", spec)
	for _, action := range []string{"source", "find", "sort", "measure", "rerun", "stale", "stage", "execute"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and denominator == {{.obs.denominator}} and workerPercent == {{.obs.workerPercent}} and staged == {{.obs.staged}}'\n", action)
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

func checkDistributionShareLedger(t *testing.T, checker *validate.Checker, steps []distributionShareStep) bool {
	t.Helper()
	names := []string{"source", "find", "sort", "measure", "rerun", "stale", "stage", "execute"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, step := range steps {
		if step.Action != names[i] {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.denominator", step.Denominator, "obs.workerPercent", step.WorkerPercent, "obs.staged", step.Staged)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateDistributionShareMutants(t *testing.T, steps []distributionShareStep) {
	t.Helper()
	checker := newDistributionSharesChecker(t)
	if !checkDistributionShareLedger(t, checker, steps) {
		t.Fatal("actual distribution-share trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]distributionShareStep(nil), steps...)
		switch mutant {
		case 0:
			altered[1].Denominator = 1
		case 1:
			altered[1].WorkerPercent = 100
		case 2:
			altered[2].Reads++
		case 3:
			altered[4].Reads--
		case 4:
			altered[5].Staged = true
		case 5:
			altered[6].Reads++
		case 6:
			altered[7].Denominator = 4
		}
		if checkDistributionShareLedger(t, checker, altered) {
			t.Fatal("incorrect distribution-share trace accepted", mutant)
		}
	}
	if checkDistributionShareLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix counted as completed journey")
	}
}

func TestQuintDistributionSharesObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_DISTRIBUTION_SHARE_LEDGER")
	if file == "" {
		t.Skip("actual browser distribution-share ledger required; set GOO_DISTRIBUTION_SHARE_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []distributionShareStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateDistributionShareMutants(t, steps)
}
