package verification

import (
	"bufio"
	"fmt"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestHegelRecordPresentationProperties(t *testing.T) {
	script := `const assert=require('node:assert/strict'),{ancestry,describe}=require('../ui/record-presentation.mjs'),rl=require('node:readline').createInterface({input:process.stdin});rl.on('line',line=>{try{let n=Number(JSON.parse(line).action),seed=n;const draw=max=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return seed%max},id=i=>i.toString(16).padStart(16,'0'),rows=Array.from({length:draw(30)+1},(_,i)=>({TraceId:(draw(3)?'1':'2').repeat(32),SpanId:id(draw(20)+1),ParentSpanId:draw(4)?id(draw(20)+1):'',SpanName:'work'})),selected=rows[draw(rows.length)],snapshot=JSON.stringify(rows),result=ancestry(selected,rows);assert.deepEqual(ancestry(selected,[...rows].reverse()),result);assert.deepEqual(ancestry(selected,rows.filter(r=>r.TraceId===selected.TraceId)),result);assert.equal(new Set(result.path).size,result.path.length);assert.ok(result.path.length<=8);for(const span of result.path)assert.ok(rows.some(r=>r.TraceId===selected.TraceId&&r.SpanId===span));for(let i=1;i<result.path.length;i++)assert.ok(rows.some(r=>r.TraceId===selected.TraceId&&r.SpanId===result.path[i]&&r.ParentSpanId===result.path[i-1]));result.path.push('tampered');assert.equal(JSON.stringify(rows),snapshot);const ns=(BigInt(n)*10000000000n+123456789n).toString(),record={timestamp_ns:ns,duration_ns:'0',StatusCode:'Unset',SpanName:'work'},view=describe(record,'spans');assert.ok(view.time.title.includes(ns));assert.equal(view.duration,'0.000 ms');assert.equal(view.label,'Unset');console.log(JSON.stringify({count:rows.length}))}catch(e){console.log(JSON.stringify({error:e.message}))}});`
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
