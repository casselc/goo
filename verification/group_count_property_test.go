package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

const groupCountParityJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import {validateGroupComparison} from '../ui/group-count-comparison.mjs';import * as S from '../ui/source-snapshots.mjs';const {request,result}=JSON.parse(fs.readFileSync(0,'utf8')),g=result.groupComparison;assert.equal(validateGroupComparison(g).ok,true);const imported=S.importComparisonResult(result,request);assert.ok(imported.groupComparison);const mutable=structuredClone(result),detached=S.importComparisonResult(mutable,request),length=detached.groupComparison.rows.length;mutable.groupComparison.rows=[];assert.equal(detached.groupComparison.rows.length,length);assert.equal(Object.isFrozen(detached.groupComparison),true);const key=v=>v===null?'0':'1'+v,values=[...new Set([...request.baseline.groups,...request.selection.groups].map(x=>key(x.g0)))];assert.equal(g.rows.length,values.length);assert.deepEqual(new Set(g.rows.map(r=>r.key)),new Set(values));for(const side of ['baseline','selection']){const c=request[side],diagnostic=c.state==='changed_between_reads',map=new Map(c.groups.map(x=>[key(x.g0),x.count]));assert.equal(g[side+'Accounting'],diagnostic?'diagnostic':'coherent');for(const row of g.rows){const cell=row[side];if(diagnostic){assert.equal(cell.presence,map.has(row.key)?'present':'unknown');assert.equal(cell.state,'unavailable');assert.equal(cell.value,null)}else{assert.equal(cell.presence,map.has(row.key)?'present':'absent');assert.equal(cell.state,'derived');assert.equal(cell.value,map.get(row.key)||'0')}}}if(g.state==='diagnostic'){for(const r of g.rows){assert.equal(r.change.state,'unavailable');assert.equal(r.change.value,null)}}else{assert.equal(g.rows.reduce((n,r)=>n+BigInt(r.change.value),0n),BigInt(request.selection.count)-BigInt(request.baseline.count));for(const r of g.rows)assert.equal(BigInt(r.change.value),BigInt(r.selection.value)-BigInt(r.baseline.value))}const reject=bad=>assert.equal(validateGroupComparison(bad).ok,false);if(g.rows.length){let b=structuredClone(g);b.rows[0].key+='x';reject(b);b=structuredClone(g);b.rows.push(structuredClone(b.rows[0]));reject(b);b=structuredClone(g);b.rows[0].change={state:'derived',value:'18446744073709551616',unit:'records'};reject(b);b=structuredClone(g);b.rows[0].value=b.rows[0].value===null?'':null;reject(b);if(g.state!=='diagnostic'){b=structuredClone(g);b.rows[0].change.value=String(BigInt(b.rows[0].change.value)+1n);reject(b);b=structuredClone(g);b.rows.splice(0,1);reject(b)}else{const side=g.baselineAccounting==='diagnostic'?'baseline':'selection';b=structuredClone(g);b.rows[0][side]={presence:'absent',state:'derived',value:'0'};reject(b)}}if(g.rows.length){const forged=structuredClone(result),row=forged.groupComparison.rows[0];row.value='forged-category-291';row.key='1'+row.value;assert.equal(validateGroupComparison(forged.groupComparison).ok,true);assert.throws(()=>S.importComparisonResult(forged,request));}let bad=structuredClone(result);bad.groupComparison.baseline.capturedAt='2026-10-30T00:00:00Z';assert.throws(()=>S.importComparisonResult(bad,request));bad=structuredClone(result);bad.groupComparison.extra='unchecked';assert.throws(()=>S.importComparisonResult(bad,request));`

func TestHegelGroupCountCanonicalUnionAndAbsence(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 5))
		q := exploration.Query{Signal: []string{"logs", "spans"}[seed%2], Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}}
		other := q
		other.FromMS = 121000
		other.ToMS = 241000
		other.Limit = 20
		other.Offset = seed % 1001
		categories := []*string{nil, textPointer291(""), textPointer291("Missing value"), textPointer291("shared"), textPointer291("\ue000"), textPointer291("😀"), textPointer291(strings.Repeat("x", 1200))}
		makeSide := func(question exploration.Query, side int) exploration.Cohort {
			groups := []exploration.ValueCount{}
			total := new(big.Int)
			for i, v := range categories {
				if (seed+i+side)%3 == 0 {
					continue
				}
				n := fmt.Sprint(1 + (seed*(i+1)+side)%1000)
				groups = append(groups, exploration.ValueCount{Value: v, Count: n})
				x, _ := new(big.Int).SetString(n, 10)
				total.Add(total, x)
			}
			if variant == 1 && side == 0 {
				groups = []exploration.ValueCount{}
				total.SetInt64(0)
			}
			if variant == 5 {
				n := "18446744073709551615"
				if side == seed%2 {
					n = "0"
				}
				groups = []exploration.ValueCount{}
				if n != "0" {
					groups = append(groups, exploration.ValueCount{Value: textPointer291("max"), Count: n})
				}
				total.SetString(n, 10)
			}
			state := "observed"
			if total.Sign() == 0 {
				state = "no_data"
			}
			if variant == 2 && side == 1 || variant == 3 && side == 0 || variant == 4 {
				state = "changed_between_reads"
				total.Add(total, big.NewInt(1)) // Actual diagnostic accounting mismatch, not just a flag.
			}
			return exploration.Cohort{Question: question, Basis: fmt.Sprintf("side-%d-%d", seed, side), CapturedAt: "2026-10-31T00:00:00Z", Count: total.String(), State: state, Groups: groups}
		}
		request := exploration.ComparisonRequest{Baseline: makeSide(q, 0), Selection: makeSide(other, 1)}
		before, _ := json.Marshal(request)
		result, err := exploration.CompareCohorts(request)
		if err != nil || result.GroupComparison == nil {
			tc.Fatalf("canonical group result unavailable: %v", err)
		}
		after, _ := json.Marshal(request)
		if string(before) != string(after) {
			tc.Fatal("comparison mutated source cohorts")
		}
		permuted := request
		permuted.Baseline.Groups = append([]exploration.ValueCount{}, request.Baseline.Groups...)
		permuted.Selection.Groups = append([]exploration.ValueCount{}, request.Selection.Groups...)
		for _, groups := range [][]exploration.ValueCount{permuted.Baseline.Groups, permuted.Selection.Groups} {
			for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
		otherResult, permutationErr := exploration.CompareCohorts(permuted)
		aGroup, _ := json.Marshal(result.GroupComparison)
		var bGroup []byte
		if otherResult != nil {
			bGroup, _ = json.Marshal(otherResult.GroupComparison)
		}
		if permutationErr != nil || string(aGroup) != string(bGroup) {
			tc.Fatal("group input order changed canonical held comparison", permutationErr)
		}
		if err := runNativeHistogramJS(map[string]any{"request": request, "result": result}, groupCountParityJS); err != nil {
			tc.Fatal(err)
		}
	}, options(100)...)
}
func textPointer291(v string) *string { return &v }

func TestGroupCountFullUnionBoundary291(t *testing.T) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}}
	makeSide := func(prefix string) exploration.Cohort {
		groups := make([]exploration.ValueCount, 1000)
		for i := range groups {
			groups[i] = exploration.ValueCount{Value: textPointer291(fmt.Sprintf("%s-%04d", prefix, i)), Count: "1"}
		}
		return exploration.Cohort{Question: q, Basis: prefix, CapturedAt: "2026-10-31T00:00:00Z", Count: "1000", State: "observed", Groups: groups}
	}
	request := exploration.ComparisonRequest{Baseline: makeSide("baseline"), Selection: makeSide("current")}
	request.Selection.Question.FromMS = 121000
	request.Selection.Question.ToMS = 181000
	result, err := exploration.CompareCohorts(request)
	if err != nil || result.GroupComparison == nil || len(result.GroupComparison.Rows) != 2000 {
		t.Fatalf("full union unavailable: %v", err)
	}
	if err := runNativeHistogramJS(map[string]any{"request": request, "result": result}, groupCountParityJS); err != nil {
		t.Fatal(err)
	}
}

func TestGroupCountBothEmpty291(t *testing.T) {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}}
	empty := func(basis string) exploration.Cohort {
		return exploration.Cohort{Question: q, Basis: basis, CapturedAt: "2026-10-31T00:00:00Z", Count: "0", State: "no_data", Groups: []exploration.ValueCount{}}
	}
	request := exploration.ComparisonRequest{Baseline: empty("baseline"), Selection: empty("current")}
	request.Selection.Question.FromMS = 121000
	request.Selection.Question.ToMS = 181000
	result, err := exploration.CompareCohorts(request)
	if err != nil || result.GroupComparison == nil || len(result.GroupComparison.Rows) != 0 {
		t.Fatalf("known empty union unavailable: %v", err)
	}
	if err := runNativeHistogramJS(map[string]any{"request": request, "result": result}, groupCountParityJS); err != nil {
		t.Fatal(err)
	}
}
