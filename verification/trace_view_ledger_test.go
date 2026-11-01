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

type traceViewStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Keeps     int    `json:"keeps"`
	Preserved bool   `json:"preserved"`
}

var traceViewActions = []string{"parent", "inspectTrace", "inspectorPage", "stage", "read", "list", "keep", "page", "save", "stale", "back", "reopen", "rerun"}

func checkTraceView(t *testing.T, c *validate.Checker, steps []traceViewStep) bool {
	t.Helper()
	if len(steps) != len(traceViewActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != traceViewActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.keeps", s.Keeps)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateTraceView(t *testing.T, steps []traceViewStep) {
	t.Helper()
	spec, e := filepath.Abs("model/trace_view_303.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: trace_view_303\nconstants: {MAX_READS: '7'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range traceViewActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and keeps == {{.obs.keeps}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkTraceView(t, checker, steps) {
		t.Fatal("trace handoff journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]traceViewStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Reads++
		case 1:
			bad[6].Keeps--
		case 2:
			bad[5].Reads++
		case 3:
			bad[10].Pages++
		case 4:
			bad[10].Reads++
		case 5:
			bad[7].Pages--
		case 6:
			bad[8].Preserved = false
		}
		if checkTraceView(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkTraceView(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintTraceViewModelSmoke(t *testing.T) {
	raw := `[{"action":"parent","reads":3,"pages":0,"saves":0,"keeps":0,"preserved":true},{"action":"inspectTrace","reads":5,"pages":0,"saves":0,"keeps":0,"preserved":true},{"action":"inspectorPage","reads":7,"pages":0,"saves":0,"keeps":0,"preserved":true},{"action":"stage","reads":7,"pages":0,"saves":0,"keeps":0,"preserved":true},{"action":"read","reads":7,"pages":1,"saves":0,"keeps":0,"preserved":true},{"action":"list","reads":7,"pages":1,"saves":0,"keeps":0,"preserved":true},{"action":"keep","reads":7,"pages":1,"saves":0,"keeps":1,"preserved":true},{"action":"page","reads":7,"pages":2,"saves":0,"keeps":1,"preserved":true},{"action":"save","reads":7,"pages":2,"saves":1,"keeps":1,"preserved":true},{"action":"stale","reads":7,"pages":2,"saves":1,"keeps":1,"preserved":true},{"action":"back","reads":7,"pages":2,"saves":1,"keeps":1,"preserved":true},{"action":"reopen","reads":7,"pages":2,"saves":1,"keeps":1,"preserved":true},{"action":"rerun","reads":7,"pages":3,"saves":1,"keeps":1,"preserved":true}]`
	var steps []traceViewStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateTraceView(t, steps)
}
func TestQuintTraceViewObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_TRACE_VIEW_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []traceViewStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateTraceView(t, steps)
}
