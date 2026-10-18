package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"testing"
)

type telemetryPackageMachine struct {
	p     exploration.TelemetryPackage
	store exploration.PackageStore
	saved map[string]string
	clock int64
}

func (m *telemetryPackageMachine) RuleEdit(tc hegel.TestCase) {
	values := []string{"", "雪\\path\t", "x'] OR 1=1 --", "region"}
	v := values[hegel.Draw(tc, hegel.Integers(0, 3))]
	m.p.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: v}
}
func (m *telemetryPackageMachine) RuleWindow(tc hegel.TestCase) {
	m.clock += int64(hegel.Draw(tc, hegel.Integers(1, 100000)))
}
func (m *telemetryPackageMachine) RuleSave(tc hegel.TestCase) {
	saved, e := m.store.Save(m.p)
	if e != nil {
		panic(e)
	}
	data, _ := json.Marshal(saved.Definition)
	m.saved[saved.Revision] = string(data)
}
func (m *telemetryPackageMachine) InvariantRoundtripAndImmutable(tc hegel.TestCase) {
	plan, e := exploration.PlanTelemetryPackage(m.p, m.clock, m.clock+60000, nil)
	if e != nil {
		panic(e)
	}
	parsed, e := exploration.ReadTelemetryPackage(plan.YAML)
	if e != nil {
		panic(e)
	}
	id, e := parsed.Revision()
	if e != nil || id != plan.Revision {
		panic("YAML changed meaning")
	}
	later, e := exploration.PlanTelemetryPackage(parsed, m.clock+1, m.clock+60001, nil)
	if e != nil || later.Revision != id || later.Question.Offset != 0 || later.Question.Limit != 100 {
		panic("preview changed reusable definition")
	}
	for revision, expected := range m.saved {
		saved, e := m.store.Load(revision)
		if e != nil {
			panic(e)
		}
		data, _ := json.Marshal(saved.Definition)
		if string(data) != expected {
			panic("old revision mutated")
		}
	}
	bad := parsed
	bad.Measure = "business-success"
	if bad.Validate() == nil {
		panic("inferred business semantics accepted")
	}
}
func telemetryDefinition(t *testing.T) exploration.TelemetryPackage {
	t.Helper()
	p, e := exploration.DraftTelemetryPackage(exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1, ToMS: 60001, Limit: 10, Offset: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "region"}}}, "jobs", "Jobs", 5)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestHegelTelemetryPackageStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &telemetryPackageMachine{p: telemetryDefinition(t), store: exploration.PackageStore{Dir: t.TempDir()}, saved: map[string]string{}, clock: 1000}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantRoundtripAndImmutable"))
	}, options(100)...)
}
func TestQuintObservedTelemetryPackages(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("packages"), qobs.WithProcess("fixture"))
	p := telemetryDefinition(t)
	store := exploration.PackageStore{Dir: t.TempDir()}
	original, e := store.Save(p)
	if e != nil {
		t.Fatal(e)
	}
	for _, action := range []string{"draft", "save", "edit", "reopen", "apply"} {
		if action == "edit" {
			p.Title = "Edited Jobs"
		}
		plan, e := exploration.PlanTelemetryPackage(p, 1000, 61000, nil)
		if e != nil {
			t.Fatal(e)
		}
		parsed, e := exploration.ReadTelemetryPackage(plan.YAML)
		if e != nil {
			t.Fatal(e)
		}
		revision, e := parsed.Revision()
		if e != nil {
			t.Fatal(e)
		}
		old, e := store.Load(original.Revision)
		if e != nil {
			t.Fatal(e)
		}
		reusable := revision == plan.Revision && plan.Question.Offset == 0 && plan.Question.Where == nil
		immutable := old.Definition.Title == "Jobs"
		if action == "apply" {
			later, e := exploration.PlanTelemetryPackage(parsed, 9999, 69999, nil)
			if e != nil {
				t.Fatal(e)
			}
			reusable = reusable && later.Revision == revision && later.Question.FromMS != plan.Question.FromMS
		}
		rec.Record(context.Background(), action, nil, "obs.reusable", reusable, "obs.immutable", immutable)
	}
	checker, e := validate.NewChecker("model/telemetry_packages.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	artifact("telemetry-packages-validation.txt", []byte(report.Summary()))
	for _, control := range []struct {
		i   int
		key string
	}{{3, "immutable"}, {4, "reusable"}} {
		steps[control.i].Obs[control.key] = false
		bad, e := checker.Check(context.Background(), steps)
		if e != nil || bad.OK() {
			t.Fatalf("negative control accepted %v", e)
		}
		artifact("telemetry-packages-negative-"+control.key+".txt", []byte(bad.Summary()))
		steps[control.i].Obs[control.key] = true
	}
}
