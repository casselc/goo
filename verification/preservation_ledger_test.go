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

type preservationStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Captures  int    `json:"captures"`
	Staged    bool   `json:"staged"`
	Questions int    `json:"questions"`
	Views     int    `json:"views"`
	Evidence  int    `json:"evidence"`
}

var preservationActions = []string{"openQuestion", "shareQuestion", "pinQuestion", "saveQuestion", "queryQuestion", "openCapture", "captureEvidence", "reopenEvidence", "pinEvidence", "multiView", "chooseSignal", "singleView", "saveView", "pinView", "saveFinding"}

func checkPreservation(t *testing.T, c *validate.Checker, steps []preservationStep) bool {
	t.Helper()
	if len(steps) != len(preservationActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != preservationActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.staged", s.Staged, "obs.captures", s.Captures, "obs.questions", s.Questions, "obs.views", s.Views, "obs.evidence", s.Evidence)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validatePreservation(t *testing.T, steps []preservationStep) {
	t.Helper()
	spec, e := filepath.Abs("model/preservation_288.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: preservation_288\nconstants: {MAX_READS: '1'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range preservationActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and staged == {{.obs.staged}} and captures == {{.obs.captures}} and questions == {{.obs.questions}} and views == {{.obs.views}} and evidence == {{.obs.evidence}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkPreservation(t, c, steps) {
		t.Fatal("preservation journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]preservationStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[5].Captures++
		case 2:
			bad[6].Captures--
		case 3:
			bad[9].Views++
		case 4:
			bad[13].Questions--
		case 5:
			bad[14].Evidence--
		}
		if checkPreservation(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkPreservation(t, c, steps[:14]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintPreservationModelSmoke(t *testing.T) {
	reads := []int{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	captures := []int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	staged := []bool{true, true, true, true, false, false, false, false, false, true, true, true, true, true, true}
	questions := []int{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	views := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	evidence := []int{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1}
	steps := []preservationStep{}
	for i, a := range preservationActions {
		steps = append(steps, preservationStep{Action: a, Reads: reads[i], Captures: captures[i], Staged: staged[i], Questions: questions[i], Views: views[i], Evidence: evidence[i]})
	}
	validatePreservation(t, steps)
}

func TestQuintPreservationObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_PRESERVATION_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []preservationStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validatePreservation(t, steps)
}
