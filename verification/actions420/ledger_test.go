package actions420

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
	Name              string
	Reads, SQL, Pairs int
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
		t.Fatalf("native Quint required, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		recorder.Record(context.Background(), s.Name, nil, "obs.reads", s.Reads, "obs.sql", s.SQL, "obs.pairs", s.Pairs)
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

func TestExplicitActionAndPendingCompletionProtocol(t *testing.T) {
	good := []modelStep{{"local", 0, 0, 0}, {"stage_compare", 0, 0, 0}, {"run_compare", 2, 6, 1}, {"return_service", 2, 6, 1}, {"stage_records", 2, 6, 1}, {"read_records", 3, 10, 1}, {"inspect_record", 3, 10, 1}, {"read_trace", 5, 17, 1}, {"return_trace", 5, 17, 1}, {"row_trace", 7, 24, 1}, {"return_row", 7, 24, 1}}
	if !checkModel(t, good) {
		t.Fatal("valid explicit source and local-return protocol refused")
	}
	for _, at := range []int{0, 3, 6, 8, 10} {
		bad := append([]modelStep(nil), good...)
		bad[at].Reads++
		bad[at].SQL++
		if checkModel(t, bad) {
			t.Fatal("local gesture acquired source telemetry")
		}
	}
	if !checkModel(t, []modelStep{{"begin_customize", 0, 0, 0}, {"complete_customize", 0, 0, 0}}) {
		t.Fatal("current pending configuration completion refused")
	}
	if checkModel(t, []modelStep{{"begin_customize", 0, 0, 0}, {"retire_customize", 0, 0, 0}, {"complete_customize", 0, 0, 0}}) {
		t.Fatal("retired pending completion revived owner")
	}
	if checkModel(t, []modelStep{{"read_trace", 2, 7, 0}}) {
		t.Fatal("trace acquisition before inspected records accepted")
	}
}

type counters struct{ Source, Compare, SQL int }
type fact struct {
	Name          string
	Before, After counters
}
type journeyReport struct {
	Version, Status string
	Facts           []fact
	Errors          []string
}

func TestObservedContextualActionLedger(t *testing.T) {
	path := os.Getenv("GOO_ACTIONS420_REPORT")
	if path == "" {
		t.Skip("actual browser ledger supplied by GOO_ACTIONS420_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report journeyReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.contextualActions420/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Facts) < 15 {
		t.Fatal("incomplete actual contextual action report")
	}
	steps := []modelStep{}
	state := modelStep{}
	seen := map[string]bool{}
	actions := map[string]bool{"stage_compare": true, "run_compare": true, "return_service": true, "stage_records": true, "read_records": true, "inspect_record": true, "read_trace": true, "return_trace": true, "row_trace": true, "return_row": true}
	for i, f := range report.Facts {
		if i > 0 && f.Before != report.Facts[i-1].After {
			t.Fatalf("unrecorded effect before %s", f.Name)
		}
		if f.After.Source < f.Before.Source || f.After.SQL < f.Before.SQL || f.After.Compare < f.Before.Compare {
			t.Fatalf("regressed effect counters at %s", f.Name)
		}
		state.Name = "local"
		if actions[f.Name] {
			state.Name = f.Name
			seen[f.Name] = true
		} else if len(f.Name) < 6 || f.Name[:6] != "local_" {
			t.Fatalf("unmapped actual fact %q", f.Name)
		}
		state.Reads += f.After.Source - f.Before.Source
		state.SQL += f.After.SQL - f.Before.SQL
		state.Pairs += f.After.Compare - f.Before.Compare
		steps = append(steps, state)
	}
	for name := range actions {
		if !seen[name] {
			t.Fatalf("required path absent: %s", name)
		}
	}
	if !checkModel(t, steps) {
		t.Fatal("actual action/source/SQL budgets violate finite protocol")
	}
	bad := append([]modelStep(nil), steps...)
	bad[len(bad)-1].Reads++
	bad[len(bad)-1].SQL++
	if checkModel(t, bad) {
		t.Fatal("mutated actual retained local action gained a source read")
	}
}
