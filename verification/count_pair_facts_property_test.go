package verification

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// Actual comparison DTOs exercise the shared supplied-facts display boundary;
// arithmetic, acquisition and witness admission remain their existing owners.
func TestHegelCountPairSuppliedFacts323(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		variant := hegel.Draw(tc, hegel.Integers(0, 3))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"named", "empty", "unreported"}))
		scope := identity.EnvironmentScope{Mode: mode}
		if mode == "named" {
			scope.Value = fmt.Sprintf("prod-%d'雪", seed%9)
		}
		selection, e := identity.EffectiveEnvironment(scope)
		if e != nil {
			tc.Fatal(e)
			return
		}
		selector, _ := selection.Selector()
		q := exploration.Query{Signal: []string{"logs", "spans"}[seed%2], EnvironmentSelector: &selector, FromMS: 1790000000000 + int64(seed)*1000, ToMS: 1790000060000 + int64(seed)*1000, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}}
		other := q
		other.FromMS += 60000
		other.ToMS += 60000
		makeSide := func(q exploration.Query, side int) exploration.Cohort {
			count := strconv.Itoa(seed%99 + 1 + side)
			state := "observed"
			if variant == 1 && side == 0 {
				count = "0"
				state = "no_data"
			}
			groups := []exploration.ValueCount{}
			if count != "0" {
				value := "same category"
				groups = append(groups, exploration.ValueCount{Value: &value, Count: count})
			}
			if variant == 2 && side == 0 {
				state = "changed_between_reads"
				count = strconv.Itoa(seed%99 + 2)
			}
			if variant == 3 {
				count = "18446744073709551615"
				groups[0].Count = count
			}
			return exploration.Cohort{Question: q, Basis: fmt.Sprintf("basis-%d-%d", seed, side), CapturedAt: []string{"2026-11-04T01:00:00.123456789Z", "2026-11-04T01:00:01.987654321Z"}[side], Count: count, State: state, Groups: groups}
		}
		request := exploration.ComparisonRequest{Baseline: makeSide(q, 0), Selection: makeSide(other, 1)}
		result, e := exploration.CompareCohorts(request)
		if e != nil || result.ViewComparison == nil || result.GroupComparison == nil {
			tc.Fatal("actual comparison unavailable", e)
			return
		}
		if e = runNativeHistogramJS(map[string]any{"request": request, "result": result}, countPairFacts323JS); e != nil {
			tc.Fatal(e)
			return
		}
	}, options(30)...)
}

const countPairFacts323JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import * as V from '../ui/comparison-visual.mjs';import * as E from '../ui/evidence-scope.mjs';import * as S from '../ui/source-snapshots.mjs';
const {request,result}=JSON.parse(fs.readFileSync(0,'utf8')),admitted=S.importComparisonResult(result,request);
const sides=['baseline','selection'].map((key,i)=>({label:i?'Selected window':'Previous window',data:E.comparisonContextSide(request[key].question,{mode:i?'fixed':'frozen',state:request[key].state,count:request[key].count,readAt:request[key].capturedAt,basis:request[key].basis,...(!i?{capturedAt:'2026-11-04T02:00:00.111222333Z'}:{})})}));
const input={comparison:admitted.viewComparison,sides},before=JSON.stringify(input),facts=V.countPairFacts(input);assert.equal(facts.valid,true);
for(const [i,key]of ['baseline','selection'].entries()){const expected=request[key],side=facts.sides[i];assert.equal(side.label,sides[i].label);assert.equal(side.count,expected.count);assert.equal(side.fromMs,expected.question.fromMs);assert.equal(side.toMs,expected.question.toMs);assert.equal(side.readAt,expected.capturedAt);assert.deepEqual(side.data.details,sides[i].data.details);}
assert.equal(facts.sides[0].data.details.find(d=>d.label==='Stored custody time').value,'2026-11-04T02:00:00.111222333Z');assert.equal(facts.sides[0].count,request.baseline.count);assert.equal(JSON.stringify(input),before);facts.sides[0].data.details[0].value='changed';facts.sides[0].label='changed';assert.equal(JSON.stringify(input),before);
for(const [index,change] of [x=>x.sides.pop(),x=>x.sides[0].label='',x=>x.sides[0].label='x'.repeat(121),x=>x.sides[0].data.state='healthy',x=>x.comparison.baseline.scope.selector.fromMs=-1,x=>x.comparison.baseline.scope.selector.toMs=Number.MAX_SAFE_INTEGER,x=>x.comparison.baseline.rows[0].values.count='-1',x=>x.comparison.baseline.rows[0].values.count='01',x=>x.comparison.baseline.capturedAt='not-a-clock'].entries()){const bad=structuredClone(input);change(bad);assert.equal(V.countPairFacts(bad).valid,false,'negative mutation '+index);}
const absent=structuredClone(input);delete absent.comparison.baseline.capturedAt;const unknown=V.countPairFacts(absent);assert.equal(unknown.valid,true);assert.equal(unknown.sides[0].readAt,null);assert.equal(unknown.sides[0].count,request.baseline.count);assert.equal(JSON.stringify(input),before);
`
