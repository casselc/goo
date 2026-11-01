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

type pairedWindowStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Compares  int    `json:"compares"`
	Preserved bool   `json:"preserved"`
}

var pairedWindowActions = []string{"narrow", "parent", "restCompared", "promote", "save", "originalRun", "invalidDraft", "discardDraft", "stageWindow", "stale", "saveWindow", "reopen", "newRun", "encodeArtifact"}

func checkPairedWindow(t *testing.T, c *validate.Checker, steps []pairedWindowStep) bool {
	t.Helper()
	if len(steps) != len(pairedWindowActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != pairedWindowActions[i] || !s.Preserved {
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
func validatePairedWindow(t *testing.T, steps []pairedWindowStep) {
	t.Helper()
	spec, e := filepath.Abs("model/paired_window_309.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: paired_window_309\nconstants: {MAX_READS: '10'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range pairedWindowActions {
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
	if !checkPairedWindow(t, checker, steps) {
		t.Fatal("authored cohort journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]pairedWindowStep(nil), steps...)
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
		if checkPairedWindow(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkPairedWindow(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintPairedWindowModelSmoke(t *testing.T) {
	raw := `[{"action":"narrow","reads":3,"pages":1,"compares":0,"saves":0,"preserved":true},{"action":"parent","reads":4,"pages":1,"compares":0,"saves":1,"preserved":true},{"action":"restCompared","reads":6,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"promote","reads":6,"pages":1,"compares":1,"saves":1,"preserved":true},{"action":"save","reads":6,"pages":1,"compares":1,"saves":2,"preserved":true},{"action":"originalRun","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"invalidDraft","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"discardDraft","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"stageWindow","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"stale","reads":8,"pages":1,"compares":2,"saves":2,"preserved":true},{"action":"saveWindow","reads":8,"pages":1,"compares":2,"saves":3,"preserved":true},{"action":"reopen","reads":8,"pages":1,"compares":2,"saves":3,"preserved":true},{"action":"newRun","reads":10,"pages":1,"compares":3,"saves":3,"preserved":true},{"action":"encodeArtifact","reads":10,"pages":1,"compares":3,"saves":3,"preserved":true}]`
	var steps []pairedWindowStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validatePairedWindow(t, steps)
}
func TestQuintPairedWindowObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_PAIRED_WINDOW_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []pairedWindowStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validatePairedWindow(t, steps)
}
