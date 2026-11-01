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

type comparisonTemplateStep struct {
	Action       string `json:"action"`
	Reads        int    `json:"reads"`
	Pages        int    `json:"pages"`
	Saves        int    `json:"saves"`
	Captures     int    `json:"captures"`
	Compares     int    `json:"compares"`
	Expands      int    `json:"expands"`
	CaptureReads int    `json:"captureReads"`
	Preserved    bool   `json:"preserved"`
}

var comparisonTemplateActions = []string{"source", "readSource", "captureSource", "prepareSource", "saveTemplate", "target", "readTarget", "captureTarget", "refreshTarget", "missing", "previewWrong", "applyWrong", "rejectWrong", "retireOld", "previewTarget", "applyTarget", "prepareTarget", "page", "saveTarget", "reopen", "rerun"}

func checkComparisonTemplate(t *testing.T, c *validate.Checker, steps []comparisonTemplateStep) bool {
	t.Helper()
	if len(steps) != len(comparisonTemplateActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != comparisonTemplateActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.captures", s.Captures, "obs.compares", s.Compares, "obs.expands", s.Expands, "obs.captureReads", s.CaptureReads)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateComparisonTemplate(t *testing.T, steps []comparisonTemplateStep) {
	t.Helper()
	spec, e := filepath.Abs("model/comparison_template_300.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: comparison_template_300\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range comparisonTemplateActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and captures == {{.obs.captures}} and compares == {{.obs.compares}} and expands == {{.obs.expands}} and captureReads == {{.obs.captureReads}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkComparisonTemplate(t, checker, steps) {
		t.Fatal("authored template journey refused")
	}
	for i := 0; i < 8; i++ {
		bad := append([]comparisonTemplateStep(nil), steps...)
		switch i {
		case 0:
			bad[9].CaptureReads++
		case 1:
			bad[11].Reads++
		case 2:
			bad[12].Compares++
		case 3:
			bad[15].CaptureReads++
		case 4:
			bad[17].Pages--
		case 5:
			bad[19].Reads++
		case 6:
			bad[20].Saves++
		case 7:
			bad[14].Preserved = false
		}
		if checkComparisonTemplate(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}

	if checkComparisonTemplate(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintComparisonTemplateModelSmoke(t *testing.T) {
	raw := `[{"action":"source","reads":1,"pages":0,"saves":0,"captures":0,"compares":0,"expands":1,"captureReads":0,"preserved":true},{"action":"readSource","reads":2,"pages":0,"saves":0,"captures":0,"compares":0,"expands":1,"captureReads":0,"preserved":true},{"action":"captureSource","reads":2,"pages":0,"saves":0,"captures":1,"compares":0,"expands":1,"captureReads":0,"preserved":true},{"action":"prepareSource","reads":2,"pages":0,"saves":0,"captures":1,"compares":1,"expands":1,"captureReads":1,"preserved":true},{"action":"saveTemplate","reads":2,"pages":0,"saves":1,"captures":1,"compares":1,"expands":1,"captureReads":1,"preserved":true},{"action":"target","reads":3,"pages":0,"saves":1,"captures":1,"compares":1,"expands":2,"captureReads":1,"preserved":true},{"action":"readTarget","reads":4,"pages":0,"saves":1,"captures":1,"compares":1,"expands":2,"captureReads":1,"preserved":true},{"action":"captureTarget","reads":4,"pages":0,"saves":1,"captures":2,"compares":1,"expands":2,"captureReads":1,"preserved":true},{"action":"refreshTarget","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":2,"captureReads":1,"preserved":true},{"action":"missing","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":2,"captureReads":1,"preserved":true},{"action":"previewWrong","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":3,"captureReads":1,"preserved":true},{"action":"applyWrong","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":3,"captureReads":1,"preserved":true},{"action":"rejectWrong","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":3,"captureReads":2,"preserved":true},{"action":"retireOld","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":4,"captureReads":2,"preserved":true},{"action":"previewTarget","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":5,"captureReads":2,"preserved":true},{"action":"applyTarget","reads":5,"pages":0,"saves":1,"captures":2,"compares":1,"expands":5,"captureReads":2,"preserved":true},{"action":"prepareTarget","reads":5,"pages":0,"saves":1,"captures":2,"compares":2,"expands":5,"captureReads":3,"preserved":true},{"action":"page","reads":5,"pages":1,"saves":1,"captures":2,"compares":2,"expands":5,"captureReads":3,"preserved":true},{"action":"saveTarget","reads":5,"pages":1,"saves":2,"captures":2,"compares":2,"expands":5,"captureReads":3,"preserved":true},{"action":"reopen","reads":5,"pages":1,"saves":2,"captures":2,"compares":2,"expands":5,"captureReads":3,"preserved":true},{"action":"rerun","reads":6,"pages":1,"saves":2,"captures":2,"compares":3,"expands":5,"captureReads":4,"preserved":true}]`
	var steps []comparisonTemplateStep
	if e := json.Unmarshal([]byte(raw), &steps); e != nil {
		t.Fatal(e)
	}
	validateComparisonTemplate(t, steps)
}
func TestQuintComparisonTemplateObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_COMPARISON_TEMPLATE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []comparisonTemplateStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateComparisonTemplate(t, steps)
}
