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

type recordStepHandoffStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Pinned bool   `json:"pinned"`
}

var recordStepHandoffActions = []string{"open", "start", "analyze", "execute", "share", "link", "pin", "recover", "savedoc", "question", "rerun"}

func checkRecordStepHandoff(t *testing.T, c *validate.Checker, steps []recordStepHandoffStep) bool {
	t.Helper()
	if len(steps) != len(recordStepHandoffActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != recordStepHandoffActions[i] {
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
func validateRecordStepHandoff(t *testing.T, steps []recordStepHandoffStep) {
	t.Helper()
	spec, e := filepath.Abs("model/record_step_handoff_284.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: record_step_handoff_284\nconstants: {MAX_READS: '3'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range recordStepHandoffActions {
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
	if !checkRecordStepHandoff(t, c, steps) {
		t.Fatal("Record step handoff refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]recordStepHandoffStep(nil), steps...)
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
			bad[10].Reads--
		case 5:
			bad[0].Pinned = false
		}
		if checkRecordStepHandoff(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkRecordStepHandoff(t, c, steps[:10]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintRecordStepHandoffModelSmoke(t *testing.T) {
	reads := []int{0, 1, 1, 2, 2, 2, 2, 2, 2, 2, 3}
	staged := []bool{true, false, true, false, false, true, true, true, true, true, false}
	steps := []recordStepHandoffStep{}
	for i, a := range recordStepHandoffActions {
		steps = append(steps, recordStepHandoffStep{Action: a, Reads: reads[i], Staged: staged[i], Pinned: true})
	}
	validateRecordStepHandoff(t, steps)
}

func TestQuintRecordStepHandoffObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_RECORD_STEP_HANDOFF_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []recordStepHandoffStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordStepHandoff(t, steps)
}
