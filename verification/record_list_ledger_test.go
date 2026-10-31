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

type recordListStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Saves     int    `json:"saves"`
	Pages     int    `json:"pages"`
	Kind      string `json:"kind"`
	Staged    bool   `json:"staged"`
	Preserved bool   `json:"preserved"`
}

var recordListActions = []string{"service", "customize", "readSource", "selectionRefusal", "list", "details", "saveTemplate", "target", "groupTarget", "readTarget", "preview", "applyList", "readApplied", "arrival", "page", "retiredControl", "tableLocal", "listLocal", "saveDestination", "reopen", "rerun"}

func checkRecordList(t *testing.T, c *validate.Checker, steps []recordListStep) bool {
	t.Helper()
	if len(steps) != len(recordListActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != recordListActions[i] || !s.Preserved || (s.Kind != "record-table" && s.Kind != "record-list") {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.saves", s.Saves, "obs.staged", s.Staged, "obs.pages", s.Pages, "obs.kind", s.Kind)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateRecordList(t *testing.T, steps []recordListStep) {
	t.Helper()
	spec, e := filepath.Abs("model/record_list_297.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: record_list_297\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range recordListActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and saves == {{.obs.saves}} and staged == {{.obs.staged}} and pages == {{.obs.pages}} and kind == {{.obs.kind}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkRecordList(t, checker, steps) {
		t.Fatal("authored template journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]recordListStep(nil), steps...)
		switch i {
		case 0:
			bad[4].Reads++
		case 1:
			bad[10].Reads++
		case 2:
			bad[14].Pages--
		case 3:
			bad[18].Saves--
		case 4:
			bad[15].Preserved = false
		case 5:
			bad[19].Kind = "record-table"
		}
		if checkRecordList(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkRecordList(t, checker, steps[:20]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintRecordListModelSmoke(t *testing.T) {
	reads, saves, pages := 0, 0, 0
	steps := []recordListStep{}
	for i, a := range recordListActions {
		if i == 0 || i == 2 || i == 7 || i == 9 || i == 12 || i == 20 {
			reads++
		}
		if i == 6 || i == 18 {
			saves++
		}
		if i == 14 {
			pages++
		}
		kind := "record-list"
		if i <= 3 || i == 7 || i == 8 || i == 9 || i == 10 || i == 16 {
			kind = "record-table"
		}
		steps = append(steps, recordListStep{Action: a, Reads: reads, Saves: saves, Pages: pages, Kind: kind, Staged: i == 1 || i == 7 || i == 8 || i == 19, Preserved: true})
	}
	validateRecordList(t, steps)
}
func TestQuintRecordListObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_RECORD_LIST_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []recordListStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordList(t, steps)
}
