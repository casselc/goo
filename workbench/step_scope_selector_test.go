package workbench

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func typedRecordStep(t *testing.T, policy, mode, value string) Step {
	t.Helper()
	selector := identity.EnvironmentSelector{Version: "goo/environment-selector/v1", Policy: policy, Scope: identity.EnvironmentScope{Mode: mode, Value: value}}
	if _, err := selector.Selection(); err != nil {
		t.Fatal(err)
	}
	q := exploration.Query{Signal: "logs", EnvironmentSelector: &selector, FromMS: 1000, ToMS: 61000, Limit: 10, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "canary"}}
	return Step{Version: 1, Kind: "record", Scope: StepScope{EnvironmentSelector: &selector, FromMS: q.FromMS, ToMS: q.ToMS}, View: "Records", Record: &q}
}

func TestRecordStepTypedScopeAndImmutableDocumentRoundTrip(t *testing.T) {
	for _, tc := range []struct{ policy, mode, value string }{{"current-only", "named", "prod"}, {"effective", "named", "prod"}, {"effective", "empty", ""}, {"effective", "unreported", ""}} {
		t.Run(tc.policy+"/"+tc.mode, func(t *testing.T) {
			step := typedRecordStep(t, tc.policy, tc.mode, tc.value)
			if err := step.Validate(); err != nil {
				t.Fatal(err)
			}
			store := Store{Dir: t.TempDir()}
			card := Card{Kind: "question", Step: &step, Key: strings.Repeat("a", 64)}
			first, err := store.CreateDocument(DocumentInput{Title: "Typed question", Cards: []Card{card}}, "", nil)
			if err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(first)
			card.Caption = "New note"
			child, err := store.CreateDocument(DocumentInput{Title: "Typed question", Cards: []Card{card}}, first.ID, nil)
			if err != nil || child.Parent != first.ID {
				t.Fatal(child, err)
			}
			old, err := store.LoadDocument(first.ID)
			after, _ := json.Marshal(old)
			if err != nil || string(before) != string(after) || !reflect.DeepEqual(*old.Cards[0].Step, step) {
				t.Fatal("scope or predicates changed", err)
			}
		})
	}
}

func TestRecordStepChecksPolicyBoundsAndMeasurementCapability(t *testing.T) {
	step := typedRecordStep(t, "effective", "named", "prod")
	clone := func() Step { raw, _ := json.Marshal(step); var s Step; json.Unmarshal(raw, &s); return s }
	for _, mutate := range []func(*Step){
		func(s *Step) { s.Scope.EnvironmentSelector = nil; s.Scope.Environment = "prod" },
		func(s *Step) { s.Scope.Environment = "prod" },
		func(s *Step) { s.Scope.EnvironmentSelector.Policy = "current-only" },
		func(s *Step) { s.Scope.EnvironmentSelector.Scope.Value = "other" },
		func(s *Step) { s.Scope.ToMS++ }, func(s *Step) { s.Scope.FromNS = "1000000000" },
		func(s *Step) { s.Scope.EnvironmentSelector.Scope.Mode = "unreported" },
	} {
		bad := clone()
		mutate(&bad)
		if bad.Validate() == nil {
			t.Fatalf("foreign/ambiguous scope accepted: %#v", bad.Scope)
		}
	}
	current := typedRecordStep(t, "current-only", "named", "prod")
	current.Scope.EnvironmentSelector = nil
	current.Scope.Environment = "prod"
	if err := current.Validate(); err != nil {
		t.Fatal("equivalent current-only representation refused", err)
	}
	current.Kind = "measurement"
	current.Record = nil
	if current.Validate() == nil {
		t.Fatal("measurement without definition accepted")
	}
}

func TestTypedStepScopeLeavesMeasurementCapabilityExplicit(t *testing.T) {
	raw, err := os.ReadFile("../config/measurements/worker-utilization.yaml")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := measurements.Read(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	step := Step{Version: 1, Kind: "measurement", Scope: StepScope{Environment: definition.Environment, FromMS: 1000, ToMS: 2000}, View: "Overview", Measurement: &definition}
	if err := step.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, policy := range []string{"current-only", "effective"} {
		step.Scope.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: policy, Scope: identity.EnvironmentScope{Mode: "named", Value: definition.Environment}}
		if step.Validate() == nil {
			t.Fatal("measurement typed scope implicitly enabled")
		}
	}
}

func TestTypedRecordStepGoJSProtocolParity(t *testing.T) {
	for _, tc := range []struct{ policy, mode, value string }{{"current-only", "named", "prod"}, {"effective", "named", "prod"}, {"effective", "empty", ""}, {"effective", "unreported", ""}} {
		step := typedRecordStep(t, tc.policy, tc.mode, tc.value)
		raw, _ := json.Marshal(step)
		script := `const fs=require('node:fs'),assert=require('node:assert/strict'),p=require('../ui/question-protocol.mjs'),s=JSON.parse(fs.readFileSync(0,'utf8'));p.validateStep(s);const out=p.readStepLink(p.stepLink(s));assert.deepEqual(out,s);assert.deepEqual(p.captureQuestion(out).explore,s.record);process.stdout.write(JSON.stringify(out));`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = strings.NewReader(string(raw))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("typed JS roundtrip: %v %s", err, out)
		}
		var back Step
		if err = json.Unmarshal(out, &back); err != nil || !reflect.DeepEqual(step, back) {
			t.Fatal("Go/JS changed typed step", err, string(out))
		}
	}
}
