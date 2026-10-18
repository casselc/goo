package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type crossAppEvent struct {
	Action, Run, Subject, Trace, Environment, Members string
	FromMS                                            int64 `json:"fromMs"`
	ToMS                                              int64 `json:"toMs"`
	Count, Steps                                      int
	Passed                                            bool
}
type crossAppLedger struct {
	Version int
	Run     string
	Events  []crossAppEvent
}

// A run must contain all causal checkpoints. This gate prevents a truncated or
// empty trace from being reported as a passing experiment by a safety checker.
func (l crossAppLedger) complete() error {
	if l.Version != 1 || l.Run == "" || len(l.Events) != 4 {
		return fmt.Errorf("require version 1 and four correlated checkpoints")
	}
	seen := map[string]bool{}
	for _, e := range l.Events {
		if e.Run != l.Run || seen[e.Action] {
			return fmt.Errorf("mixed run or duplicate checkpoint")
		}
		switch e.Action {
		case "accepted", "source_ready", "displayed", "explored":
		default:
			return fmt.Errorf("unknown checkpoint %q", e.Action)
		}
		seen[e.Action] = true
	}
	return nil
}

func checkCrossApp(t *testing.T, l crossAppLedger) bool {
	t.Helper()
	if err := l.complete(); err != nil {
		return false
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink, qobs.WithActor("cross-app-coordinator"), qobs.WithProcess(l.Run))
	ctx := qobs.Thread(context.Background())
	for _, e := range l.Events {
		recorder.Record(ctx, e.Action, nil, "subject", e.Subject, "trace", e.Trace,
			"environment", e.Environment, "fromMs", e.FromMS, "toMs", e.ToMS,
			"count", e.Count, "members", e.Members, "steps", e.Steps, "passed", e.Passed)
	}
	checker, err := validate.NewChecker("model/cross_app.yaml", validate.Options{
		Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(ctx, sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(report.Summary())
	if path := os.Getenv("GOO_CROSS_APP_LEDGER"); path != "" && l.Run != "fixture" {
		dir := filepath.Dir(path)
		if err := os.WriteFile(filepath.Join(dir, "quint-report.txt"), []byte(report.Summary()), 0644); err != nil {
			t.Fatal(err)
		}
		var lines []byte
		for _, step := range sink.Steps() {
			line, err := step.MarshalLine()
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, line...)
			lines = append(lines, '\n')
		}
		if err := os.WriteFile(filepath.Join(dir, "quint-steps.jsonl"), lines, 0644); err != nil {
			t.Fatal(err)
		}
		for _, actor := range report.Actors {
			if actor.ITF != "" {
				data, err := os.ReadFile(actor.ITF)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, "quint-trace.itf.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	return report.OK()
}

func TestQuintCrossAppProtocolRejectsMutations(t *testing.T) {
	// Explicit protocol fixture. Actual browser evidence is checked separately.
	evidence := crossAppEvent{Run: "fixture", Subject: "job-1", Trace: strings.Repeat("a", 32),
		Environment: "test", FromMS: 1000, ToMS: 61000, Count: 1, Members: "0000000000000001"}
	source, display := evidence, evidence
	source.Action, display.Action = "source_ready", "displayed"
	base := crossAppLedger{Version: 1, Run: "fixture", Events: []crossAppEvent{
		{Action: "accepted", Run: "fixture", Subject: "job-1"}, source, display,
		{Action: "explored", Run: "fixture", Steps: 10, Passed: true},
	}}
	if !checkCrossApp(t, base) {
		t.Fatal("valid protocol rejected")
	}
	controls := []struct {
		name   string
		mutate func(*crossAppLedger)
	}{
		{"wrong_trace", func(l *crossAppLedger) { l.Events[2].Trace = strings.Repeat("b", 32) }},
		{"wrong_subject", func(l *crossAppLedger) { l.Events[1].Subject = "other-job" }},
		{"wrong_scope", func(l *crossAppLedger) { l.Events[2].Environment = "other" }},
		{"changed_window", func(l *crossAppLedger) { l.Events[2].FromMS++ }},
		{"wrong_members", func(l *crossAppLedger) { l.Events[2].Members = "0000000000000002" }},
		{"wrong_count", func(l *crossAppLedger) { l.Events[2].Count++ }},
		{"premature_display", func(l *crossAppLedger) { l.Events[1], l.Events[2] = l.Events[2], l.Events[1] }},
		{"unexercised", func(l *crossAppLedger) { l.Events[3].Steps = 0 }},
		{"explorer_failed", func(l *crossAppLedger) { l.Events[3].Passed = false }},
		{"mixed_run", func(l *crossAppLedger) { l.Events[2].Run = "other" }},
		{"truncated", func(l *crossAppLedger) { l.Events = l.Events[:3] }},
		{"empty", func(l *crossAppLedger) { l.Events = nil }},
	}
	for _, c := range controls {
		t.Run(c.name, func(t *testing.T) {
			l := base
			l.Events = append([]crossAppEvent(nil), base.Events...)
			c.mutate(&l)
			if checkCrossApp(t, l) {
				t.Fatal("invalid protocol accepted")
			}
		})
	}
}

func TestQuintCrossAppLiveLedger(t *testing.T) {
	path := os.Getenv("GOO_CROSS_APP_LEDGER")
	if path == "" {
		t.Skip("optional: set GOO_CROSS_APP_LEDGER to a real coordinated journey ledger")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var ledger crossAppLedger
	if err = json.NewDecoder(f).Decode(&ledger); err != nil {
		t.Fatal(err)
	}
	if !checkCrossApp(t, ledger) {
		t.Fatal("live cross-app ledger rejected")
	}
}
