package verification

import (
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

func histogramComparisonFixture(typ string, scale, seed, mode int) (*exploration.ViewHistogramComparison, error) {
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.amount"}, Type: typ, Unit: "items", Parse: "canonical-decimal"}
	if typ == "decimal64" {
		binding.Scale = scale
	}
	cohorts := make([]exploration.Cohort, 2)
	for side := range cohorts {
		values := []*string{nil}
		for _, v := range []string{"", "bad", "0", "1", "2", "9223372036854775807"} {
			text := v
			if typ == "decimal64" && v != "" && v != "bad" {
				text = decText(v, scale)
			}
			values = append(values, &text)
		}
		for i := 0; i < (seed+side*3)%11; i++ {
			v := "1"
			if typ == "decimal64" {
				v = decText(v, scale)
			}
			values = append(values, &v)
		}
		if side == 0 && mode == 1 {
			values = []*string{nil}
		}
		if side == 0 && mode == 2 || mode == 4 {
			values = nil
		}
		numeric, err := exploration.EvaluateNumericValues(binding, values)
		if err != nil {
			return nil, err
		}
		hist, err := exploration.EvaluateNumericHistogram(binding, values)
		if err != nil {
			return nil, err
		}
		state := "observed"
		if len(values) == 0 {
			state = "no_data"
		}
		if side == 0 && mode == 3 {
			state = "changed_between_reads"
		}
		cohorts[side] = exploration.Cohort{Question: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000 + int64(side), ToMS: 61000 + int64(side), Limit: 100, Numeric: &binding, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}}, Basis: "claimed-" + strconv.Itoa(side), State: state, Count: numeric.Count, Numeric: &numeric, NumericHistogram: &hist}
	}
	if seed%2 == 1 {
		cohorts[0], cohorts[1] = cohorts[1], cohorts[0]
	}
	result, err := exploration.CompareCohorts(exploration.ComparisonRequest{Distribution: true, Numeric: true, Baseline: cohorts[0], Selection: cohorts[1]})
	if err != nil {
		return nil, err
	}
	return result.HistogramComparison, nil
}

const histogramComparisonScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {asHistogramComparison,isHistogramComparison,hasGuarantee} from '../ui/snapshots.mjs';import {validateHistogramComparison} from '../ui/view-histogram-comparison.mjs';import {model,sideData} from '../ui/aligned-histogram.mjs';
const {comparison:raw,mutation=-1}=JSON.parse(fs.readFileSync(0,'utf8'));let reads=0;globalThis.fetch=()=>{reads++;throw Error('No display IO')};
if(mutation>=0){const c=structuredClone(raw);switch(mutation){case 0:c.selection.histogram.binding.unit='other';break;case 1:c.selection.scope.selector.environment='other';break;case 2:c.selection.scope.signal='logs';break;case 3:c.selection.rows.at(-1).values.to='0';break;case 4:c.selection.histogram.support.total='18446744073709551616';break;case 5:c.state='invalid';break;case 6:c.reason='other';break;case 7:c.selection.rows[0].values.count=9007199254740992;break;case 8:c.selection.rows=c.selection.rows.slice(1);break;case 9:c.selection.histogram.support.valid=String(BigInt(c.selection.histogram.support.valid)+1n);break}assert(!validateHistogramComparison(c).ok);assert.throws(()=>asHistogramComparison(c));process.stdout.write('ok');}
else {assert(validateHistogramComparison(raw).ok);const c=asHistogramComparison(raw),before=JSON.stringify(c);assert(isHistogramComparison(c));assert.equal(asHistogramComparison(c),c);assert(Object.isFrozen(c)&&Object.isFrozen(c.baseline.rows[0].values));for(const s of [c.baseline,c.selection]){assert.equal(s.origin,'browser-held');assert.equal(s.quality.extent,'unknown');assert(!hasGuarantee(s,'bins'));assert.equal(s.completeness,undefined)}
const fraction=(n,d)=>{if(d===0n)return null;let a=n,b=d;while(b){const t=a%b;a=b;b=t}return d/a===1n?String(n/a):n/a+'/'+(d/a)};
for(const scale of ['count','population-share','valid-share']){const m=model(c,{scale});assert(m);assert.equal(m.all.length,c.baseline.rows.length);assert.equal(m.ceiling,scale==='count'?m.ceiling:'100');assert.deepEqual(m.denominators,[c.baseline,c.selection].map(s=>s.histogram.support[scale==='valid-share'?'valid':'total']));for(const [i,row]of m.all.entries()){const counts=[c.baseline.rows[i].values.count,c.selection.rows[i].values.count];assert.deepEqual(row.counts,counts);assert.equal(row.difference,c.state==='diagnostic'?null:String(BigInt(counts[1])-BigInt(counts[0])));for(let side=0;side<2;side++){const expected=c.state==='diagnostic'?null:fraction(BigInt(counts[side])*100n,BigInt(m.denominators[side]));assert.equal(row.shares[side],expected);assert.equal(row.values[side],scale==='count'?counts[side]:expected)}}for(let side=0;side<2;side++)assert.equal(sideData(m,side).interactive,c.state!=='diagnostic')}
assert.equal(model(c,{scale:'pooled-share'}),null);assert.equal(model(c,{requiresComplete:true}),null);assert.equal(JSON.stringify(c),before);assert.equal(reads,0);process.stdout.write('ok');}
`

func TestHegelCanonicalHistogramComparisonOperands(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		typ := []string{"int64", "uint64", "decimal64"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		c, err := histogramComparisonFixture(typ, seed%9+1, seed, mode)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(nil, exploration.Query{}, map[string]any{"comparison": c}, histogramComparisonScript); err != nil {
			tc.Fatal(err)
		}
	}, options(60)...)
}
func TestHegelCanonicalHistogramComparisonMalformed(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mutation := hegel.Draw(tc, hegel.Integers(0, 9))
		c, err := histogramComparisonFixture("int64", 0, seed, 0)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(nil, exploration.Query{}, map[string]any{"comparison": c, "mutation": mutation}, histogramComparisonScript); err != nil {
			tc.Fatal(err)
		}
	}, options(40)...)
}
