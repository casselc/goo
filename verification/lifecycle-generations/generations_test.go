package generations_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chucklehead-dev/gotogether/connect"
	"github.com/chucklehead-dev/gotogether/itf"
	"github.com/chucklehead-dev/gotogether/qobs"
	"github.com/chucklehead-dev/gotogether/qtest"
	"github.com/chucklehead-dev/gotogether/qtrace"
	"github.com/chucklehead-dev/gotogether/validate"
)

const bindingPath = "model/generations.binding.yaml"

func recorder(sink qobs.Sink) *qobs.Recorder {
	return qobs.NewRecorder(sink, qobs.WithActor("generations"), qobs.WithProcess("kernel"))
}
func save(t testing.TB, name string, value any) {
	t.Helper()
	dir := os.Getenv("GENERATIONS_OUT")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	var data []byte
	var err error
	if steps, ok := value.([]qtrace.Step); ok {
		f, e := os.Create(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		err = qtrace.WriteJSONL(f, steps)
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	} else {
		data, err = json.MarshalIndent(value, "", "  ")
		if err == nil {
			err = os.WriteFile(filepath.Join(dir, name), append(data, '\n'), 0644)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
}
func directed(t testing.TB, ops []operation, complete bool) []qtrace.Step {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := recorder(sink)
	w := newWorld()
	for _, op := range ops {
		if err := w.apply(t.Context(), rec, op); err != nil {
			t.Fatalf("%s: %v", op.Action, err)
		}
	}
	if complete {
		if err := drain(t.Context(), w, rec); err != nil {
			t.Fatal(err)
		}
		rec.Settled(t.Context())
	}
	return sink.Steps()
}
func drain(ctx context.Context, w *world, rec *qobs.Recorder) error {
	if w.done {
		return nil
	}
	if w.phase == 0 {
		if err := w.apply(ctx, rec, operation{Action: "freeze"}); err != nil {
			return err
		}
	}
	if w.phase == 1 {
		if err := w.apply(ctx, rec, operation{Action: "correct"}); err != nil {
			return err
		}
	}
	for _, g := range []int{0, 1} {
		if err := w.apply(ctx, rec, operation{Action: "selectGeneration", Generation: g}); err != nil {
			return err
		}
	}
	if len(w.snapshots[0].value.Facts) > 0 && len(w.snapshots[1].value.Facts) > 0 {
		if err := w.apply(ctx, rec, operation{Action: "mixed"}); err != nil {
			return err
		}
	}
	if err := w.apply(ctx, rec, operation{Action: "changedScope"}); err != nil {
		return err
	}
	return w.apply(ctx, rec, operation{Action: "finish"})
}
func TestDirectedContract(t *testing.T) {
	scenarios := map[string][]operation{
		"scoped-identities":    {{Action: "accept", Tenant: 0}, {Action: "accept", Tenant: 1}, {Action: "terminal", Tenant: 0, Successful: true}, {Action: "terminal", Tenant: 1}, {Action: "freeze"}},
		"missing-to-good":      {{Action: "accept"}, {Action: "freeze"}, {Action: "terminal", Successful: true}},
		"good-to-unknown":      {{Action: "accept"}, {Action: "terminal", Successful: true}, {Action: "freeze"}, {Action: "terminal", Successful: false}},
		"orphan-to-known":      {{Action: "terminal", Tenant: 1, Successful: true}, {Action: "freeze"}, {Action: "accept", Tenant: 1}},
		"duplicate-redelivery": {{Action: "accept"}, {Action: "terminal", Successful: true}, {Action: "duplicate"}, {Action: "freeze"}, {Action: "duplicate"}},
	}
	for name, ops := range scenarios {
		t.Run(name, func(t *testing.T) {
			steps := directed(t, ops, true)
			qtest.Validate(t, bindingPath, steps).Expect(t, qtest.Conforms(), qtest.Settles())
			save(t, name+".jsonl", steps)
		})
	}
}
func TestGeneratedModelTraces(t *testing.T) {
	n := 12
	if testing.Short() {
		n = 6
	}
	files, _, err := connect.GenerateRun(t.Context(), connect.Config{Spec: "model/generations.qnt", Main: "generations", MaxSamples: n, MaxSteps: 25, Seed: "42", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	total, active := 0, 0
	for i, file := range files {
		trace, err := itf.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		w := newWorld()
		sink := &qobs.MemorySink{}
		rec := recorder(sink)
		d := connect.Switch{"init": connect.Skip}
		for _, name := range []string{"accept", "terminal", "duplicate", "freeze", "correct", "selectGeneration", "mixed", "changedScope", "finish", "idle"} {
			action := name
			d[action] = func(s *connect.Step) error {
				op := operation{Action: action}
				if action != "idle" {
					active++
				}
				if action == "accept" || action == "terminal" || action == "duplicate" {
					op.Tenant = connect.MustPick[int](s, "tenant")
				}
				if action == "terminal" {
					op.Successful = connect.MustPick[bool](s, "successful")
				}
				if action == "selectGeneration" {
					op.Generation = connect.MustPick[int](s, "generation")
				}
				return w.apply(t.Context(), rec, op)
			}
		}
		if err := connect.Replay(trace, d, connect.ReplayOptions{}); err != nil {
			t.Fatal(err)
		}
		// Judge the generated prefix before appending a deterministic completion.
		// A finite prefix is useful evidence but is not labelled a finished run.
		qtest.Validate(t, bindingPath, sink.Steps()).Expect(t, qtest.Conforms())
		if err := drain(t.Context(), w, rec); err != nil {
			t.Fatal(err)
		}
		rec.Settled(t.Context())
		qtest.Validate(t, bindingPath, sink.Steps()).Expect(t, qtest.Conforms(), qtest.Settles())
		save(t, fmt.Sprintf("generated-%02d.jsonl", i), sink.Steps())
		total += len(trace.States) - 1
	}
	if len(files) != n || total == 0 {
		t.Fatalf("missing generated traces: %d, %d steps", len(files), total)
	}
	t.Logf("%d model-generated traces / %d generated steps (%d non-idle); prefixes and completed extensions checked", len(files), total, active)
	save(t, "generated-summary.json", map[string]any{"traces": len(files), "generatedSteps": total, "nonIdleSteps": active, "seed": 42, "maxSteps": 25})
}
func TestObservationBinding(t *testing.T) {
	steps := directed(t, []operation{{Action: "accept"}, {Action: "terminal", Successful: true}, {Action: "freeze"}, {Action: "terminal"}}, true)
	rep := qtest.MutateSteps(t, bindingPath, steps, qtest.StepMutations(qtrace.MutateOps("corrupt"), qtrace.MutateLimit(12)))
	// Deleting a required observation is refused by the binding before a
	// semantic verdict. Keep those malformed logs outside the kill score.
	if len(rep.Killed()) == 0 || len(rep.Survived()) != 0 || len(rep.Undecided()) != 0 {
		t.Fatalf("observation mutations left unchecked: %s", rep)
	}
	if len(rep.Killed())+len(rep.Malformed()) != len(rep.Mutants) {
		t.Fatalf("unaccounted observation mutations: %s", rep)
	}
}

func TestFinitePrefix(t *testing.T) {
	steps := directed(t, []operation{{Action: "accept"}, {Action: "freeze"}}, false)
	qtest.Validate(t, bindingPath, steps).Expect(t, qtest.Conforms(), qtest.Not(qtest.Settles()))
	last := steps[len(steps)-1]
	settled := append(append([]qtrace.Step{}, steps...), qtrace.Step{Action: "settled", Actor: last.Actor, Recorder: last.Recorder, Seq: last.Seq + 1, Time: last.Time, Record: qtrace.Record{Settled: true}})
	qtest.Validate(t, bindingPath, settled).Expect(t, qtest.Violates(validate.EventuallyName("complete")))
}
