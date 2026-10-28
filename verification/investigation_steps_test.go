package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelInvestigationSteps(t *testing.T) {
	b, e := os.ReadFile("../config/measurements/worker-utilization.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(b))
	if e != nil {
		t.Fatal(e)
	}
	hegel.Test(t, func(tc *hegel.T) {
		from := 1700000000000 + hegel.Draw(tc, hegel.Integers(100000, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7200000))
		q := measurements.Question{Definition: d, FromMS: int64(from), ToMS: int64(from + width)}
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			q.FromNS = strconv.FormatInt(q.FromMS*1000000+123, 10)
			q.ToNS = strconv.FormatInt(q.ToMS*1000000-456, 10)
		}
		input, _ := json.Marshal(q)
		script := `const assert=require('node:assert/strict'),w=require('../ui/question-protocol.mjs'),compat=require('../ui/workspace.mjs');for(const key of Object.keys(w))assert.equal(w[key],compat[key]);const q=JSON.parse(process.argv[1]),s=w.measurementStep(q),r=w.recordStep({signal:'spans',environment:q.definition.environment,fromMs:q.fromMs,toMs:q.toMs,limit:100,offset:0});assert.equal(w.sameWindow(r.scope,s.scope),!q.fromNS);assert(!w.sameWindow(r.scope,{...s.scope,toMs:s.scope.toMs+1}));const actual=w.readStepLink(w.stepLink(s)),empty=w.captureQuestion(actual),omitted=structuredClone(actual);empty.measurement.definition.source.version='';empty.measurement.definition.source.schemaURL='';delete omitted.measurement.source.version;delete omitted.measurement.source.schemaURL;assert.deepEqual(w.evidenceDifference(omitted,empty),[]);empty.measurement.definition.source.version='different';assert(w.evidenceDifference(omitted,empty).includes('Question definition differs'));q.definition.source.unit='mutated';assert.notEqual(actual.measurement.source.unit,'mutated');process.stdout.write(JSON.stringify(actual));`
		out, e := exec.Command("node", "-e", script, string(input)).CombinedOutput()
		if e != nil {
			tc.Fatalf("actual JS: %s %v", out, e)
		}
		var step workbench.Step
		if json.Unmarshal(out, &step) != nil || step.Validate() != nil || !reflect.DeepEqual(step.Measurement, &d) || step.Scope.FromMS != q.FromMS || step.Scope.ToMS != q.ToMS || step.Scope.FromNS != q.FromNS || step.Scope.ToNS != q.ToNS {
			tc.Fatal("step roundtrip changed native identity or scope")
		}
	}, options(100)...)
}
func TestQuintInvestigationStepsLive(t *testing.T) {
	name := os.Getenv("GOO_STEPS_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Saved                                                 workbench.Document
			Record                                                exploration.Result
			Measurement                                           measurements.Result
			BeforeReopen, AfterReopen, RecipientReads, BeforeSave int
			Events                                                []struct {
				Action string
				Reads  int
			}
			Rerun measurements.Result
		}
		if json.Unmarshal(data, &l) != nil {
			t.Fatal("bad ledger")
		}
		if mutation == 0 {
			l.Saved.Cards[0].Step.Scope.ToMS++
		}
		if mutation == 1 {
			l.Measurement.Question.Definition.Source.Unit = "foreign"
		}
		if mutation == 2 {
			l.AfterReopen++
		}
		if mutation == 3 {
			l.RecipientReads++
		}
		if mutation == 5 {
			l.Events[1].Reads++
		}
		if mutation == 6 {
			l.Events[2].Reads--
		}
		if mutation == 4 {
			l.Saved.Cards[3].Step.Scope.ToMS = l.Saved.Cards[2].Step.Scope.ToMS
		}
		coherent := l.Saved.DocumentInput.Validate() == nil && len(l.Saved.Cards) == 4 && l.BeforeReopen == l.AfterReopen && l.RecipientReads == 0 && len(l.Events) == 3 && reflect.DeepEqual(l.Rerun.Question, l.Measurement.Question)
		if coherent {
			a, b, c := l.Saved.Cards[0].Step, l.Saved.Cards[2].Step, l.Saved.Cards[3].Step
			coherent = a != nil && b != nil && c != nil && a.Scope == b.Scope && b.Scope != c.Scope && a.Record.Environment == l.Record.Question.Environment && a.Record.FromMS == l.Record.Question.FromMS && a.Record.ToMS == l.Record.Question.ToMS && reflect.DeepEqual(b.Measurement, &l.Measurement.Question.Definition) && b.Scope.FromMS == l.Measurement.Question.FromMS && b.Scope.ToMS == l.Measurement.Question.ToMS && l.Saved.Cards[1].Kind == "capture"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			if event.Action == "run" {
				event.Action = "executeRead"
			}
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads-l.BeforeSave, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/investigation_steps.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed mixed-signal investigation rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

// Extend the existing investigation swarm with the new typed question cards.
func (m *notesContextMachine) RuleQuestionStep(tc hegel.TestCase) {
	if len(m.draft.Cards) >= 6 {
		return
	}
	scope := workbench.StepScope{Environment: m.environment, FromMS: m.from, ToMS: m.from + 100}
	step := workbench.Step{Version: 1, Scope: scope}
	if hegel.Draw(tc, hegel.Integers(0, 1)) == 0 {
		step.Kind = "record"
		step.View = "Records"
		step.Record = &exploration.Query{Signal: "spans", Environment: scope.Environment, FromMS: scope.FromMS, ToMS: scope.ToMS, Limit: 100}
	} else {
		b, e := os.ReadFile("../config/measurements/worker-utilization.yaml")
		if e != nil {
			panic(e)
		}
		d, e := measurements.Read(string(b))
		if e != nil {
			panic(e)
		}
		d.Environment = scope.Environment
		step.Kind = "measurement"
		step.View = "Overview"
		step.Measurement = &d
	}
	m.draft.Cards = append(m.draft.Cards, workbench.Card{Kind: "question", Step: &step})
}
func (m *notesContextMachine) RuleReviseQuestionStep(tc hegel.TestCase) {
	indices := []int{}
	for i, c := range m.draft.Cards {
		if c.Kind == "question" {
			indices = append(indices, i)
		}
	}
	if len(indices) == 0 {
		return
	}
	index := indices[hegel.Draw(tc, hegel.Integers(0, len(indices)-1))]
	// Reopening and editing owns a detached definition, including after forking.
	raw, e := json.Marshal(m.draft.Cards[index].Step)
	if e != nil {
		panic(e)
	}
	var step workbench.Step
	if e = json.Unmarshal(raw, &step); e != nil {
		panic(e)
	}
	step.Scope.Environment = "revised"
	step.Scope.FromMS++
	step.Scope.ToMS++
	if step.Record != nil {
		step.Record.Environment = step.Scope.Environment
		step.Record.FromMS = step.Scope.FromMS
		step.Record.ToMS = step.Scope.ToMS
	} else {
		step.Measurement.Environment = step.Scope.Environment
	}
	m.draft.Cards[index].Step = &step
}
