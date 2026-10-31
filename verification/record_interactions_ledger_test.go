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

type recordInteractionStep struct {
	Action   string `json:"action"`
	Reads    int    `json:"reads"`
	Pages    int    `json:"pages"`
	Keeps    int    `json:"keeps"`
	Saves    int    `json:"saves"`
	Kind     string `json:"kind"`
	Staged   bool   `json:"staged"`
	Selected int    `json:"selected"`
}

var recordInteractionActions = []string{"service", "customize", "read", "list", "inspect", "trace", "selected", "keep", "facet", "back", "retired", "page", "pageKeep", "table", "listAgain", "save"}

func checkRecordInteractions(t *testing.T, c *validate.Checker, steps []recordInteractionStep) bool {
	t.Helper()
	if len(steps) != len(recordInteractionActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != recordInteractionActions[i] || (s.Kind != "record-table" && s.Kind != "record-list") {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.keeps", s.Keeps, "obs.saves", s.Saves, "obs.kind", s.Kind, "obs.staged", s.Staged, "obs.selected", s.Selected)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateRecordInteractions(t *testing.T, steps []recordInteractionStep) {
	t.Helper()
	spec, e := filepath.Abs("model/record_interactions_298.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: record_interactions_298\nconstants: {MAX_READS: '4'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range recordInteractionActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and keeps == {{.obs.keeps}} and saves == {{.obs.saves}} and kind == {{.obs.kind}} and staged == {{.obs.staged}} and selectionCount == {{.obs.selected}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkRecordInteractions(t, checker, steps) {
		t.Fatal("shared record journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]recordInteractionStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Reads++
		case 1:
			bad[10].Reads++
		case 2:
			bad[11].Pages--
		case 3:
			bad[12].Keeps--
		case 4:
			bad[6].Selected--
		case 5:
			bad[13].Kind = "record-list"
		case 6:
			bad[15].Saves--
		}
		if checkRecordInteractions(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkRecordInteractions(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintRecordInteractionsModelSmoke(t *testing.T) {
	raw := `[{"action":"service","reads":1,"pages":0,"keeps":0,"saves":0,"kind":"record-table","staged":false,"selected":0},{"action":"customize","reads":1,"pages":0,"keeps":0,"saves":0,"kind":"record-table","staged":true,"selected":0},{"action":"read","reads":2,"pages":0,"keeps":0,"saves":0,"kind":"record-table","staged":false,"selected":0},{"action":"list","reads":2,"pages":0,"keeps":0,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"inspect","reads":2,"pages":0,"keeps":0,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"trace","reads":4,"pages":0,"keeps":0,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"selected","reads":4,"pages":0,"keeps":0,"saves":0,"kind":"record-list","staged":false,"selected":2},{"action":"keep","reads":4,"pages":0,"keeps":1,"saves":0,"kind":"record-list","staged":false,"selected":2},{"action":"facet","reads":4,"pages":0,"keeps":1,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"back","reads":4,"pages":0,"keeps":1,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"retired","reads":4,"pages":0,"keeps":1,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"page","reads":4,"pages":1,"keeps":1,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"pageKeep","reads":4,"pages":1,"keeps":2,"saves":0,"kind":"record-list","staged":false,"selected":1},{"action":"table","reads":4,"pages":1,"keeps":2,"saves":0,"kind":"record-table","staged":false,"selected":0},{"action":"listAgain","reads":4,"pages":1,"keeps":2,"saves":0,"kind":"record-list","staged":false,"selected":0},{"action":"save","reads":4,"pages":1,"keeps":2,"saves":1,"kind":"record-list","staged":false,"selected":0}]`
	var steps []recordInteractionStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordInteractions(t, steps)
}
func TestQuintRecordInteractionsObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_RECORD_INTERACTIONS_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []recordInteractionStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateRecordInteractions(t, steps)
}
