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

type activitySelectorViewStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Pinned bool   `json:"pinned"`
}

var activitySelectorViewActions = []string{"open", "start", "pause", "saveview", "catalog", "reopen", "restart", "stop"}

func checkActivitySelectorView(t *testing.T, c *validate.Checker, steps []activitySelectorViewStep) bool {
	t.Helper()
	if len(steps) != len(activitySelectorViewActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != activitySelectorViewActions[i] {
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
func validateActivitySelectorView(t *testing.T, steps []activitySelectorViewStep) {
	t.Helper()
	spec, e := filepath.Abs("model/activity_selector_view_283.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: activity_selector_view_283\nconstants: {MAX_READS: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range activitySelectorViewActions {
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
	if !checkActivitySelectorView(t, c, steps) {
		t.Fatal("Activity selector lifecycle refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]activitySelectorViewStep(nil), steps...)
		switch i {
		case 0:
			bad[0].Reads++
		case 1:
			bad[5].Staged = false
		case 2:
			bad[3].Pinned = false
		case 3:
			bad[4].Reads++
		case 4:
			bad[6].Reads--
		case 5:
			bad[0].Pinned = false
		}
		if checkActivitySelectorView(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkActivitySelectorView(t, c, steps[:7]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintActivitySelectorViewModelSmoke(t *testing.T) {
	reads := []int{0, 1, 1, 1, 1, 1, 2, 2}
	staged := []bool{true, false, false, false, false, true, false, false}
	steps := []activitySelectorViewStep{}
	for i, a := range activitySelectorViewActions {
		steps = append(steps, activitySelectorViewStep{Action: a, Reads: reads[i], Staged: staged[i], Pinned: true})
	}
	validateActivitySelectorView(t, steps)
}

func TestQuintActivitySelectorViewObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_ACTIVITY_SELECTOR_VIEW_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []activitySelectorViewStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateActivitySelectorView(t, steps)
}
