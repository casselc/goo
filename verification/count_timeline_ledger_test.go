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

type countTimelineStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Held   bool   `json:"held"`
}

var countTimelineActions = []string{"read", "configure", "execute", "page", "stage", "narrow", "reset", "save", "reopen", "rerun"}

func checkCountTimeline(t *testing.T, c *validate.Checker, steps []countTimelineStep) bool {
	t.Helper()
	if len(steps) != len(countTimelineActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != countTimelineActions[i] || !s.Held {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.staged", s.Staged, "obs.held", s.Held)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateCountTimeline(t *testing.T, steps []countTimelineStep) {
	t.Helper()
	spec, e := filepath.Abs("model/count_timeline_281.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: count_timeline_281\nconstants: {MAX_READS: '5'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range countTimelineActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and staged == {{.obs.staged}} and held == {{.obs.held}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkCountTimeline(t, c, steps) {
		t.Fatal("view composition journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]countTimelineStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[4].Staged = false
		case 2:
			bad[3].Held = false
		case 3:
			bad[8].Reads++
		case 4:
			bad[9].Reads--
		case 5:
			bad[6].Staged = false
		}
		if checkCountTimeline(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkCountTimeline(t, c, steps[:9]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintCountTimelineModelSmoke(t *testing.T) {
	reads := []int{1, 1, 2, 3, 3, 4, 4, 4, 4, 5}
	staged := []bool{false, true, false, false, true, false, true, true, true, false}
	steps := []countTimelineStep{}
	for i, a := range countTimelineActions {
		steps = append(steps, countTimelineStep{Action: a, Reads: reads[i], Staged: staged[i], Held: true})
	}
	validateCountTimeline(t, steps)
}

func TestQuintCountTimelineObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_COUNT_TIMELINE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []countTimelineStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateCountTimeline(t, steps)
}
