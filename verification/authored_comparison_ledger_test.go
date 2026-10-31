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

type authoredComparisonStep struct {
	Action    string `json:"action"`
	Full      int    `json:"full"`
	Pages     int    `json:"pages"`
	Baselines int    `json:"baselines"`
	Compares  int    `json:"compares"`
	Saves     int    `json:"saves"`
	Held      int    `json:"held"`
	Fresh     int    `json:"fresh"`
	Pair      int    `json:"pair"`
}

var authoredComparisonActions = []string{"open", "readPopulation", "prepareBaseline", "arrival", "pageRecords", "saveDefinition", "reopenDefinition", "readAgain", "prepareZero", "refuseIncompatible", "refuseMissing", "retirePreparation"}

func checkAuthoredComparison(t *testing.T, c *validate.Checker, steps []authoredComparisonStep) bool {
	t.Helper()
	if len(steps) != len(authoredComparisonActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != authoredComparisonActions[i] {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.full", s.Full, "obs.pages", s.Pages, "obs.baselines", s.Baselines, "obs.compares", s.Compares, "obs.saves", s.Saves, "obs.pair", s.Pair, "obs.held", s.Held, "obs.fresh", s.Fresh)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateAuthoredComparison(t *testing.T, steps []authoredComparisonStep) {
	t.Helper()
	spec, e := filepath.Abs("model/authored_comparison_290.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: authored_comparison_290\nconstants: {MAX_FULL: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range authoredComparisonActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'full == {{.obs.full}} and pages == {{.obs.pages}} and baselines == {{.obs.baselines}} and compares == {{.obs.compares}} and saves == {{.obs.saves}} and pair == {{.obs.pair}} and held == {{.obs.held}} and fresh == {{.obs.fresh}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkAuthoredComparison(t, c, steps) {
		t.Fatal("authored comparison journey refused")
	}
	for i := 0; i < 6; i++ {
		bad := append([]authoredComparisonStep(nil), steps...)
		switch i {
		case 0:
			bad[1].Full++
		case 1:
			bad[4].Pages--
		case 2:
			bad[4].Held++
		case 3:
			bad[6].Full++
		case 4:
			bad[9].Pair++
		case 5:
			bad[11].Compares++
		}
		if checkAuthoredComparison(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkAuthoredComparison(t, c, steps[:11]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintAuthoredComparisonModelSmoke(t *testing.T) {
	steps := []authoredComparisonStep{}
	values := [][8]int{{0, 0, 0, 0, 0, 0, 0, 0}, {1, 0, 0, 0, 0, 3, 3, 0}, {1, 0, 1, 1, 0, 3, 3, 1}, {1, 0, 1, 1, 0, 3, 3, 1}, {1, 1, 1, 1, 0, 3, 4, 1}, {1, 1, 1, 1, 1, 3, 4, 1}, {1, 1, 1, 1, 1, 0, 0, 0}, {2, 1, 2, 2, 1, 4, 4, 1}, {2, 1, 3, 3, 1, 4, 4, 1}, {2, 1, 4, 3, 1, 4, 4, 0}, {2, 1, 5, 3, 1, 4, 4, 0}, {2, 1, 6, 3, 1, 0, 0, 0}}
	for i, a := range authoredComparisonActions {
		v := values[i]
		steps = append(steps, authoredComparisonStep{Action: a, Full: v[0], Pages: v[1], Baselines: v[2], Compares: v[3], Saves: v[4], Held: v[5], Fresh: v[6], Pair: v[7]})
	}
	validateAuthoredComparison(t, steps)
}

func TestQuintAuthoredComparisonObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_AUTHORED_COMPARISON_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []authoredComparisonStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateAuthoredComparison(t, steps)
}
