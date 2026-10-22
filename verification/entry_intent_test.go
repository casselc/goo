package verification

import (
	"bufio"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

type entryIntentMachine struct{ worker journalMachine }

func (m *entryIntentMachine) RuleVisible(tc hegel.TestCase)                { m.worker.call("visible") }
func (m *entryIntentMachine) RuleHidden(tc hegel.TestCase)                 { m.worker.call("hidden") }
func (m *entryIntentMachine) RuleLeave(tc hegel.TestCase)                  { m.worker.call("leave") }
func (m *entryIntentMachine) RuleMeasurements(tc hegel.TestCase)           { m.worker.call("measurements") }
func (m *entryIntentMachine) RulePause(tc hegel.TestCase)                  { m.worker.call("pause") }
func (m *entryIntentMachine) RuleExplicit(tc hegel.TestCase)               { m.worker.call("explicit") }
func (m *entryIntentMachine) RuleStoredVisit(tc hegel.TestCase)            { m.worker.call("stored") }
func (m *entryIntentMachine) RuleFreshVisit(tc hegel.TestCase)             { m.worker.call("fresh") }
func (m *entryIntentMachine) InvariantNoImplicitRestart(tc hegel.TestCase) { m.worker.call("inspect") }
func TestHegelEntryIntentStateSwarm(t *testing.T) {
	script := `const{EntryGate}=require('../ui/stream-entry'),rl=require('node:readline').createInterface({input:process.stdin});let g,claims=0,allowed=false,used=false;function reset(fresh){g=new EntryGate(fresh?'':'#step=held');allowed=fresh;used=false;claims=0}reset(true);rl.on('line',line=>{try{const{action}=JSON.parse(line);if(action==='reset'||action==='fresh')reset(true);if(action==='stored')reset(false);if(['visible','hidden','leave','measurements'].includes(action)){const c={active:action!=='leave',visible:action!=='hidden',records:action!=='measurements'},result=g.claim(c);if(result){if(!allowed||used||!c.active||!c.visible||!c.records)throw Error('implicit source claim outside fresh visible records intent');used=true;claims++}else if(action==='visible'&&allowed&&!used)throw Error('fresh visible entry failed to claim')}if(action==='pause'){g.pause();if(used)allowed=false}if(action==='explicit'){g.explicit();used=true;allowed=false}if(claims>1)throw Error('duplicate automatic claim');console.log(JSON.stringify({count:claims}))}catch(e){console.log(JSON.stringify({error:e.message}))}});`
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
		m := &entryIntentMachine{worker: journalMachine{in: in, out: scanner}}
		m.worker.call("reset")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24), hegel.WithAlwaysCheckInvariants("InvariantNoImplicitRestart"))
	}, options(80)...)
}
