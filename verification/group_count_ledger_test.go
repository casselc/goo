package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"testing"
)

type groupCountStep struct {
	Action    string `json:"action"`
	Full      int    `json:"full"`
	Pages     int    `json:"pages"`
	Baselines int    `json:"baselines"`
	Compares  int    `json:"compares"`
	Saves     int    `json:"saves"`
	Held      int    `json:"held"`
	Fresh     int    `json:"fresh"`
	Groups    int    `json:"groups"`
	Selected  int    `json:"selected"`
}

var groupCountActions = []string{"open", "readPopulation", "prepareBaseline", "chooseGroups", "inspectLocally", "clearSelection", "arrival", "pageRecords", "saveDefinition", "reopenDefinition", "readAgain", "prepareZero"}

func checkGroupCount(t *testing.T, c *validate.Checker, steps []groupCountStep) bool {
	t.Helper()
	if len(steps) != len(groupCountActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != groupCountActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.full", s.Full, "obs.pages", s.Pages, "obs.baselines", s.Baselines, "obs.compares", s.Compares, "obs.saves", s.Saves, "obs.groups", s.Groups, "obs.selected", s.Selected, "obs.held", s.Held, "obs.fresh", s.Fresh)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateGroupCount(t *testing.T, steps []groupCountStep) {
	t.Helper()
	spec, e := filepath.Abs("model/group_count_291.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: group_count_291\nconstants: {MAX_FULL: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range groupCountActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'full == {{.obs.full}} and pages == {{.obs.pages}} and baselines == {{.obs.baselines}} and compares == {{.obs.compares}} and saves == {{.obs.saves}} and groups == {{.obs.groups}} and selected == {{.obs.selected}} and held == {{.obs.held}} and fresh == {{.obs.fresh}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkGroupCount(t, c, steps) {
		t.Fatal("group-count comparison journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]groupCountStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Compares++
		case 1:
			bad[4].Selected--
		case 2:
			bad[7].Held++
		case 3:
			bad[9].Baselines++
		case 4:
			bad[8].Saves--
		case 5:
			bad[11].Groups--
		}
		if checkGroupCount(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkGroupCount(t, c, steps[:11]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintGroupCountModelSmoke(t *testing.T) {
	steps := []groupCountStep{}
	values := [][9]int{{0, 0, 0, 0, 0, 0, 0, 0, 0}, {1, 0, 0, 0, 0, 63, 63, 0, 0}, {1, 0, 1, 1, 0, 63, 63, 0, 0}, {1, 0, 1, 1, 0, 63, 63, 1, 0}, {1, 0, 1, 1, 0, 63, 63, 1, 1}, {1, 0, 1, 1, 0, 63, 63, 1, 0}, {1, 0, 1, 1, 0, 63, 63, 1, 0}, {1, 1, 1, 1, 0, 63, 64, 1, 0}, {1, 1, 1, 1, 1, 63, 64, 1, 0}, {1, 1, 1, 1, 1, 0, 0, 0, 0}, {2, 1, 2, 2, 1, 64, 64, 1, 0}, {2, 1, 3, 3, 1, 64, 64, 1, 0}}
	for i, a := range groupCountActions {
		v := values[i]
		steps = append(steps, groupCountStep{Action: a, Full: v[0], Pages: v[1], Baselines: v[2], Compares: v[3], Saves: v[4], Held: v[5], Fresh: v[6], Groups: v[7], Selected: v[8]})
	}
	validateGroupCount(t, steps)
}

func TestQuintGroupCountObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_GROUP_COUNT_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []groupCountStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateGroupCount(t, steps)
}
