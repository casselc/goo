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

type focusAuthoringStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Plans  int    `json:"plans"`
}

var focusAuthoringActions = []string{"open", "recent", "pageRecords", "focusLogs", "authorLogs", "readLogs", "saveLogs", "returnParent", "focusSpans", "groupSpans", "authorSpans", "readSpans", "saveSpans", "reopen", "finish"}

func checkFocusAuthoring(t *testing.T, c *validate.Checker, steps []focusAuthoringStep) bool {
	t.Helper()
	if len(steps) != len(focusAuthoringActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != focusAuthoringActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.staged", s.Staged, "obs.plans", s.Plans)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateFocusAuthoring(t *testing.T, steps []focusAuthoringStep) {
	t.Helper()
	spec, e := filepath.Abs("model/focus_authoring_286.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: focus_authoring_286\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range focusAuthoringActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and staged == {{.obs.staged}} and plans == {{.obs.plans}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkFocusAuthoring(t, c, steps) {
		t.Fatal("Focused authoring refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]focusAuthoringStep(nil), steps...)
		switch i {
		case 0:
			bad[0].Reads++
		case 1:
			bad[1].Plans--
		case 2:
			bad[3].Staged = false
		case 3:
			bad[3].Reads++
		case 4:
			bad[14].Reads--
		case 5:
			bad[3].Plans++
		}
		if checkFocusAuthoring(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkFocusAuthoring(t, c, steps[:14]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintFocusAuthoringModelSmoke(t *testing.T) {
	reads := []int{0, 2, 3, 3, 3, 4, 4, 4, 4, 4, 4, 5, 5, 5, 6}
	staged := []bool{true, false, false, true, true, false, false, false, true, true, true, false, false, true, false}
	steps := []focusAuthoringStep{}
	for i, a := range focusAuthoringActions {
		plans := 1
		if i == 0 {
			plans = 0
		}
		steps = append(steps, focusAuthoringStep{Action: a, Reads: reads[i], Staged: staged[i], Plans: plans})
	}
	validateFocusAuthoring(t, steps)
}

func TestQuintFocusAuthoringObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_FOCUS_AUTHORING_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []focusAuthoringStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateFocusAuthoring(t, steps)
}
