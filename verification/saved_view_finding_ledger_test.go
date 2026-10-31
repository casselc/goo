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

type savedViewFindingStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Pinned bool   `json:"pinned"`
}

var savedViewFindingActions = []string{"read", "dirty", "saveview", "attach", "savedoc", "reopen", "definition", "execute"}

func checkSavedViewFinding(t *testing.T, c *validate.Checker, steps []savedViewFindingStep) bool {
	t.Helper()
	if len(steps) != len(savedViewFindingActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != savedViewFindingActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.staged", s.Staged, "obs.pinned", s.Pinned)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateSavedViewFinding(t *testing.T, steps []savedViewFindingStep) {
	t.Helper()
	spec, e := filepath.Abs("model/saved_view_finding_282.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: saved_view_finding_282\nconstants: {MAX_READS: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range savedViewFindingActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and staged == {{.obs.staged}} and pinned == {{.obs.pinned}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkSavedViewFinding(t, c, steps) {
		t.Fatal("view composition journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]savedViewFindingStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[6].Staged = false
		case 2:
			bad[3].Pinned = false
		case 3:
			bad[5].Reads++
		case 4:
			bad[7].Reads--
		case 5:
			bad[1].Pinned = true
		}
		if checkSavedViewFinding(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkSavedViewFinding(t, c, steps[:7]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintSavedViewFindingModelSmoke(t *testing.T) {
	reads := []int{1, 1, 1, 1, 1, 1, 1, 2}
	staged := []bool{false, false, false, false, false, false, true, false}
	steps := []savedViewFindingStep{}
	for i, a := range savedViewFindingActions {
		steps = append(steps, savedViewFindingStep{Action: a, Reads: reads[i], Staged: staged[i], Pinned: i >= 3})
	}
	validateSavedViewFinding(t, steps)
}

func TestQuintSavedViewFindingObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_SAVED_VIEW_FINDING_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []savedViewFindingStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateSavedViewFinding(t, steps)
}
