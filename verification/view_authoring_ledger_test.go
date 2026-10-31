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

type viewAuthoringStep struct {
	Action      string `json:"action"`
	Reads       int    `json:"reads"`
	Kind        string `json:"kind"`
	SaveBlocked bool   `json:"saveBlocked"`
	Title       string `json:"title"`
}

func checkViewAuthoring(t *testing.T, c *validate.Checker, steps []viewAuthoringStep) bool {
	t.Helper()
	names := []string{"staged", "read", "edit", "invalid", "apply", "save", "reopen", "rerun"}
	if len(steps) != len(names) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != names[i] {
			return false
		}
		if i >= 2 && s.Kind != "table" {
			return false
		}
		if i == 3 && !s.SaveBlocked {
			return false
		}
		if i >= 4 && s.Title != "Team request counts" {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateViewAuthoring(t *testing.T, steps []viewAuthoringStep) {
	t.Helper()
	spec, e := filepath.Abs("model/view_authoring_275.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: view_authoring_275\nconstants: {MAX_READS: '2'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range []string{"staged", "read", "edit", "invalid", "apply", "save", "reopen", "rerun"} {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkViewAuthoring(t, c, steps) {
		t.Fatal("authored view journey refused")
	}
	for mutant := 0; mutant < 6; mutant++ {
		bad := append([]viewAuthoringStep(nil), steps...)
		switch mutant {
		case 0:
			bad[2].Reads++
		case 1:
			bad[6].Reads++
		case 2:
			bad[3].SaveBlocked = false
		case 3:
			bad[4].Title = "Old title"
		case 4:
			bad[7].Kind = "bar"
		case 5:
			bad[7].Reads--
		}
		if checkViewAuthoring(t, c, bad) {
			t.Fatal("mutant passed", mutant)
		}
	}
	if checkViewAuthoring(t, c, steps[:7]) {
		t.Fatal("unfinished prefix passed")
	}
}
func viewAuthoringSmoke() []viewAuthoringStep {
	steps := []viewAuthoringStep{}
	for i, a := range []string{"staged", "read", "edit", "invalid", "apply", "save", "reopen", "rerun"} {
		s := viewAuthoringStep{Action: a, Kind: "table", Title: "Team request counts"}
		if i > 0 {
			s.Reads = 1
		}
		if i == 7 {
			s.Reads = 2
		}
		s.SaveBlocked = i == 3
		steps = append(steps, s)
	}
	return steps
}
func TestQuintViewAuthoringModelSmoke(t *testing.T) { validateViewAuthoring(t, viewAuthoringSmoke()) }
func TestQuintViewAuthoringObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_VIEW_AUTHORING_LEDGER")
	if file == "" {
		t.Skip("real observed browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []viewAuthoringStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateViewAuthoring(t, steps)
}
