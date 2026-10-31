package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const recordStepHandoffJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as P from '../ui/question-protocol.mjs';const {step,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));let out;try{out=P.validateStep(step)}catch{}assert.equal(!!out,accepted);if(out){assert.deepEqual(P.readStepLink(P.stepLink(step)),step);assert.deepEqual(P.captureQuestion(step).explore,step.record);const generated=P.recordStep(step.record,step.view);assert.deepEqual(generated.scope.environmentSelector,step.record.environmentSelector);generated.record.environmentSelector.scope.mode='changed';assert.notEqual(step.record.environmentSelector.scope.mode,'changed');}`

func TestHegelRecordStepHandoffParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 5))
		scope := identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}
		if scope.Mode == "named" {
			scope.Value = "prod"
		}
		selection := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: scope}
		copied := *selection
		q := exploration.Query{Signal: []string{"logs", "spans"}[seed%2], EnvironmentSelector: selection, FromMS: 1001, ToMS: 61001, Limit: 10, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "canary"}}
		step := workbench.Step{Version: 1, Kind: "record", Scope: workbench.StepScope{EnvironmentSelector: &copied, FromMS: q.FromMS, ToMS: q.ToMS}, View: "Records", Record: &q}
		switch variant {
		case 1:
			step.Scope.EnvironmentSelector.Policy = "current-only"
			step.Scope.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "prod"}
		case 2:
			step.Scope.ToMS++
		case 3:
			step.Record.Environment = "ambiguous"
		case 4:
			step.Scope.FromNS = "1001000000"
		case 5:
			step.Scope.EnvironmentSelector.Policy = "future"
		}
		accepted := step.Validate() == nil
		if accepted != (variant == 0) {
			tc.Fatal("unexpected validation", variant)
		}
		before, _ := json.Marshal(step)
		if e := runNativeHistogramJS(map[string]any{"step": step, "accepted": accepted}, recordStepHandoffJS); e != nil {
			tc.Fatal(e)
		}
		after, _ := json.Marshal(step)
		if string(before) != string(after) {
			tc.Fatal("property mutated question")
		}
	}, options(100)...)
}
