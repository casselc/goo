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

type recordOnlyStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Keeps     int    `json:"keeps"`
	Preserved bool   `json:"preserved"`
}

var recordOnlyActions = []string{"sourceFixed", "sourceCustomize", "sourceRead", "listInspect", "keepLog", "saveLog", "pageLog", "reopenLog", "rerunLog", "targetRead", "applyLog", "saveTarget", "spanRead", "waterfall", "keepSpan", "pageSpan", "saveSpan", "reopenSpan", "rerunSpan"}

func checkRecordOnly(t *testing.T, c *validate.Checker, steps []recordOnlyStep) bool {
	t.Helper()
	if len(steps) != len(recordOnlyActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != recordOnlyActions[i] || !s.Preserved {
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
func validateRecordOnly(t *testing.T, steps []recordOnlyStep) {
	t.Helper()
	spec, e := filepath.Abs("model/record_only_302.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: record_only_302\nconstants: {MAX_READS: '9'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range recordOnlyActions {
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
	if !checkRecordOnly(t, checker, steps) {
		t.Fatal("records-only journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]recordOnlyStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Reads++
		case 1:
			bad[4].Keeps--
		case 2:
			bad[5].Reads++
		case 3:
			bad[10].Pages++
		case 4:
			bad[12].Reads++
		case 5:
			bad[15].Pages--
		case 6:
			bad[8].Preserved = false
		}
		if checkRecordOnly(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkRecordOnly(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintRecordOnlyModelSmoke(t *testing.T) {
	raw := `[{"action":"sourceFixed","reads":1,"pages":0,"saves":0,"keeps":0,"preserved":true},{"action":"sourceCustomize","reads":3,"pages":0,"saves":0,"keeps":0,"preserved":true},{"action":"sourceRead","reads":3,"pages":1,"saves":0,"keeps":0,"preserved":true},{"action":"listInspect","reads":3,"pages":1,"saves":0,"keeps":0,"preserved":true},{"action":"keepLog","reads":3,"pages":1,"saves":0,"keeps":1,"preserved":true},{"action":"saveLog","reads":3,"pages":1,"saves":1,"keeps":1,"preserved":true},{"action":"pageLog","reads":3,"pages":2,"saves":1,"keeps":1,"preserved":true},{"action":"reopenLog","reads":3,"pages":2,"saves":1,"keeps":1,"preserved":true},{"action":"rerunLog","reads":3,"pages":3,"saves":1,"keeps":1,"preserved":true},{"action":"targetRead","reads":6,"pages":4,"saves":1,"keeps":1,"preserved":true},{"action":"applyLog","reads":6,"pages":4,"saves":1,"keeps":1,"preserved":true},{"action":"saveTarget","reads":6,"pages":4,"saves":2,"keeps":1,"preserved":true},{"action":"spanRead","reads":9,"pages":5,"saves":2,"keeps":1,"preserved":true},{"action":"waterfall","reads":9,"pages":5,"saves":2,"keeps":1,"preserved":true},{"action":"keepSpan","reads":9,"pages":5,"saves":2,"keeps":2,"preserved":true},{"action":"pageSpan","reads":9,"pages":6,"saves":2,"keeps":2,"preserved":true},{"action":"saveSpan","reads":9,"pages":6,"saves":3,"keeps":2,"preserved":true},{"action":"reopenSpan","reads":9,"pages":6,"saves":3,"keeps":2,"preserved":true},{"action":"rerunSpan","reads":9,"pages":7,"saves":3,"keeps":2,"preserved":true}]`
	var steps []recordOnlyStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordOnly(t, steps)
}
func TestQuintRecordOnlyObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_RECORD_ONLY_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []recordOnlyStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordOnly(t, steps)
}
