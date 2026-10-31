package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"testing"
)

const focusAuthoringJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as P from '../ui/population-workspace.mjs';const {queries,signal,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));const before=structuredClone(queries);let out;try{out=P.focusedQuestion(queries,signal)}catch{}assert.equal(!!out,accepted);assert.deepEqual(queries,before);if(out){const expected=structuredClone(queries.find(q=>q.signal===signal));expected.offset=0;assert.deepEqual(out,expected);assert.equal(Object.hasOwn(out,'groupBy'),Object.hasOwn(expected,'groupBy'));out.where.all[0].value='changed';assert.deepEqual(queries,before);out.environmentSelector.scope.mode='changed';assert.deepEqual(queries,before);}`

func TestHegelFocusedQuestionPreservesRoot(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		bad := hegel.Draw(tc, hegel.Integers(0, 3))
		scope := identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}
		if scope.Mode == "named" {
			scope.Value = "prod"
		}
		selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: scope}
		if seed%4 == 0 {
			selector.Policy = "current-only"
			selector.Scope = identity.EnvironmentScope{Mode: "named", Value: "prod"}
		}
		primary := exploration.Query{Signal: "logs", EnvironmentSelector: selector, FromMS: 1001, ToMS: 61001, Limit: seed%99 + 1, Offset: 0, BucketMS: 1000, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "canary"}}}, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
		other := primary
		other.Signal = "spans"
		other.GroupBy = nil
		other.Summary = "span-duration"
		other.SpanStatus = true
		queries := []exploration.Query{primary, other}
		signal := []string{"logs", "spans"}[seed%2]
		accepted := true
		switch bad {
		case 1:
			signal = "metrics"
			accepted = false
		case 2:
			queries = append(queries, queries[seed%2])
			accepted = false
		case 3:
			queries = nil
			accepted = false
		}
		before, _ := json.Marshal(queries)
		if e := runNativeHistogramJS(map[string]any{"queries": queries, "signal": signal, "accepted": accepted}, focusAuthoringJS); e != nil {
			tc.Fatal(e)
		}
		after, _ := json.Marshal(queries)
		if string(before) != string(after) {
			tc.Fatal("focus mutated parent")
		}
	}, options(100)...)
}
