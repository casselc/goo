package constraints424

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type modelStep struct {
	Name                string
	Reads, SQL, Counted int
}

func checkModel(t *testing.T, steps []modelStep) bool {
	t.Helper()
	quint := os.Getenv("QUINT_BIN")
	if quint == "" {
		quint = "../../node_modules/.bin/quint"
	}
	quint, err := filepath.Abs(quint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(quint); err != nil {
		t.Fatalf("native Quint mandatory, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		recorder.Record(context.Background(), s.Name, nil, "obs.reads", s.Reads, "obs.sql", s.SQL, "obs.counted", s.Counted)
	}
	checker, err := validate.NewChecker("binding.yaml", validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestConstraintReadAndStageProtocol(t *testing.T) {
	good := []modelStep{{"inspect", 0, 0, 0}, {"context", 0, 0, 0}, {"edit", 0, 0, 0}, {"load", 1, 2, 1}, {"local", 1, 2, 1}, {"remove", 1, 2, 1}, {"local", 1, 2, 1}, {"apply", 2, 3, 1}}
	if !checkModel(t, good) {
		t.Fatal("valid contextual editing protocol refused")
	}
	for _, at := range []int{1, 2, 4, 5, 6} {
		bad := append([]modelStep(nil), good...)
		bad[at].Reads++
		if checkModel(t, bad) {
			t.Fatal("local gesture gained acquisition", at)
		}
	}
	for _, bad := range [][]modelStep{
		{{"apply", 1, 1, 0}},
		{{"inspect", 0, 0, 0}, {"context", 0, 0, 0}, {"choose_or", 0, 0, 0}, {"remove", 0, 0, 0}},
		{{"inspect", 0, 0, 0}, {"context", 0, 0, 0}, {"retire", 0, 0, 0}, {"remove", 0, 0, 0}},
		{{"inspect", 0, 0, 0}, {"context", 0, 0, 0}, {"remove", 0, 0, 0}, {"load", 1, 2, 1}},
	} {
		if checkModel(t, bad) {
			t.Fatal("unsafe, stale or out-of-order action admitted", bad)
		}
	}
}

type counters struct{ Source, SQL, Counted int }
type fact struct {
	Name          string
	Before, After counters
}
type journeyReport struct {
	Version, Status string
	Facts           []fact
	Errors          []string
}

func TestObservedConstraintLedger(t *testing.T) {
	path := os.Getenv("GOO_CONSTRAINTS424_REPORT")
	if path == "" {
		t.Skip("actual completed browser ledger supplied by GOO_CONSTRAINTS424_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report journeyReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.fieldConstraints424/1" || report.Status != "passed" || len(report.Errors) > 0 {
		t.Fatal("incomplete real contextual constraint report")
	}
	actions := map[string]string{"inspect": "inspect", "open_constraints": "context", "edit_exact": "edit", "load_values": "load", "guard_actions": "local", "local_frames": "local", "remove_condition": "remove", "stale_actions": "local", "apply_broader": "apply"}
	seen := map[string]bool{}
	steps := []modelStep{}
	state := modelStep{}
	for i, f := range report.Facts {
		if i > 0 && f.Before != report.Facts[i-1].After {
			t.Fatal("unrecorded effect", f.Name)
		}
		if f.After.Source < f.Before.Source || f.After.SQL < f.Before.SQL || f.After.Counted < f.Before.Counted {
			t.Fatal("regressed counters", f.Name)
		}
		name, ok := actions[f.Name]
		if !ok {
			t.Fatal("unmapped fact", f.Name)
		}
		if seen[f.Name] {
			t.Fatal("duplicate fact", f.Name)
		}
		seen[f.Name] = true
		state.Name = name
		state.Reads += f.After.Source - f.Before.Source
		state.SQL += f.After.SQL - f.Before.SQL
		state.Counted += f.After.Counted - f.Before.Counted
		steps = append(steps, state)
	}
	for name := range actions {
		if !seen[name] {
			t.Fatal("required journey path absent", name)
		}
	}
	if !checkModel(t, steps) {
		t.Fatal("actual contextual constraint source protocol failed")
	}
	bad := append([]modelStep(nil), steps...)
	bad[len(bad)-2].Reads++
	if checkModel(t, bad) {
		t.Fatal("mutated stale action gained acquisition")
	}
}
