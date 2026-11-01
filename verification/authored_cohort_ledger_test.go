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

type authoredCohortStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Compares  int    `json:"compares"`
	Preserved bool   `json:"preserved"`
}

var authoredCohortActions = []string{"narrow", "parent", "include", "stageParent", "parentCompared", "backParent", "stale", "excludeFacet", "stageRest", "restCompared", "save", "backRest", "reopen", "rerun"}

func checkAuthoredCohort(t *testing.T, c *validate.Checker, steps []authoredCohortStep) bool {
	t.Helper()
	if len(steps) != len(authoredCohortActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != authoredCohortActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.compares", s.Compares)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateAuthoredCohort(t *testing.T, steps []authoredCohortStep) {
	t.Helper()
	spec, e := filepath.Abs("model/authored_cohort_307.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: authored_cohort_307\nconstants: {MAX_READS: '8'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range authoredCohortActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and compares == {{.obs.compares}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkAuthoredCohort(t, checker, steps) {
		t.Fatal("authored cohort journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]authoredCohortStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Reads++
		case 1:
			bad[4].Reads++
		case 2:
			bad[6].Compares++
		case 3:
			bad[8].Reads++
		case 4:
			bad[9].Reads--
		case 5:
			bad[10].Saves--
		case 6:
			bad[11].Preserved = false
		}
		if checkAuthoredCohort(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkAuthoredCohort(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintAuthoredCohortModelSmoke(t *testing.T) {
	raw := `[{"action":"narrow","reads":3,"pages":1,"compares":0,"saves":0,"preserved":true},{"action":"parent","reads":4,"pages":1,"compares":0,"saves":1,"preserved":true},{"action":"include","reads":4,"pages":1,"compares":0,"saves":1,"preserved":true},{"action":"stageParent","reads":4,"pages":1,"compares":0,"saves":1,"preserved":true},{"action":"parentCompared","reads":5,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"backParent","reads":5,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"stale","reads":5,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"excludeFacet","reads":5,"pages":2,"compares":1,"saves":1,"preserved":true},{"action":"stageRest","reads":5,"pages":2,"compares":1,"saves":1,"preserved":true},{"action":"restCompared","reads":7,"pages":2,"compares":2,"saves":1,"preserved":true},{"action":"save","reads":7,"pages":2,"compares":2,"saves":2,"preserved":true},{"action":"backRest","reads":7,"pages":2,"compares":2,"saves":2,"preserved":true},{"action":"reopen","reads":7,"pages":2,"compares":2,"saves":2,"preserved":true},{"action":"rerun","reads":8,"pages":2,"compares":2,"saves":2,"preserved":true}]`
	var steps []authoredCohortStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateAuthoredCohort(t, steps)
}
func TestQuintAuthoredCohortObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_AUTHORED_COHORT_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []authoredCohortStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateAuthoredCohort(t, steps)
}
