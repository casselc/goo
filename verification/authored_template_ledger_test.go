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

type authoredTemplateStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Saves     int    `json:"saves"`
	Staged    bool   `json:"staged"`
	Preserved bool   `json:"preserved"`
}

var authoredTemplateActions = []string{"service", "customize", "readSource", "saveRegular", "disabledProfile", "author", "saveTemplate", "target", "groupTarget", "readTarget", "preview", "apply", "readApplied", "saveDestination", "reopen", "rerun"}

func checkAuthoredTemplate(t *testing.T, c *validate.Checker, steps []authoredTemplateStep) bool {
	t.Helper()
	if len(steps) != len(authoredTemplateActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != authoredTemplateActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.saves", s.Saves, "obs.staged", s.Staged)
	}
	result, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func validateAuthoredTemplate(t *testing.T, steps []authoredTemplateStep) {
	t.Helper()
	spec, e := filepath.Abs("model/authored_template_296.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: authored_template_296\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range authoredTemplateActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and saves == {{.obs.saves}} and staged == {{.obs.staged}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkAuthoredTemplate(t, checker, steps) {
		t.Fatal("authored template journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]authoredTemplateStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Reads++
		case 1:
			bad[10].Reads++
		case 2:
			bad[11].Saves++
		case 3:
			bad[14].Reads++
		case 4:
			bad[12].Preserved = false
		case 5:
			bad[15].Reads--
		}
		if checkAuthoredTemplate(t, checker, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkAuthoredTemplate(t, checker, steps[:15]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintAuthoredTemplateModelSmoke(t *testing.T) {
	reads, saves := 0, 0
	steps := []authoredTemplateStep{}
	for i, a := range authoredTemplateActions {
		if i == 0 || i == 2 || i == 7 || i == 9 || i == 12 || i == 15 {
			reads++
		}
		if i == 3 || i == 6 || i == 13 {
			saves++
		}
		staged := i == 1 || i == 7 || i == 8 || i == 14
		steps = append(steps, authoredTemplateStep{a, reads, saves, staged, true})
	}
	validateAuthoredTemplate(t, steps)
}
func TestQuintAuthoredTemplateObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_AUTHORED_TEMPLATE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []authoredTemplateStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateAuthoredTemplate(t, steps)
}
