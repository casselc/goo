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

type durationTimelineStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Held   bool   `json:"held"`
}

var durationTimelineActions = []string{"read", "chooseMean", "chooseP95", "chooseTable", "pageRecords", "chooseBars", "previewRange", "stageWindow", "narrow", "resetWindow", "saveView", "reopenView", "rerun"}

func checkDurationTimeline(t *testing.T, c *validate.Checker, steps []durationTimelineStep) bool {
	t.Helper()
	if len(steps) != len(durationTimelineActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != durationTimelineActions[i] || !s.Held {
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
func validateDurationTimeline(t *testing.T, steps []durationTimelineStep) {
	t.Helper()
	spec, e := filepath.Abs("model/duration_timeline_287.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: duration_timeline_287\nconstants: {MAX_READS: '4'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range durationTimelineActions {
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
	if !checkDurationTimeline(t, c, steps) {
		t.Fatal("view composition journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]durationTimelineStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[2].Staged = true
		case 2:
			bad[4].Held = false
		case 3:
			bad[7].Staged = false
		case 4:
			bad[9].Staged = false
		case 5:
			bad[12].Reads--
		}
		if checkDurationTimeline(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkDurationTimeline(t, c, steps[:12]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintDurationTimelineModelSmoke(t *testing.T) {
	reads := []int{1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4}
	staged := []bool{false, false, false, false, false, false, false, true, false, true, true, true, false}
	steps := []durationTimelineStep{}
	for i, a := range durationTimelineActions {
		steps = append(steps, durationTimelineStep{Action: a, Reads: reads[i], Staged: staged[i], Held: true})
	}
	validateDurationTimeline(t, steps)
}

func TestQuintDurationTimelineObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_DURATION_TIMELINE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []durationTimelineStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateDurationTimeline(t, steps)
}
