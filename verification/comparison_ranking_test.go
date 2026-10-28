package verification

import (
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

func TestHegelComparisonRanking(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(1, 1000))
		b := hegel.Draw(tc, hegel.Integers(1, 1000))
		c := hegel.Draw(tc, hegel.Integers(1, 1000))
		script := `const assert=require('node:assert/strict'),{rankRows,compareRows}=require('../ui/comparison-breakdown.mjs');const seeds=process.argv.slice(1).map(Number),rows=seeds.map((n,i)=>{const x=String(n),y=String(n+i+1),d=(v,count)=>({state:'observed',count,sumNs:String(18446744073709551616n+BigInt(v)),p95Ns:String(v),p99Ns:String(v+1)});return {value:String(i),baselineCount:x,selectionCount:y,measures:{baselineDuration:d(n,x),selectionDuration:d(n+i,y),baselineStatus:{count:x,error:'1'},selectionStatus:{count:y,error:String(i)}}}}),before=JSON.stringify(rows),totals={baselineCount:String(seeds.reduce((a,b)=>a+b,0)),selectionCount:String(seeds.reduce((a,b)=>a+b,6))};for(const metric of ['share','count','mean','p95','p99','error-rate']){const ranked=rankRows(rows,metric,totals);assert.deepEqual(ranked,rankRows([...rows].reverse(),metric,totals));for(const a of rows)for(const b of rows)assert.equal(Math.sign(compareRows(a,b,metric,totals))+Math.sign(compareRows(b,a,metric,totals)),0);assert(compareRows(ranked[0],ranked[2],metric,totals)<=0)}assert.equal(JSON.stringify(rows),before);`
		if out, err := exec.Command("node", "-e", script, strconv.Itoa(a), strconv.Itoa(b), strconv.Itoa(c)).CombinedOutput(); err != nil {
			tc.Fatal(err, string(out))
		}
	}, options(40)...)
}
