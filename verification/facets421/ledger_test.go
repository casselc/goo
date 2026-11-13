package facets421

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
func TestFacetEntrySeparatesLocalIntentFromExplicitRead(t *testing.T) {
	good := []modelStep{{"inspect", 0, 0, 0}, {"local", 0, 0, 0}, {"facet", 0, 0, 0}, {"load", 1, 2, 1}, {"local", 1, 2, 1}, {"facet", 1, 2, 1}, {"load", 2, 4, 2}, {"stage", 2, 4, 2}, {"local", 2, 4, 2}}
	if !checkModel(t, good) {
		t.Fatal("valid facet source/local protocol refused")
	}
	for _, at := range []int{0, 1, 2, 4, 7, 8} {
		bad := append([]modelStep(nil), good...)
		bad[at].Reads++
		bad[at].SQL += 2
		bad[at].Counted++
		if checkModel(t, bad) {
			t.Fatal("local gesture acquired source", at)
		}
	}
	if checkModel(t, []modelStep{{"load", 1, 2, 1}}) {
		t.Fatal("load before field selection admitted")
	}
	if !checkModel(t, []modelStep{{"inspect", 0, 0, 0}, {"facet", 0, 0, 0}, {"begin_load", 1, 2, 1}, {"finish_load", 1, 2, 1}}) {
		t.Fatal("valid pending load completion refused")
	}
	if checkModel(t, []modelStep{{"inspect", 0, 0, 0}, {"facet", 0, 0, 0}, {"begin_load", 1, 2, 1}, {"retire", 1, 2, 1}, {"finish_load", 1, 2, 1}}) {
		t.Fatal("retired completion revived scope")
	}
	bad := append([]modelStep(nil), good...)
	bad[3].SQL++
	if checkModel(t, bad) {
		t.Fatal("excess SQL admitted")
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

func TestObservedActivityFacetLedger(t *testing.T) {
	path := os.Getenv("GOO_FACETS421_REPORT")
	if path == "" {
		t.Skip("actual completed browser ledger supplied by GOO_FACETS421_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report journeyReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.activityFacetEntry421/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Facts) < 11 {
		t.Fatal("incomplete actual contextual facet report")
	}
	actions := map[string]string{"inspect": "inspect", "guard_entry": "local", "facet_attribute": "facet", "local_frames": "local", "load_attribute": "load", "browse_counted": "local", "facet_resource": "facet", "load_resource": "load", "guard_load": "local", "stage_empty": "stage", "stale_entry": "local"}
	seen := map[string]bool{}
	steps := []modelStep{}
	state := modelStep{}
	for i, f := range report.Facts {
		if i > 0 && f.Before != report.Facts[i-1].After {
			t.Fatalf("unrecorded effect before %s", f.Name)
		}
		if f.After.Source < f.Before.Source || f.After.SQL < f.Before.SQL || f.After.Counted < f.Before.Counted {
			t.Fatal("regressed counters", f.Name)
		}
		name, ok := actions[f.Name]
		if !ok {
			t.Fatalf("unmapped fact %q", f.Name)
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
			t.Fatalf("required path absent: %s", name)
		}
	}
	if !checkModel(t, steps) {
		t.Fatal("actual scope/read protocol failed")
	}
	bad := append([]modelStep(nil), steps...)
	bad[len(bad)-1].Reads++
	bad[len(bad)-1].SQL += 2
	bad[len(bad)-1].Counted++
	if checkModel(t, bad) {
		t.Fatal("mutated actual local action gained source read")
	}
}
