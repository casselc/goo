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

type groupSideStep struct {
	Action    string `json:"action"`
	Full      int    `json:"full"`
	Pages     int    `json:"pages"`
	Baselines int    `json:"baselines"`
	Compares  int    `json:"compares"`
	Child     int    `json:"child"`
	Staged    int    `json:"staged"`
	Held      int    `json:"held"`
}

var groupSideActions = []string{"open", "readPopulation", "pageParent", "stageBaseline", "readBaseline", "returnBaseline", "rejectOld", "stageCurrent", "readCurrent", "returnCurrent", "stageMissing", "readMissing", "returnMissing", "stageEmpty", "readEmpty", "returnEmpty", "retire"}

func checkGroupSide(t *testing.T, c *validate.Checker, steps []groupSideStep) bool {
	t.Helper()
	if len(steps) != len(groupSideActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != groupSideActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.full", s.Full, "obs.pages", s.Pages, "obs.baselines", s.Baselines, "obs.compares", s.Compares, "obs.child", s.Child, "obs.staged", s.Staged, "obs.held", s.Held)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateGroupSide(t *testing.T, steps []groupSideStep) {
	t.Helper()
	spec, e := filepath.Abs("model/group_side_292.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: group_side_292\nconstants: {MAX_FULL: '5'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range groupSideActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'full == {{.obs.full}} and pages == {{.obs.pages}} and baselines == {{.obs.baselines}} and compares == {{.obs.compares}} and child == {{.obs.child}} and staged == {{.obs.staged}} and held == {{.obs.held}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkGroupSide(t, c, steps) {
		t.Fatal("group-side staging journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]groupSideStep(nil), steps...)
		switch i {
		case 0:
			bad[3].Full++
		case 1:
			bad[5].Baselines++
		case 2:
			bad[6].Child++
		case 3:
			bad[7].Staged--
		case 4:
			bad[9].Held++
		case 5:
			bad[16].Compares++
		}
		if checkGroupSide(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkGroupSide(t, c, steps[:16]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintGroupSideModelSmoke(t *testing.T) {
	steps := []groupSideStep{}
	values := [][7]int{{0, 0, 0, 0, 0, 1, 0}, {1, 0, 1, 1, 0, 0, 6}, {1, 1, 1, 1, 0, 0, 6}, {1, 1, 1, 1, 1, 1, 6}, {2, 1, 1, 1, 1, 0, 6}, {2, 1, 1, 1, 0, 0, 6}, {2, 1, 1, 1, 0, 0, 6}, {2, 1, 1, 1, 1, 1, 6}, {3, 1, 1, 1, 1, 0, 6}, {3, 1, 1, 1, 0, 0, 6}, {3, 1, 1, 1, 1, 1, 6}, {4, 1, 1, 1, 1, 0, 6}, {4, 1, 1, 1, 0, 0, 6}, {4, 1, 1, 1, 1, 1, 6}, {5, 1, 1, 1, 1, 0, 6}, {5, 1, 1, 1, 0, 0, 6}, {5, 1, 1, 1, 0, 1, 0}}
	for i, a := range groupSideActions {
		v := values[i]
		steps = append(steps, groupSideStep{Action: a, Full: v[0], Pages: v[1], Baselines: v[2], Compares: v[3], Child: v[4], Staged: v[5], Held: v[6]})
	}
	validateGroupSide(t, steps)
}

func TestQuintGroupSideObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_GROUP_SIDE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []groupSideStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateGroupSide(t, steps)
}
