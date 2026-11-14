package shadow

import (
	"context"
	"github.com/chucklehead-dev/gotogether/qdriver"
	"github.com/chucklehead-dev/gotogether/qtest"
	"github.com/chucklehead-dev/gotogether/qtrace"
	"github.com/chucklehead-dev/gotogether/validate"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const binding = "acquisition.binding.yaml"

func options() validate.Options { return validate.Options{Backend: "typescript", Eval: "quint"} }
func TestModelPossible(t *testing.T) {
	t.Setenv("QUINT_BACKEND", "typescript")
	// A retained display before acquisition has no matching local view: the
	// model must catch that bad observation, with its witness replayed by Quint.
	qtest.Model(t, binding).Expect(t,
		qtest.Possible(Scenario()),
		qtest.Refutes("retainedOwnership").InAtMost(1))
	cmd := exec.Command("quint", "run", "model/acquisition.qnt", "--backend=typescript", "--max-samples=10", "--max-steps=10")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("model simulation: %v: %s", err, out)
	}
}

func copyBinding(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "model"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{binding, "model/acquisition.qnt"} {
		b, e := os.ReadFile(f)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(d, f), b, 0644); e != nil {
			t.Fatal(e)
		}
	}
	return filepath.Join(d, binding)
}
func compile(t *testing.T, b string) string {
	t.Helper()
	a, e := validate.Compile(b, options())
	if e != nil {
		t.Fatal(e)
	}
	p := validate.CompiledFor(b)
	if e = a.Write(p); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestRecordedJourney(t *testing.T) {
	p := os.Getenv("GOO_GOTOGETHER_REQUESTS")
	if p == "" {
		t.Skip("live browser log: run npm run test:journey:first-use-acquisition with GOO_GOTOGETHER_SHADOW=1")
	}
	l, e := qdriver.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	steps, e := Steps(l)
	if e != nil {
		t.Fatal(e)
	}
	b := copyBinding(t)
	artifact := compile(t, b)
	t.Setenv("GOTOGETHER_COMPILED", "0")
	start := time.Now()
	run := qtest.Validate(t, b, steps, qtest.Quint(options()))
	run.Expect(t, qtest.Finishes(), Scenario())
	sourceElapsed := time.Since(start)
	t.Setenv("GOTOGETHER_COMPILED", "1")
	t.Run("compiled-without-tools", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		for _, tool := range []string{"quint", "node"} {
			if path, err := exec.LookPath(tool); err == nil {
				t.Fatalf("%s unexpectedly available: %s", tool, path)
			}
		}
		start := time.Now()
		native := qtest.Validate(t, b, steps)
		native.Expect(t, qtest.Finishes(), Scenario())
		t.Logf("same request log: source cold wall=%s compiled cold wall=%s records=%d; Node/Quint absent from PATH; existing native admission gate remains authoritative", sourceElapsed, time.Since(start), len(l.Records))
	})
	for _, mode := range []string{"unanswered-call", "missing-final-show", "stale-bank", "hidden-final"} {
		t.Run(mode, func(t *testing.T) {
			mutated := append([]qdriver.Record(nil), l.Records...)
			for i := len(mutated) - 1; i >= 0; i-- {
				r := mutated[i]
				if mode == "unanswered-call" && r.Event == qdriver.Response {
					mutated = append(mutated[:i], mutated[i+1:]...)
					break
				}
				if r.Action == "show" && r.Args["phase"] == "final" {
					if mode == "missing-final-show" {
						mutated = append(mutated[:i], mutated[i+1:]...)
						break
					}
					if mode == "stale-bank" || mode == "hidden-final" {
						a := map[string]any{}
						for k, v := range r.Args {
							a[k] = v
						}
						if mode == "stale-bank" {
							a["bank"] = "obsolete-success-bank"
						} else {
							a["visible"] = false
						}
						mutated[i].Args = a
						break
					}
				}
			}
			ml, e := qdriver.New(l.Header, mutated)
			if e != nil {
				t.Fatal(e)
			}
			ms, e := Steps(ml)
			if e != nil {
				t.Fatal(e)
			}
			assertControls(t, b, artifact, ms, mode)

		})
	}
	t.Run("truncated-tail", func(t *testing.T) {
		raw, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		raw = append(raw, []byte(`{"event":`)...)
		f := filepath.Join(t.TempDir(), "torn.jsonl")
		if e = os.WriteFile(f, raw, 0644); e != nil {
			t.Fatal(e)
		}
		torn, err := qdriver.ReadFile(f)
		if err == nil || torn == nil {
			t.Fatalf("torn tail missing error/log: %v", err)
		}
		ts, err := Steps(torn)
		if err != nil {
			t.Fatal(err)
		}
		assertControls(t, b, artifact, ts, "truncated-tail")
	})
	t.Run("stale-artifact", func(t *testing.T) {
		model := filepath.Join(filepath.Dir(b), "model/acquisition.qnt")
		f, e := os.OpenFile(model, os.O_APPEND|os.O_WRONLY, 0644)
		if e != nil {
			t.Fatal(e)
		}
		_, e = f.WriteString("\n// deliberately stale\n")
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		c, e := validate.NewCompiledChecker(b, artifact, options())
		if c != nil {
			c.Close()
		}
		if e == nil || !strings.Contains(e.Error(), "stale") {
			t.Fatalf("stale model error=%v", e)
		}
	})
}

