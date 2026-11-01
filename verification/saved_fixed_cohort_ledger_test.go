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

type savedFixedCohortStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Compares  int    `json:"compares"`
	Preserved bool   `json:"preserved"`
}

var savedFixedCohortActions = []string{"narrow", "parent", "restCompared", "promote", "save", "stale", "reopen", "freshRun", "encodeArtifact", "importDefinition", "openImported", "runImported"}

func checkSavedFixedCohort(t *testing.T, c *validate.Checker, steps []savedFixedCohortStep) bool {
	t.Helper()
	if len(steps) != len(savedFixedCohortActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != savedFixedCohortActions[i] || !s.Preserved {
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
func validateSavedFixedCohort(t *testing.T, steps []savedFixedCohortStep) {
	t.Helper()
	spec, e := filepath.Abs("model/saved_fixed_cohort_308.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: saved_fixed_cohort_308\nconstants: {MAX_READS: '10'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range savedFixedCohortActions {
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
	if !checkSavedFixedCohort(t, checker, steps) {
		t.Fatal("authored cohort journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]savedFixedCohortStep(nil), steps...)
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
		if checkSavedFixedCohort(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkSavedFixedCohort(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintSavedFixedCohortModelSmoke(t *testing.T) {
	raw := `[{"action":"narrow","reads":3,"pages":1,"compares":0,"saves":0,"preserved":true},{"action":"parent","reads":4,"pages":1,"compares":0,"saves":1,"preserved":true},{"action":"restCompared","reads":6,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"promote","reads":6,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"save","reads":6,"pages":1,"compares":1,"saves":2,"preserved":true},{"action":"stale","reads":6,"pages":1,"compares":1,"saves":2,"preserved":true},{"action":"reopen","reads":6,"pages":1,"compares":1,"saves":2,"preserved":true},{"action":"freshRun","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"encodeArtifact","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"importDefinition","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"openImported","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"runImported","reads":10,"pages":1,"compares":3,"saves":2,"preserved":true}]`
	var steps []savedFixedCohortStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateSavedFixedCohort(t, steps)
}
func TestQuintSavedFixedCohortObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_SAVED_FIXED_COHORT_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []savedFixedCohortStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateSavedFixedCohort(t, steps)
}
