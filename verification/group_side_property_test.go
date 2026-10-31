package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"testing"
)

const groupSideParityJS = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import * as P from '../ui/group-side-staging.mjs';import * as S from '../ui/source-snapshots.mjs';import {identity} from '../ui/question-session.mjs';
const {request,result,cases}=JSON.parse(fs.readFileSync(0,'utf8')),g=result.groupComparison,admitted=S.importComparisonResult(result,request).groupComparison;
const before=JSON.stringify(g),heldBefore=JSON.stringify(admitted);
function mutate(intent,kind){const copy={...intent};if(kind===1)copy.side=copy.side==='baseline'?'selection':'baseline';if(kind===2)copy.key='1unknown';if(kind===3)copy.revision='stale';if(kind===4)copy.generation='old';return copy;}
for(const c of cases){
 let actual,error;try{actual=P.stageGroupSide(g,c.intent)}catch(e){error=e.message}
 assert.equal(!!actual,c.accepted,JSON.stringify({frame:'wire',intent:c.intent,error}));
 const frame=admitted[c.side];let local,localIntent;try{localIntent=mutate({type:'scope-group-side',side:c.side,key:c.key,datasetId:frame.id,revision:frame.revision,scopeId:frame.scope.id,generation:frame.scope.generation},c.mutation);local=P.stageGroupSide(admitted,localIntent)}catch{}
 assert.equal(!!local,c.accepted,JSON.stringify({frame:'admitted',side:c.side,key:c.key,mutation:c.mutation}));
 if(actual){assert.equal(identity(actual),identity(c.query));assert.equal(identity(local),identity(c.query));assert.equal(actual.offset,0);
  const root=g[c.intent.side].scope.selector,row=g.rows.find(r=>r.key===c.intent.key),leaf={field:g.field,op:row.value===null?'missing':'eq',...(row.value===null?{}:{value:row.value})};
  assert.deepEqual(actual.where,{all:[root.where,leaf]});assert.deepEqual(local.where,{all:[root.where,leaf]});
  assert.throws(()=>P.stageGroupSide(admitted,c.intent));assert.throws(()=>P.stageGroupSide(g,localIntent));
  actual.where.all[0].any[0].value='changed';local.where.all[0].any[0].value='changed';
 }
 assert.equal(JSON.stringify(g),before);assert.equal(JSON.stringify(admitted),heldBefore);
}
for(const side of ['baseline','selection'])for(const row of g.rows){const expected=row[side].presence==='present'&&row[side].state==='derived'&&g.state!=='diagnostic';let made;try{made=P.groupSideIntent(g,side,row.key)}catch{}assert.equal(!!made,expected);if(made){assert.equal(made.type,'scope-group-side');assert.equal(made.datasetId,g[side].id);assert.equal(made.revision,g[side].revision);assert.throws(()=>P.stageGroupSide(g,{...made,query:request[side].question}))}}
`

func TestHegelGroupSideStagesExactDetachedQuestion(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 4))
		selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}}
		if selector.Scope.Mode == "named" {
			selector.Scope.Value = "prod"
		}
		if seed%4 == 0 {
			selector.Policy = "current-only"
			selector.Scope = identity.EnvironmentScope{Mode: "named", Value: "prod"}
		}
		q := exploration.Query{Signal: []string{"spans", "logs"}[seed%2], EnvironmentSelector: selector, FromMS: 1000, ToMS: 61000, Limit: 1 + seed%20, Offset: seed % 1001, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}, BucketMS: 1000, Where: &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "run.id"}, Op: "eq", Value: fmt.Sprint(seed)}, {Field: &exploration.Field{Kind: "attribute", Key: "backup.id"}, Op: "eq", Value: "quoted'\\雪"}}}}
		other := q
		other.FromMS = 121000
		other.ToMS = 241000
		other.Limit = 100
		other.Offset = (seed + 1) % 1001
		values := []*string{nil, textPointer292(""), textPointer292("shared'\\😀"), textPointer292("side" + fmt.Sprint(seed))}
		makeSide := func(question exploration.Query, side int) exploration.Cohort {
			groups := []exploration.ValueCount{}
			for i, v := range values {
				if i == 3 && side == seed%2 {
					continue
				}
				groups = append(groups, exploration.ValueCount{Value: v, Count: "1"})
			}
			count := len(groups)
			state := "observed"
			if variant == 1 && side == 0 {
				groups = []exploration.ValueCount{}
				count = 0
				state = "no_data"
			}
			if variant == 2 && side == 1 {
				state = "changed_between_reads"
				count++
			}
			return exploration.Cohort{Question: question, Basis: fmt.Sprintf("side%d-%d", seed, side), CapturedAt: "2026-10-31T00:00:00Z", Count: fmt.Sprint(count), State: state, Groups: groups}
		}
		request := exploration.ComparisonRequest{Baseline: makeSide(q, 0), Selection: makeSide(other, 1)}
		raw, _ := json.Marshal(request)
		result, err := exploration.CompareCohorts(request)
		if err != nil || result.GroupComparison == nil {
			tc.Fatal(err)
		}
		cases := []map[string]any{}
		g := result.GroupComparison
		for _, side := range []string{"baseline", "selection"} {
			snapshot := g.Baseline
			if side == "selection" {
				snapshot = g.Selection
			}
			for _, row := range g.Rows {
				intent := exploration.GroupSideIntent{Type: "scope-group-side", Side: side, Key: row.Key, DatasetID: snapshot.ID, Revision: snapshot.Revision, ScopeID: snapshot.Scope.ID, Generation: snapshot.Scope.Generation}
				for bad := 0; bad < 5; bad++ {
					candidate := intent
					switch bad {
					case 1:
						candidate.Side = map[string]string{"baseline": "selection", "selection": "baseline"}[side]
					case 2:
						candidate.Key = "1unknown"
					case 3:
						candidate.Revision = "stale"
					case 4:
						candidate.Generation = "old"
					}
					out, e := exploration.StageGroupSide(request, candidate)
					cases = append(cases, map[string]any{"intent": candidate, "accepted": e == nil, "query": out, "side": side, "key": row.Key, "mutation": bad})
				}
			}
		}
		after, _ := json.Marshal(request)
		if string(raw) != string(after) {
			tc.Fatal("planner mutated request")
		}
		if err := runNativeHistogramJS(map[string]any{"request": request, "result": result, "cases": cases}, groupSideParityJS); err != nil {
			tc.Fatal(err)
		}
	}, options(100)...)
}
func textPointer292(value string) *string { return &value }

func TestGroupSidePredicateBudgetRefusesWidening292(t *testing.T) {
	leaf := exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "run.id"}, Op: "eq", Value: "budget"}
	where := leaf
	for range 4 {
		where = exploration.Predicate{All: []exploration.Predicate{where}}
	}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}, Where: &where}
	c := exploration.Cohort{Question: q, Basis: "budget", CapturedAt: "2026-10-31T00:00:00Z", Count: "1", State: "observed", Groups: []exploration.ValueCount{{Value: textPointer292("shared"), Count: "1"}}}
	req := exploration.ComparisonRequest{Baseline: c, Selection: c}
	result, err := exploration.CompareCohorts(req)
	if err != nil {
		t.Fatal(err)
	}
	side := result.GroupComparison.Baseline
	intent := exploration.GroupSideIntent{Type: "scope-group-side", Side: "baseline", Key: "1shared", DatasetID: side.ID, Revision: side.Revision, ScopeID: side.Scope.ID, Generation: side.Scope.Generation}
	if _, err = exploration.StageGroupSide(req, intent); err == nil {
		t.Fatal("overflowing new predicate accepted")
	}
	if err = runNativeHistogramJS(map[string]any{"request": req, "result": result, "cases": []map[string]any{{"intent": intent, "accepted": false, "side": "baseline", "key": "1shared", "mutation": 0}}}, groupSideParityJS); err != nil {
		t.Fatal(err)
	}
}
