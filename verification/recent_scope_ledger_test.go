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

type recentScopeStep struct {
	Action string `json:"action"`
	Reads  int    `json:"reads"`
	Staged bool   `json:"staged"`
	Plans  int    `json:"plans"`
}

var recentScopeActions = []string{"open", "recent", "preview", "link", "child", "execute", "save", "reopen", "rerun"}

func checkRecentScope(t *testing.T, c *validate.Checker, steps []recentScopeStep) bool {
	t.Helper()
	if len(steps) != len(recentScopeActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != recentScopeActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.staged", s.Staged, "obs.plans", s.Plans)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateRecentScope(t *testing.T, steps []recentScopeStep) {
	t.Helper()
	spec, e := filepath.Abs("model/recent_scope_285.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: recent_scope_285\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range recentScopeActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and staged == {{.obs.staged}} and plans == {{.obs.plans}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkRecentScope(t, c, steps) {
		t.Fatal("Recent scope handoff refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]recentScopeStep(nil), steps...)
		switch i {
		case 0:
			bad[0].Reads++
		case 1:
			bad[1].Plans--
		case 2:
			bad[4].Staged = false
		case 3:
			bad[3].Reads++
		case 4:
			bad[8].Reads--
		case 5:
			bad[3].Plans++
		}
		if checkRecentScope(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkRecentScope(t, c, steps[:8]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintRecentScopeModelSmoke(t *testing.T) {
	reads := []int{0, 2, 2, 2, 2, 4, 4, 4, 6}
	staged := []bool{true, false, false, false, true, false, false, true, false}
	steps := []recentScopeStep{}
	for i, a := range recentScopeActions {
		plans := 1
		if i == 0 {
			plans = 0
		}
		steps = append(steps, recentScopeStep{Action: a, Reads: reads[i], Staged: staged[i], Plans: plans})
	}
	validateRecentScope(t, steps)
}

func TestQuintRecentScopeObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_RECENT_SCOPE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []recentScopeStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateRecentScope(t, steps)
}
