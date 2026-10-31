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

type viewReuseStep struct {
	Action    string `json:"action"`
	Reads     int    `json:"reads"`
	Reused    bool   `json:"reused"`
	Preserved bool   `json:"preserved"`
}

var reuseActions = []string{"population", "comparison", "inspect", "apply", "save", "reopen", "rerun"}

func checkViewReuse(t *testing.T, c *validate.Checker, steps []viewReuseStep) bool {
	t.Helper()
	if len(steps) != len(reuseActions) {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for i, s := range steps {
		if s.Action != reuseActions[i] || !s.Preserved {
			return false
		}
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.reused", s.Reused)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func validateViewReuse(t *testing.T, steps []viewReuseStep) {
	t.Helper()
	spec, e := filepath.Abs("model/view_reuse_277.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: view_reuse_277\nconstants: {MAX_READS: '6'}\ninvariants: [safety]\nactions:\n", spec)
	for _, a := range reuseActions {
		binding += fmt.Sprintf("  %s:\n    expect: 'reads == {{.obs.reads}} and reused == {{.obs.reused}}'\n", a)
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: filepath.Dir(file)})
	if e != nil {
		t.Fatal(e)
	}
	if !checkViewReuse(t, c, steps) {
		t.Fatal("reuse journey refused")
	}
	for i := 0; i < 5; i++ {
		bad := append([]viewReuseStep(nil), steps...)
		switch i {
		case 0:
			bad[2].Reads++
		case 1:
			bad[3].Reused = false
		case 2:
			bad[4].Preserved = false
		case 3:
			bad[5].Reads++
		case 4:
			bad[6].Reads--
		}
		if checkViewReuse(t, c, bad) {
			t.Fatal("mutant accepted", i)
		}
	}
	if checkViewReuse(t, c, steps[:6]) {
		t.Fatal("unfinished prefix accepted")
	}
}
func TestQuintViewReuseModelSmoke(t *testing.T) {
	steps := []viewReuseStep{}
	for i, a := range reuseActions {
		reads := 4
		if i == 0 {
			reads = 3
		}
		if i == 6 {
			reads = 6
		}
		steps = append(steps, viewReuseStep{Action: a, Reads: reads, Reused: i >= 3, Preserved: true})
	}
	validateViewReuse(t, steps)
}
func TestQuintViewReuseObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_VIEW_REUSE_LEDGER")
	if file == "" {
		t.Skip("real browser ledger not supplied")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var steps []viewReuseStep
	if e = json.Unmarshal(raw, &steps); e != nil {
		t.Fatal(e)
	}
	validateViewReuse(t, steps)
}
