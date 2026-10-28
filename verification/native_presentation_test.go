package verification

import (
	"bufio"
	"fmt"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestHegelNativePresentationProperties(t *testing.T) {
	script := `const assert=require('node:assert/strict'),p=require('../ui/native-presentation.mjs'),rl=require('node:readline').createInterface({input:process.stdin});rl.on('line',line=>{try{let seed=Number(JSON.parse(line).action);const draw=max=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return seed%max},n=draw(30)+1,base=1790000000000000000n,intervals=Array.from({length:n},(_,i)=>({fromNS:String(base+BigInt(i)),toNS:String(base+BigInt(i+1)),state:draw(4)?'ok':'unknown',perSecond:draw(3),reason:''}));for(const x of intervals)if(x.state!=='ok')x.reason='counter-decrease';const stream={name:'requests',unit:'1',resource:{'service.namespace':''},attributes:{}},s={stream,state:intervals.some(x=>x.state!=='ok')?'unknown':'ok',intervals},before=JSON.stringify(s),last=intervals[n-1],view=p.latest(s,'counter'),reverse=p.latest({...s,intervals:[...intervals].reverse()},'counter');assert.deepEqual(view,reverse);assert.equal(view.toNS,last.toNS);assert.equal(view.value,last.state==='ok'?last.perSecond+' 1/s':'Unavailable');const d=p.streamDiagnostics(s,'counter');assert.equal(d.recorded+d.review,n);assert.equal(d.review,intervals.filter(x=>x.state!=='ok').length);assert.equal(JSON.stringify(s),before);view.stream.resource['service.namespace']='changed';assert.equal(stream.resource['service.namespace'],'');assert.equal(p.windowCoverage(null).validated,false);assert.equal(p.windowCoverage({state:'unknown',reasons:[]}).validated,false);const delta=BigInt(draw(100000)),den=BigInt(draw(100)+1),summary={state:'observed',reasons:[],delta:String(delta),rate:delta+'/'+den},snapshot=JSON.stringify(summary);assert.equal(p.windowValue(stream,summary),delta+' increments · '+p.rateText(summary.rate)+' 1/s');assert.equal(p.windowValue(stream,{...summary,state:'unknown'}),'Unavailable');assert.equal(p.windowValue(stream,{...summary,reasons:['missing-coverage']}),'Unavailable');assert.equal(JSON.stringify(summary),snapshot);console.log(JSON.stringify({count:n}))}catch(e){console.log(JSON.stringify({error:e.message}))}});`
	cmd := exec.Command("node", "-e", script)
	in, e := cmd.StdinPipe()
	if e != nil {
		t.Fatal(e)
	}
	out, e := cmd.StdoutPipe()
	if e != nil {
		t.Fatal(e)
	}
	if e = cmd.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() { in.Close(); cmd.Wait() }()
	worker := journalMachine{in: in, out: bufio.NewScanner(out)}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 2147483647))
		worker.call(fmt.Sprint(seed))
	}, options(200)...)
}
