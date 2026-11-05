package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// The record drill's authority is the selected operand's exact scope and read
// identity. A matching displayed count cannot renew a retired operand intent.
func TestHegelComparisonRecordDrill334(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := []string{"named", "empty", "unreported"}[seed%3]
		env := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
		if mode == "named" {
			env.Scope.Value = "prod-'雪"
			if seed%2 == 0 {
				env.Policy = "current-only"
			}
		}
		ns := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}
		if seed%3 == 0 {
			ns.Op = "missing"
		} else if seed%3 == 2 {
			ns.Value = "team-'雪"
		}
		from := int64(1790000000000 + seed*1000)
		q := exploration.Query{Signal: []string{"logs", "spans"}[seed%2], EnvironmentSelector: env, FromMS: from, ToMS: from + 60000, Limit: 1 + seed%100, Offset: seed % 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}, BucketMS: 1000, Where: &exploration.Predicate{All: []exploration.Predicate{ns, {Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: fmt.Sprintf("'雪%d", seed)}, {Field: &exploration.Field{Kind: "attribute", Key: "fallback"}, Op: "exists"}}}}}}
		b := q
		b.FromMS -= 60000
		b.ToMS -= 60000
		value := []*string{nil, textPointer292(""), textPointer292("api-'雪")}[seed%3]
		cohort := func(question exploration.Query, basis string) exploration.Cohort {
			return exploration.Cohort{Question: question, Basis: basis, CapturedAt: "2026-11-05T10:00:00Z", Count: "107", State: "observed", Groups: []exploration.ValueCount{{Value: value, Count: "107"}}}
		}
		request := exploration.ComparisonRequest{Baseline: cohort(b, "baseline-source"), Selection: cohort(q, "selected-source")}
		result, err := exploration.CompareCohorts(request)
		if err != nil {
			tc.Fatal(err)
			return
		}
		before, _ := json.Marshal(request)
		for _, side := range []string{"baseline", "selection"} {
			snapshot := result.GroupComparison.Baseline
			root := b
			if side == "selection" {
				snapshot = result.GroupComparison.Selection
				root = q
			}
			intent := exploration.GroupSideIntent{Type: "scope-group-side", Side: side, Key: result.GroupComparison.Rows[0].Key, DatasetID: snapshot.ID, Revision: snapshot.Revision, ScopeID: snapshot.Scope.ID, Generation: snapshot.Scope.Generation}
			next, err := exploration.StageGroupSide(request, intent)
			if err != nil {
				tc.Fatal(err)
				return
			}
			if err = next.Validate(); err != nil {
				tc.Fatal(err)
			}
			plan, err := exploration.Compile(next)
			if err != nil {
				tc.Fatal(err)
				return
			}
			if plan.Rows == "" || plan.Count == "" || plan.Parameters["from"] != fmt.Sprint(root.FromMS) || plan.Parameters["to"] != fmt.Sprint(root.ToMS) {
				tc.Fatal("record/count plan lost operand bounds")
			}
			originalWhere := root.Where
			root.Offset = 0
			root.Where = next.Where
			if !reflect.DeepEqual(root, next) || !reflect.DeepEqual(next.Where.All[0], *originalWhere) {
				tc.Fatal("drill changed operand definition")
			}
			cases := []map[string]any{{"intent": intent, "accepted": true, "query": next, "side": side, "key": intent.Key, "mutation": 0}}
			if err = runNativeHistogramJS(map[string]any{"request": request, "result": result, "cases": cases}, comparisonDrillJS334); err != nil {
				tc.Fatal(err)
				return
			}
			// Same total, new source clock/basis: old scope action is retired.
			changed := request
			if side == "baseline" {
				changed.Baseline.Basis += "-new"
				changed.Baseline.CapturedAt = "2026-11-05T10:01:00Z"
			} else {
				changed.Selection.Basis += "-new"
				changed.Selection.CapturedAt = "2026-11-05T10:01:00Z"
			}
			if _, err = exploration.StageGroupSide(changed, intent); err == nil {
				tc.Fatal("same count renewed old source identity")
			}
		}
		after, _ := json.Marshal(request)
		if string(before) != string(after) {
			tc.Fatal("drill mutated retained parent pair")
		}
	}, options(30)...)
}

const comparisonDrillJS334 = `
import assert from 'node:assert/strict';import fs from 'node:fs';import * as P from '../ui/group-side-staging.mjs';import {identity} from '../ui/question-session.mjs';
const {result,cases}=JSON.parse(fs.readFileSync(0,'utf8')),g=result.groupComparison,before=JSON.stringify(g);globalThis.fetch=()=>{throw Error('planner attempted IO')};
for(const c of cases){const q=P.stagedGroupQuestion(g,c.intent);assert.equal(identity(q),identity(c.query));assert.equal(JSON.stringify(g),before);assert.throws(()=>P.stagedGroupQuestion(g,{...c.intent,generation:'retired'}));assert.throws(()=>P.stagedGroupQuestion(g,{...c.intent,query:q}));const invalid=structuredClone(g);invalid.state='diagnostic';assert.throws(()=>P.stagedGroupQuestion(invalid,c.intent));}
`
