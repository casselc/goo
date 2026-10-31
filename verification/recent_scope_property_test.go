package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"testing"
)

const recentScopeJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as P from '../ui/population-workspace.mjs';import * as E from '../ui/record-environment.mjs';const {question,plan,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));let out;try{out=P.plannedQueries(question,null,plan)}catch{}assert.equal(!!out,accepted);if(out){assert.deepEqual(out,plan.queries);const before=structuredClone(question),fresh=P.recentQuestion(question,1800000000000,5);assert.ok(E.sameQuestionEnvironment(fresh,question));assert.deepEqual(fresh.where,question.where);assert.equal(fresh.toMs-fresh.fromMs,300000);assert.equal(fresh.offset,0);assert.deepEqual(question,before);fresh.environmentSelector.scope.mode='changed';assert.notEqual(question.environmentSelector.scope.mode,'changed');}`

func TestHegelRecentScopePlanParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 5))
		scope := identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}
		if scope.Mode == "named" {
			scope.Value = "prod"
		}
		selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: scope}
		if seed%4 == 0 {
			selector.Policy = "current-only"
			selector.Scope = identity.EnvironmentScope{Mode: "named", Value: "prod"}
		}
		q := exploration.Query{Signal: []string{"logs", "spans"}[seed%2], EnvironmentSelector: selector, FromMS: 1001, ToMS: 61001, Limit: 10, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "canary"}}
		plan, e := exploration.PlanPopulation(exploration.PopulationRequest{Question: q})
		if e != nil {
			tc.Fatal(e)
		}
		before, _ := json.Marshal(q)
		switch variant {
		case 1:
			plan.Queries[0] = plan.Queries[0].WithEnvironmentFrom(exploration.Query{Environment: "wrong"})
		case 2:
			plan.Queries[0].ToMS++
		case 3:
			plan.Queries[0].Where.Value = "changed"
		case 4:
			plan.Queries[1].Signal = plan.Queries[0].Signal
		case 5:
			plan.Version = "future"
		}
		if e := runNativeHistogramJS(map[string]any{"question": q, "plan": plan, "accepted": variant == 0}, recentScopeJS); e != nil {
			tc.Fatal(e)
		}
		after, _ := json.Marshal(q)
		if string(before) != string(after) {
			tc.Fatal("planner or presentation mutated input")
		}
	}, options(100)...)
}
