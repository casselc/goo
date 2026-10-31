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

type timeTemplateStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Pages     int    `json:"pages"`
	Saves     int    `json:"saves"`
	Staged    bool   `json:"staged"`
	Preserved bool   `json:"preserved"`
}

var timeTemplateActions = []string{"service", "customize", "stageSource", "readSource", "author", "saveTemplate", "target", "groupTarget", "readTarget", "missing", "wrongStage", "wrongRead", "mismatch", "matchingStage", "matchingRead", "preview", "apply", "readApplied", "page", "saveDestination", "reopen", "rerun"}

func checkTimeTemplate(t *testing.T, c *validate.Checker, steps []timeTemplateStep) bool {
	t.Helper()
	if len(steps) != len(timeTemplateActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != timeTemplateActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.pages", s.Pages, "obs.saves", s.Saves, "obs.staged", s.Staged)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateTimeTemplate(t *testing.T, steps []timeTemplateStep) {
	t.Helper()
	spec, e := filepath.Abs("model/time_template_299.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: time_template_299\nconstants: {MAX_READS: '8'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range timeTemplateActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and pages == {{.obs.pages}} and saves == {{.obs.saves}} and staged == {{.obs.staged}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkTimeTemplate(t, checker, steps) {
		t.Fatal("authored template journey refused")
	}
	for i := 0; i < 7; i++ {
		bad := append([]timeTemplateStep(nil), steps...)
		switch i {
		case 0:
			bad[9].Reads++
		case 1:
			bad[12].Reads++
		case 2:
			bad[16].Saves++
		case 3:
			bad[18].Pages--
		case 4:
			bad[20].Reads++
		case 5:
			bad[17].Preserved = false
		case 6:
			bad[21].Reads--
		}
		if checkTimeTemplate(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkTimeTemplate(t, checker, steps[:len(steps)-1]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintTimeTemplateModelSmoke(t *testing.T) {
	reads, pages, saves := 0, 0, 0
	steps := []timeTemplateStep{}
	for _, a := range timeTemplateActions {
		switch a {
		case "service", "readSource", "target", "readTarget", "wrongRead", "matchingRead", "readApplied", "rerun":
			reads++
		}
		if a == "page" {
			pages++
		}
		if a == "saveTemplate" || a == "saveDestination" {
			saves++
		}
		staged := false
		switch a {
		case "customize", "stageSource", "target", "groupTarget", "wrongStage", "matchingStage", "reopen":
			staged = true
		}
		steps = append(steps, timeTemplateStep{a, reads, pages, saves, staged, true})
	}
	validateTimeTemplate(t, steps)
}
func TestQuintTimeTemplateObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_TIME_TEMPLATE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []timeTemplateStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateTimeTemplate(t, steps)
}
