package countedcompare425

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
	Name                          string
	Reads, SQL, Counted, Compares int
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
		recorder.Record(context.Background(), s.Name, nil, "obs.reads", s.Reads, "obs.sql", s.SQL, "obs.counted", s.Counted, "obs.compares", s.Compares)
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
func TestComparisonReadAndStageProtocol(t *testing.T) {
	good := []modelStep{{"inspect", 0, 0, 0, 0}, {"facet", 0, 0, 0, 0}, {"load", 1, 2, 1, 0}, {"choose_values", 1, 2, 1, 0}, {"local", 1, 2, 1, 0}, {"stage", 1, 2, 1, 0}, {"local", 1, 2, 1, 0}, {"execute_comparison", 3, 8, 1, 1}, {"local", 3, 8, 1, 1}, {"back", 3, 8, 1, 1}, {"retire", 3, 8, 1, 1}}
	if !checkModel(t, good) {
		t.Fatal("valid selected/rest protocol refused")
	}
	for _, at := range []int{0, 1, 3, 4, 5, 6, 8, 9, 10} {
		bad := append([]modelStep(nil), good...)
		bad[at].Reads++
		if checkModel(t, bad) {
			t.Fatal("local gesture gained source acquisition", at)
		}
	}
	for _, bad := range [][]modelStep{
		{{"execute_comparison", 2, 6, 0, 1}},
		{{"inspect", 0, 0, 0, 0}, {"facet", 0, 0, 0, 0}, {"load", 1, 2, 1, 0}, {"stage", 1, 2, 1, 0}},
		{{"inspect", 0, 0, 0, 0}, {"facet", 0, 0, 0, 0}, {"load", 1, 2, 1, 0}, {"choose_values", 1, 2, 1, 0}, {"retire", 1, 2, 1, 0}, {"stage", 1, 2, 1, 0}},
	} {
		if checkModel(t, bad) {
			t.Fatal("unselected, retired or unordered action admitted", bad)
		}
	}
	bad := append([]modelStep(nil), good...)
	bad[7].Compares = 0
	if checkModel(t, bad) {
		t.Fatal("missing independent comparison admitted")
	}
}

type counters struct{ Source, SQL, Counted, Compare int }
type fact struct {
	Name          string
	Before, After counters
}
type journeyReport struct {
	Version, Status string
	Facts           []fact
	Errors          []string
}

func TestObservedCountedSelectionLedger(t *testing.T) {
	path := os.Getenv("GOO_COUNTED_COMPARE425_REPORT")
	if path == "" {
		t.Skip("actual completed browser ledger supplied by GOO_COUNTED_COMPARE425_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report journeyReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.countedSelection425/1" || report.Status != "passed" || len(report.Errors) > 0 {
		t.Fatal("incomplete real counted selection comparison report")
	}
	actions := map[string]string{"inspect": "inspect", "facet": "facet", "load_values": "load", "select_values": "choose_values", "local_frames": "local", "guard_compare": "local", "stage_comparison": "stage", "retired_compare": "local", "run_comparison": "execute_comparison", "comparison_frames": "local", "return_activity": "back", "retire_owner": "retire"}
	seen := map[string]bool{}
	steps := []modelStep{}
	state := modelStep{}
	for i, f := range report.Facts {
		if i > 0 && f.Before != report.Facts[i-1].After {
			t.Fatal("unrecorded effect", f.Name)
		}
		if f.After.Source < f.Before.Source || f.After.SQL < f.Before.SQL || f.After.Counted < f.Before.Counted || f.After.Compare < f.Before.Compare {
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
		state.Compares += f.After.Compare - f.Before.Compare
		steps = append(steps, state)
	}
	for name := range actions {
		if !seen[name] {
			t.Fatal("required journey path absent", name)
		}
	}
	if !checkModel(t, steps) {
		t.Fatal("actual counted selection comparison source protocol failed")
	}
	bad := append([]modelStep(nil), steps...)
	bad[len(bad)-2].Reads++
	if checkModel(t, bad) {
		t.Fatal("mutated stale action gained acquisition")
	}
}
