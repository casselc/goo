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

type scopeActionStep struct {
	Action   string `json:"action"`
	Fence    int    `json:"fence"`
	Reads    int    `json:"reads"`
	Staged   int    `json:"staged"`
	Eligible bool   `json:"eligible"`
	Accepted bool   `json:"accepted"`
}

// Fence is a projected count of observed execution-owner transitions, not a
// serialized capability or user-controlled resolver input. Source reads are
// Explore requests; local resolution and staging must not perform source IO.
func newScopeActionsChecker(t *testing.T) *validate.Checker {
	t.Helper()
	spec, err := filepath.Abs("model/scope_actions_266.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: scope_actions_266\nconstants: {MAX_READS: '3'}\ninvariants: [safety]\nactions:\n", spec)
	for _, action := range []string{"source", "rerun", "stale", "suspend", "suspended", "resume", "stage"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'fence == {{.obs.fence}} and reads == {{.obs.reads}} and staged == {{.obs.staged}} and eligible == {{.obs.eligible}} and accepted == {{.obs.accepted}}'\n", action)
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

func checkScopeActionLedger(t *testing.T, checker *validate.Checker, steps []scopeActionStep) bool {
	t.Helper()
	names := []string{"source", "rerun", "stale", "suspend", "suspended", "resume", "stage"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for i, step := range steps {
		if step.Action != names[i] {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.fence", step.Fence, "obs.reads", step.Reads, "obs.staged", step.Staged, "obs.eligible", step.Eligible, "obs.accepted", step.Accepted)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateScopeActionMutants(t *testing.T, steps []scopeActionStep) {
	t.Helper()
	checker := newScopeActionsChecker(t)
	if !checkScopeActionLedger(t, checker, steps) {
		t.Fatal("actual execution-scope trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]scopeActionStep(nil), steps...)
		switch mutant {
		case 0:
			altered[2].Accepted = true
		case 1:
			altered[2].Staged++
		case 2:
			altered[4].Accepted = true
		case 3:
			altered[4].Reads++
		case 4:
			altered[1].Fence--
		case 5:
			altered[5].Eligible = false
		case 6:
			altered[6].Staged = 0
		}
		if checkScopeActionLedger(t, checker, altered) {
			t.Fatal("incorrect execution-scope trace accepted", mutant)
		}
	}
	if checkScopeActionLedger(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix counted as completed journey")
	}
}

func TestQuintScopeActionsObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_SCOPE_ACTION_LEDGER")
	if file == "" {
		t.Skip("actual browser execution-scope ledger required; set GOO_SCOPE_ACTION_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []scopeActionStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateScopeActionMutants(t, steps)
}
