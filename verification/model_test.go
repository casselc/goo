package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/connect"
	"github.com/casselc/observability/quintgo/itf"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func quintPath(t *testing.T) string {
	t.Helper()
	p := os.Getenv("QUINT_BIN")
	if p == "" {
		p = "../node_modules/.bin/quint"
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("Quint required: run npm ci in the repository root: %v", err)
	}
	return abs
}

func TestQuintModelTracesDriveRealPreparation(t *testing.T) {
	q := quintPath(t)
	spec, _ := filepath.Abs("model/lifecycle.qnt")
	files, err := (connect.Generate{Quint: q, Spec: spec, Main: "lifecycleInstance", Traces: 8, MaxSteps: 20, Seed: 42, Backend: "typescript", Dir: t.TempDir()}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	steps := 0
	for i, file := range files {
		trace, err := itf.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		w := newWorld()
		driver := connect.Driver{Init: func(_ map[string]itf.Value) error { return nil }, Actions: map[string]connect.Handler{}}
		for _, name := range []string{"accept", "terminal", "duplicate", "advance95", "advance96", "advance100", "snapshot"} {
			action := name
			driver.Actions[action] = func(s *connect.Step) error {
				enabled, err := w.apply(action)
				if err != nil {
					return err
				}
				if !enabled {
					return fmt.Errorf("model requested disabled action %s", action)
				}
				return nil
			}
		}
		driver.Check = func(s *connect.Step) error {
			if err := w.check(); err != nil {
				return err
			}
			vars := s.After
			if vars["accepted"].B != w.accepted || vars["terminalSeen"].B != w.terminal || vars["basis"].Int64() != int64(w.basis) || len(vars["snapshots"].Elems) != len(w.snapshots) {
				return fmt.Errorf("implementation state differs from model")
			}
			for j, v := range vars["snapshots"].Elems {
				r := w.snapshots[j].value.Receipt
				if v.Get("eligible").Int64() != int64(r.Counts["eligible"]) || v.Get("good").Int64() != int64(r.Counts["good"]) || v.Get("bad").Int64() != int64(r.Counts["bad"]) || v.Get("unknown").B != (r.State == "unknown") {
					return fmt.Errorf("snapshot %d differs", j)
				}
			}
			return nil
		}
		if err := driver.Replay(trace); err != nil {
			t.Fatal(err)
		}
		steps += len(trace.States) - 1
		b, _ := os.ReadFile(file)
		if err := artifact(fmt.Sprintf("model-trace-%02d.itf.json", i), b); err != nil {
			t.Fatal(err)
		}
	}
	if len(files) == 0 || steps == 0 {
		t.Fatal("no model traces executed")
	}
	t.Logf("replayed %d generated traces, %d steps against lifecycle.Prepare", len(files), steps)
}

func TestQuintValidatesObservedTraceAndRejectsMutation(t *testing.T) {
	q := quintPath(t)
	sink := &qobs.MemorySink{}
	w := newWorld()
	w.recorder = qobs.NewRecorder(sink, qobs.WithActor("lifecycle"), qobs.WithProcess("run"))
	for _, action := range []string{"accept", "snapshot", "terminal", "advance95", "duplicate", "advance96", "snapshot", "advance100"} {
		enabled, err := w.apply(action)
		if err != nil || !enabled {
			t.Fatalf("%s: %v", action, err)
		}
	}
	checker, err := validate.NewChecker("model/binding.yaml", validate.Options{Quint: q, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	steps := sink.Steps()
	report, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatal(report.Summary())
	}
	for i, actor := range report.Actors {
		if actor.ITF == "" {
			t.Fatal("conforming trace did not produce ITF")
		}
		b, err := os.ReadFile(actor.ITF)
		if err != nil {
			t.Fatal(err)
		}
		if err := artifact(fmt.Sprintf("observed-trace-%02d.itf.json", i), b); err != nil {
			t.Fatal(err)
		}
	}
	var lines []byte
	for i := range steps {
		b, err := steps[i].MarshalLine()
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, b...)
		lines = append(lines, '\n')
	}
	if err := artifact("observed-steps.jsonl", lines); err != nil {
		t.Fatal(err)
	}
	if err := artifact("trace-validation.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	// An impossible observed good count must fail conformance, proving this
	// validates implementation observations rather than just model syntax.
	steps[0].Obs["good"] = int64(2)
	bad, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if bad.OK() {
		t.Fatal("impossible observed good count was accepted")
	}
	b, _ := json.MarshalIndent(map[string]any{"generated_traces": 8, "trace_conformance": "pass", "mutation_rejected": true}, "", "  ")
	if err := artifact("summary.json", b); err != nil {
		t.Fatal(err)
	}
}