// assertControls compares the same immutable mutation with both evaluators.
func assertControls(t *testing.T, b, artifact string, steps []qtrace.Step, mode string) {
	t.Helper()
	var prior validate.Finish
	for _, backend := range []string{"source", "compiled"} {
		var c *validate.Checker
		var e error
		if backend == "source" {
			c, e = validate.NewChecker(context.Background(), b, options())
		} else {
			c, e = validate.NewCompiledChecker(b, artifact, options())
		}
		if e != nil {
			t.Fatal(e)
		}
		rep, e := c.Check(context.Background(), steps)
		c.Close()
		if e != nil {
			t.Fatal(e)
		}
		f := rep.Finished()
		if f.Verdict == validate.Pass {
			t.Fatalf("%s %s passed", backend, mode)
		}
		expected := map[string]string{"unanswered-call": "pending calls", "missing-final-show": "eventual", "stale-bank": "retainedOwnership", "truncated-tail": "lost tail", "hidden-final": "retainedOwnership"}[mode]
		if !strings.Contains(f.Reason+f.Why, expected) {
			t.Fatalf("%s %s unexpected failure: %+v", backend, mode, f)
		}
		if backend == "compiled" && (f.Verdict != prior.Verdict || f.Reason != prior.Reason) {
			t.Fatalf("judgment disagrees source=%+v compiled=%+v", prior, f)
		}
		prior = f
		t.Logf("%s %s: %s", backend, mode, f.Reason)
	}
}

// Protocol fixture exercises the reusable checker, not the application or DOM.
func TestProtocolControls(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "fixture.jsonl")
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	code := int64(200)
	records := []qdriver.Record{
		{Event: qdriver.Request, Req: "1", Action: "read", Time: at, Args: map[string]any{"read": 1, "epoch": 0}},
		{Event: qdriver.Response, Req: "1", Time: at.Add(time.Millisecond), Code: &code},
	}
	phases := []string{"explicit-read", "local-inspect", "retained-return", "cancel-pending-reread", "late-success-ignored", "final"}
	for i, phase := range phases {
		records = append(records, qdriver.Record{Event: qdriver.Step, Action: "show", Time: at.Add(time.Duration(i+2) * time.Millisecond), Args: map[string]any{"phase": phase, "visible": true, "bank": "DOM bank", "clock": "DOM window", "question": `{"service":"quoted \"service\"","instrument":"gauge"}`, "page": 1, "find": "needle", "detail": "DOM detail \"quote\"\nnext line", "epoch": 0}})
	}
	failedCode := int64(0) // an answered transport failure completes only its own call.
	records = append(records,
		qdriver.Record{Event: qdriver.Request, Req: "2", Action: "read", Time: at.Add(20 * time.Millisecond), Args: map[string]any{"read": 2, "epoch": 0}},
		qdriver.Record{Event: qdriver.Response, Req: "2", Time: at.Add(21 * time.Millisecond), Code: &failedCode},
	)
	records = append(records, qdriver.Record{Event: qdriver.Step, Action: "drained", Time: at.Add(time.Second)})
	l, e := qdriver.New(qdriver.Header{Protocol: qdriver.Protocol, Recorder: "fixture", Actor: "acquisition"}, records)
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	if e = l.Write(f); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	t.Setenv("GOO_GOTOGETHER_REQUESTS", p)
	TestRecordedJourney(t)
}
