package capturedrates418

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type step struct {
	Action           string
	Source, Compares int
}

func conforms(t *testing.T, steps []step) bool {
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
	rec := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		rec.Record(context.Background(), s.Action, nil, "obs.source", s.Source, "obs.compares", s.Compares)
	}
	checker, err := validate.NewChecker("binding.yaml", validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	result, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}
func TestCapturedRateProtocolSeparatesAcquisitionFromLocalDisplay(t *testing.T) {
	good := []step{{"stage_view", 0, 0}, {"read_population", 1, 1}, {"local_display", 1, 1},
		{"choose_baseline", 1, 1}, {"prepare_baseline", 1, 2}, {"choose_baseline", 1, 2}, {"prepare_baseline", 1, 3},
		{"choose_baseline", 1, 3}, {"prepare_baseline", 1, 4}, {"choose_baseline", 1, 4}, {"prepare_baseline", 1, 5},
		{"diagnostic_compare", 1, 6}, {"stage_operand", 1, 6}, {"read_operand", 2, 6}}
	if !conforms(t, good) {
		t.Fatal("explicit source/comparison schedule rejected")
	}
	for _, i := range []int{0, 2, 3, 4, 11, 12} {
		bad := append([]step(nil), good...)
		bad[i].Source++
		if conforms(t, bad) {
			t.Fatalf("implicit acquisition at action%d accepted", i)
		}
	}
	bad := append([]step(nil), good...)
	bad[2].Compares++
	if conforms(t, bad) {
		t.Fatal("local display recomputed comparison")
	}
	if conforms(t, []step{{"read_population", 1, 1}}) {
		t.Fatal("un-staged Run accepted")
	}
}

type counters struct {
	Source  int `json:"source"`
	Compare int `json:"compare"`
	Publish int `json:"publish"`
}
type fact struct {
	Name   string   `json:"name"`
	Before counters `json:"before"`
	After  counters `json:"after"`
}

func TestObservedCapturedRateReadLedger(t *testing.T) {
	path := os.Getenv("GOO_CAPTURED_RATE_REPORT")
	if path == "" {
		t.Skip("actual completed browser ledger supplied by GOO_CAPTURED_RATE_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Version string   `json:"version"`
		Status  string   `json:"status"`
		Facts   []fact   `json:"facts"`
		Errors  []string `json:"errors"`
	}
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.capturedRates418/1" || report.Status != "passed" || len(report.Errors) != 0 {
		t.Fatal("completed real-source journey required")
	}
	names := map[string]string{"stage_view": "stage_view", "read_population": "read_population", "local_read_facts": "local_display", "choose_count_measure": "choose_baseline", "prepare_count_measure": "prepare_baseline", "choose_rate_measure": "choose_baseline", "prepare_rate_measure": "prepare_baseline", "choose_changed": "choose_baseline", "prepare_changed": "prepare_baseline", "choose_zero": "choose_baseline", "prepare_zero": "prepare_baseline", "controlled_diagnostic": "diagnostic_compare", "stage_baseline_question": "stage_operand", "read_baseline_records": "read_operand"}
	if len(report.Facts) != len(names) {
		t.Fatalf("expected complete14-fact read schedule, got%d", len(report.Facts))
	}
	steps := make([]step, 0, len(report.Facts))
	previous := counters{}
	seen := map[string]bool{}
	for _, f := range report.Facts {
		action, ok := names[f.Name]
		if !ok || seen[f.Name] {
			t.Fatalf("unknown/repeated ledger fact %q", f.Name)
		}
		seen[f.Name] = true
		if f.Before.Source != previous.Source || f.Before.Compare != previous.Compare || f.Before.Publish != 0 || f.After.Publish != 0 {
			t.Fatalf("unrecorded source/compare or publication at %+v", f)
		}
		previous = f.After
		steps = append(steps, step{action, f.After.Source, f.After.Compare})
	}
	if previous.Source != 2 || previous.Compare != 6 || !conforms(t, steps) {
		t.Fatal("actual captured-rate schedule violates source authority")
	}
	bad := append([]step(nil), steps...)
	for i := range bad {
		if bad[i].Action == "prepare_baseline" {
			bad[i].Source++
			break
		}
	}
	if conforms(t, bad) {
		t.Fatal("corrupt actual Prepare acquisition accepted")
	}
}
