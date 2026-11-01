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

type addBreakdownStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Captures  int    `json:"captures"`
	Compares  int    `json:"compares"`
	Preserved bool   `json:"preserved"`
}

var addBreakdownActions = []string{"raw", "rawRead", "discard", "add", "grouped", "save", "capture", "lateRead", "compare", "reopen", "rerun", "targetRaw", "targetGrouped", "apply", "saveTarget"}

func checkAddBreakdown(t *testing.T, c *validate.Checker, steps []addBreakdownStep) bool {
	t.Helper()
	if len(steps) != len(addBreakdownActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != addBreakdownActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.captures", s.Captures, "obs.compares", s.Compares)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateAddBreakdown(t *testing.T, steps []addBreakdownStep) {
	t.Helper()
	spec, e := filepath.Abs("model/add_breakdown_304.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: add_breakdown_304\nconstants: {MAX_READS: '10'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range addBreakdownActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and captures == {{.obs.captures}} and compares == {{.obs.compares}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkAddBreakdown(t, checker, steps) {
		t.Fatal("add breakdown journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]addBreakdownStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Reads++
		case 1:
			bad[6].Captures--
		case 2:
			bad[5].Reads++
		case 3:
			bad[10].Pages++
		case 4:
			bad[10].Reads++
		case 5:
			bad[8].Compares--
		case 6:
			bad[8].Preserved = false
		}
		if checkAddBreakdown(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkAddBreakdown(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintAddBreakdownModelSmoke(t *testing.T) {
	raw := `[{"action":"raw","reads":3,"pages":0,"saves":0,"captures":0,"compares":0,"preserved":true},{"action":"rawRead","reads":3,"pages":1,"saves":0,"captures":0,"compares":0,"preserved":true},{"action":"discard","reads":3,"pages":1,"saves":0,"captures":0,"compares":0,"preserved":true},{"action":"add","reads":3,"pages":1,"saves":0,"captures":0,"compares":0,"preserved":true},{"action":"grouped","reads":4,"pages":1,"saves":0,"captures":0,"compares":0,"preserved":true},{"action":"save","reads":4,"pages":1,"saves":1,"captures":0,"compares":0,"preserved":true},{"action":"capture","reads":4,"pages":1,"saves":1,"captures":1,"compares":0,"preserved":true},{"action":"lateRead","reads":5,"pages":1,"saves":1,"captures":1,"compares":0,"preserved":true},{"action":"compare","reads":5,"pages":1,"saves":1,"captures":1,"compares":1,"preserved":true},{"action":"reopen","reads":5,"pages":1,"saves":1,"captures":1,"compares":1,"preserved":true},{"action":"rerun","reads":6,"pages":1,"saves":1,"captures":1,"compares":1,"preserved":true},{"action":"targetRaw","reads":9,"pages":2,"saves":1,"captures":1,"compares":1,"preserved":true},{"action":"targetGrouped","reads":10,"pages":2,"saves":1,"captures":1,"compares":1,"preserved":true},{"action":"apply","reads":10,"pages":2,"saves":1,"captures":1,"compares":1,"preserved":true},{"action":"saveTarget","reads":10,"pages":2,"saves":2,"captures":1,"compares":1,"preserved":true}]`
	var steps []addBreakdownStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateAddBreakdown(t, steps)
}
func TestQuintAddBreakdownObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_ADD_BREAKDOWN_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []addBreakdownStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateAddBreakdown(t, steps)
}
