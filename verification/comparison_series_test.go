package verification

import (
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

// Exercise the actual JS geometry with large counts and arbitrary window offsets.
func TestHegelComparisonSeriesGeometry(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(1000, 1000000))
		count := hegel.Draw(tc, hegel.Integers(1, 10000))
		width := hegel.Draw(tc, hegel.Integers(2, 60000))
		script := `const a=require('node:assert/strict'),{model,brush}=require('../ui/comparison-series');const [offset,count,width]=process.argv.slice(1).map(Number),big=18446744073709551616n*BigInt(count),d={sides:[0,1].map(i=>({label:String(i),fromMs:offset+i*width*2,toMs:offset+(i*2+1)*width,points:[{fromMs:offset+i*width*2,toMs:offset+(i*2+1)*width,count:String(big*BigInt(i+1)),state:'observed'}]}))},before=structuredClone(d),m=model(d);a.equal(m.ceiling,String(big*2n));a.equal(m.sides[0].points[0].height,50);a.equal(m.sides[1].points[0].height,100);a.equal(model({...d,ceiling:String(big)}).state,'overflow');const fixed=model({...d,ceiling:String(big*4n)});a.equal(fixed.sides[1].points[0].height,50);a.equal(fixed.ceiling,String(big*4n));for(const mode of ['utc','elapsed'])for(const side of [0,1])for(const [lo,hi]of [[0,1],[.2,.8],[.8,.2],[.5,.5]]){const r=brush(d,mode,side,lo,hi);if(r){a(r.fromMs>=d.sides[side].fromMs);a(r.toMs<=d.sides[side].toMs);a(r.fromMs<r.toMs);a.deepEqual(r,brush(d,mode,side,hi,lo))}}a.deepEqual(d,before);d.sides[1].points[0].state='unknown';const unknown=model(d);a.equal(unknown.ceiling,String(big));a.equal(unknown.sides[1].points[0].count,null);a.equal(brush(d,'elapsed',0,NaN,.5),null);`
		if out, e := exec.Command("node", "-e", script, strconv.Itoa(offset), strconv.Itoa(count), strconv.Itoa(width)).CombinedOutput(); e != nil {
			tc.Fatal(e, string(out))
		}
	}, options(40)...)
}
