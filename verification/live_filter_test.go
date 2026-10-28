package verification

import (
	"bufio"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

type liveFilterMachine struct{ worker journalMachine }

func (m *liveFilterMachine) RuleInclude(tc hegel.TestCase)                 { m.worker.call("include") }
func (m *liveFilterMachine) RuleExclude(tc hegel.TestCase)                 { m.worker.call("exclude") }
func (m *liveFilterMachine) RuleRemove(tc hegel.TestCase)                  { m.worker.call("remove") }
func (m *liveFilterMachine) RuleApply(tc hegel.TestCase)                   { m.worker.call("apply") }
func (m *liveFilterMachine) RuleClear(tc hegel.TestCase)                   { m.worker.call("clear") }
func (m *liveFilterMachine) InvariantAppliedScopeStable(tc hegel.TestCase) { m.worker.call("inspect") }
func TestHegelLiveFilterStateSwarm(t *testing.T) {
	script := `const{ScopeFilters}=require('../ui/stream-entry'),display=require('../ui/field-presentation.mjs'),identity=require('../ui/question-session').identity,rl=require('node:readline').createInterface({input:process.stdin});let s=new ScopeFilters(),applied=[],steps=0;rl.on('line',line=>{try{const{action}=JSON.parse(line);if(action==='reset'){s=new ScopeFilters();applied=[];steps=0}if(action==='include'||action==='exclude'){const f={kind:steps%2?'attribute':'resource',key:'region'},v=steps%3?'west':'';s.stage(f,v,action);steps++}if(action==='remove')s.remove(0);if(action==='apply'){s.apply();applied=structuredClone(s.draft)}if(action==='clear'){s.reset();applied=[]}if(JSON.stringify(s.applied)!==JSON.stringify(applied))throw Error('draft mutation changed applied selector');if(s.draft.length>16)throw Error('unbounded filter state');const seen=new Set();for(const p of s.draft){const key=JSON.stringify([p.field.kind,p.field.key,p.value]);if(seen.has(key)||typeof p.value!=='string')throw Error('duplicate or coerced filter');seen.add(key);const before=JSON.stringify(p);display.parts(p);display.selectorText(p);if(before!==JSON.stringify(p))throw Error('display mutated selector');if(p.value===''){const omitted={...p};delete omitted.value;if(identity({where:p})!==identity({where:omitted}))throw Error('empty wire default changed identity')}}const view=s.where(true);if(view)view.all[0].value='tampered';if(JSON.stringify(s.applied)!==JSON.stringify(applied))throw Error('mutable snapshot');console.log(JSON.stringify({count:s.draft.length}))}catch(e){console.log(JSON.stringify({error:e.message}))}});`
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
	scanner := bufio.NewScanner(out)
	hegel.Test(t, func(tc *hegel.T) {
		m := &liveFilterMachine{worker: journalMachine{in: in, out: scanner}}
		m.worker.call("reset")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24), hegel.WithAlwaysCheckInvariants("InvariantAppliedScopeStable"))
	}, options(80)...)
}
