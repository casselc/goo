package verification

import (
	"bufio"
	"fmt"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestHegelDefinitionLibraryProperties(t *testing.T) {
	script := `const assert=require('node:assert/strict'),p=require('../ui/definition-library.mjs'),rl=require('node:readline').createInterface({input:process.stdin});rl.on('line',line=>{try{let seed=Number(JSON.parse(line).action);const draw=max=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return seed%max},n=draw(180),items=Array.from({length:n},(_,i)=>({family:['measurement','record','document'][i%3],revision:BigInt(i).toString(16).padStart(64,'0'),search:draw(3)?'west literal':'east',title:'same title'})),before=JSON.stringify(items),query=draw(2)?'west':'',size=draw(50)+1,expected=items.filter(x=>x.search.includes(query)),found=[];for(let offset=0;offset<expected.length;offset+=size){const v=p.page(items,{query,offset,size});assert.equal(v.matched,expected.length);assert.equal(v.total,n);assert.ok(v.items.length<=size);found.push(...v.items.map(p.key))}assert.deepEqual(found,expected.map(p.key));assert.equal(new Set(found).size,found.length);assert.equal(JSON.stringify(items),before);assert.equal(p.page(items,{query:'missing',offset:1000}).offset,0);if(n)assert.notEqual(p.key(items[0]),p.key({...items[0],family:'record'}));console.log(JSON.stringify({count:n}))}catch(e){console.log(JSON.stringify({error:e.message}))}});`
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
