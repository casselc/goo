package verification

import (
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

func TestHegelHeatmapRectangle(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(0, 7))
		b := hegel.Draw(tc, hegel.Integers(a, 7))
		c := hegel.Draw(tc, hegel.Integers(0, 2))
		d := hegel.Draw(tc, hegel.Integers(c, 2))
		script := `const assert=require('node:assert/strict'),{question,rectangle}=require('../ui/population-selection.mjs');const [a,b,c,d]=process.argv.slice(1).map(Number),base={signal:'spans',environment:'test',fromMs:1000,toMs:3500,limit:1,offset:0,groupBy:[{kind:'column',key:'duration_bucket'}],where:{field:{kind:'attribute',key:'route'},op:'eq',value:''}},data={question:base,timeSeries:[{fromMs:1000,toMs:2000},{fromMs:2000,toMs:3000},{fromMs:3000,toMs:3500}]},before=JSON.stringify(data),r=rectangle(data,{fromColumn:c,toColumn:d,fromRow:a,toRow:b}),q=question(base,r);assert.equal(q.fromMs,data.timeSeries[c].fromMs);assert.equal(q.toMs,data.timeSeries[d].toMs);assert.deepEqual(q.where.all[0],base.where);assert.equal(JSON.stringify(data),before);const bounds=[0n,1000000n,5000000n,10000000n,50000000n,100000000n,500000000n,1000000000n];assert.equal(q.where.all[1].value,String(bounds[a]));if(b<7)assert.equal(q.where.all[2].value,String(bounds[b+1]-1n));else assert.equal(q.where.all.length,2);`
		if out, err := exec.Command("node", "-e", script, strconv.Itoa(a), strconv.Itoa(b), strconv.Itoa(c), strconv.Itoa(d)).CombinedOutput(); err != nil {
			tc.Fatal(err, string(out))
		}
	}, options(40)...)
}
