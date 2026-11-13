package recordintent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type protocolStep struct {
	Action       string
	Phase, Reads int
	Held         bool
}

func conforms(t *testing.T, steps []protocolStep) bool {
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
		t.Fatalf("Quint required, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		rec.Record(context.Background(), s.Action, nil, "obs.phase", s.Phase, "obs.reads", s.Reads, "obs.held", s.Held)
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
func TestRecordIntentProtocolRejectsHiddenAcquisitionAndLostHeldState(t *testing.T) {
	good := []protocolStep{{"local_intent", 0, 0, false}, {"stage_intent", 1, 0, false}, {"run_intent", 2, 1, true}, {"refuse_intent", 2, 1, true}, {"stage_intent", 1, 1, false}}
	if !conforms(t, good) {
		t.Fatal("explicit admission-stage-run-refusal protocol rejected")
	}
	for _, mutant := range []struct {
		name   string
		index  int
		change func(*protocolStep)
	}{
		{"refusal-acquires", 3, func(s *protocolStep) { s.Reads++ }},
		{"refusal-clears-held", 3, func(s *protocolStep) { s.Held = false }},
		{"stage-acquires", 1, func(s *protocolStep) { s.Reads++ }},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			bad := append([]protocolStep(nil), good...)
			mutant.change(&bad[mutant.index])
			if conforms(t, bad) {
				t.Fatal("unsafe recorded transition accepted")
			}
		})
	}
	if conforms(t, []protocolStep{{"run_intent", 2, 1, true}}) {
		t.Fatal("Run without admitted staged question accepted")
	}
}

type recordedFact struct {
	Name   string `json:"name"`
	Before *int   `json:"before"`
	After  int    `json:"after"`
	Phase  string `json:"phase"`
	Held   *bool  `json:"held"`
}
type recordedJourney struct {
	Version string         `json:"version"`
	Status  string         `json:"status"`
	Errors  []string       `json:"errors"`
	Facts   []recordedFact `json:"facts"`
	Calls   []struct {
		Path string `json:"path"`
	} `json:"calls"`
	Result struct {
		Count string            `json:"count"`
		Rows  []json.RawMessage `json:"rows"`
	} `json:"result"`
}

func TestObservedRecordIntentAdmissionLedger(t *testing.T) {
	path := os.Getenv("GOO_RECORD_INTENT_REPORT")
	if path == "" {
		t.Skip("actual completed browser ledger supplied by GOO_RECORD_INTENT_REPORT")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report recordedJourney
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.recordIntent417/1" || report.Status != "passed" || len(report.Errors) != 0 || len(report.Facts) != 5 || len(report.Calls) != 1 || report.Calls[0].Path != "/api/explore" || report.Result.Count != "140" || len(report.Result.Rows) != 100 {
		t.Fatal("complete real-source record-intent journey required")
	}
	expected := []struct{ name, action string }{
		{"shared-admission-and-history", "local_intent"},
		{"stage-exact-question", "stage_intent"},
		{"explicit-run", "run_intent"},
		{"invalid-stage-keeps-held-results", "refuse_intent"},
		{"stage-valid-after-refusal", "stage_intent"},
	}
	phases := map[string]int{"initial": 0, "draft": 1, "applied": 2}
	steps := make([]protocolStep, 0, len(report.Facts))
	previous := 0
	for i, f := range report.Facts {
		phase, ok := phases[f.Phase]
		if f.Name != expected[i].name || !ok || f.Held == nil {
			t.Fatalf("unrecognized or incomplete observed fact %+v", f)
		}
		if f.Before != nil && *f.Before != previous {
			t.Fatalf("unrecorded API call before %s", f.Name)
		}
		if f.After < previous {
			t.Fatalf("source budget moved backwards at %s", f.Name)
		}
		previous = f.After
		steps = append(steps, protocolStep{expected[i].action, phase, f.After, *f.Held})
	}
	if !conforms(t, steps) {
		t.Fatal("actual admission handoff and source budget violate model")
	}
	bad := append([]protocolStep(nil), steps...)
	bad[3].Held = false
	if conforms(t, bad) {
		t.Fatal("mutated actual refusal silently losing held evidence accepted")
	}
	bad = append([]protocolStep(nil), steps...)
	bad[3].Reads++
	if conforms(t, bad) {
		t.Fatal("mutated actual refusal acquiring a source accepted")
	}
}
