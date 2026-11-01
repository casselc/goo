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

type suppliedWaterfallStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Expands   int    `json:"expands"`
	Keeps     int    `json:"keeps"`
	Preserved bool   `json:"preserved"`
}

var suppliedWaterfallActions = []string{"source", "configure", "read", "waterfall", "keep", "inspect", "trace", "saveTemplate", "target", "readTarget", "apply", "saveTarget", "reopen", "rerun", "page"}

func checkSuppliedWaterfall(t *testing.T, c *validate.Checker, steps []suppliedWaterfallStep) bool {
	t.Helper()
	if len(steps) != len(suppliedWaterfallActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != suppliedWaterfallActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.expands", s.Expands, "obs.keeps", s.Keeps)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateSuppliedWaterfall(t *testing.T, steps []suppliedWaterfallStep) {
	t.Helper()
	spec, e := filepath.Abs("model/supplied_waterfall_301.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: supplied_waterfall_301\nconstants: {MAX_READS: '8'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range suppliedWaterfallActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and expands == {{.obs.expands}} and keeps == {{.obs.keeps}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkSuppliedWaterfall(t, checker, steps) {
		t.Fatal("authored template journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]suppliedWaterfallStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Reads++
		case 1:
			bad[4].Keeps--
		case 2:
			bad[5].Reads++
		case 3:
			bad[10].Expands--
		case 4:
			bad[12].Reads++
		case 5:
			bad[14].Pages--
		case 6:
			bad[8].Preserved = false
		}
		if checkSuppliedWaterfall(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkSuppliedWaterfall(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintSuppliedWaterfallModelSmoke(t *testing.T) {
	raw := `[{"action":"source","reads":1,"pages":0,"saves":0,"expands":1,"keeps":0,"preserved":true},{"action":"configure","reads":1,"pages":0,"saves":0,"expands":1,"keeps":0,"preserved":true},{"action":"read","reads":2,"pages":0,"saves":0,"expands":1,"keeps":0,"preserved":true},{"action":"waterfall","reads":2,"pages":0,"saves":0,"expands":1,"keeps":0,"preserved":true},{"action":"keep","reads":2,"pages":0,"saves":0,"expands":1,"keeps":1,"preserved":true},{"action":"inspect","reads":2,"pages":0,"saves":0,"expands":1,"keeps":1,"preserved":true},{"action":"trace","reads":4,"pages":0,"saves":0,"expands":1,"keeps":1,"preserved":true},{"action":"saveTemplate","reads":4,"pages":0,"saves":1,"expands":1,"keeps":1,"preserved":true},{"action":"target","reads":5,"pages":0,"saves":1,"expands":2,"keeps":1,"preserved":true},{"action":"readTarget","reads":6,"pages":0,"saves":1,"expands":2,"keeps":1,"preserved":true},{"action":"apply","reads":6,"pages":0,"saves":1,"expands":3,"keeps":1,"preserved":true},{"action":"saveTarget","reads":6,"pages":0,"saves":2,"expands":3,"keeps":1,"preserved":true},{"action":"reopen","reads":6,"pages":0,"saves":2,"expands":3,"keeps":1,"preserved":true},{"action":"rerun","reads":7,"pages":0,"saves":2,"expands":3,"keeps":1,"preserved":true},{"action":"page","reads":8,"pages":1,"saves":2,"expands":3,"keeps":1,"preserved":true}]`
	var steps []suppliedWaterfallStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateSuppliedWaterfall(t, steps)
}
func TestQuintSuppliedWaterfallObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_SUPPLIED_WATERFALL_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []suppliedWaterfallStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateSuppliedWaterfall(t, steps)
}
