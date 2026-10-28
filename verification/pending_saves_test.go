package verification

import (
	"bufio"
	"encoding/json"
	"hegel.dev/go/hegel"
	"io"
	"os/exec"
	"testing"
)

type journalMachine struct {
	in    io.Writer
	out   *bufio.Scanner
	count int
}

func (m *journalMachine) call(action string) int {
	raw, _ := json.Marshal(map[string]string{"action": action})
	if _, e := m.in.Write(append(raw, '\n')); e != nil {
		panic(e)
	}
	if !m.out.Scan() {
		panic("journal worker exited")
	}
	var result struct {
		Count int
		Error string
	}
	if e := json.Unmarshal(m.out.Bytes(), &result); e != nil || result.Error != "" {
		panic("journal property failed: " + string(m.out.Bytes()))
	}
	return result.Count
}
func (m *journalMachine) RuleStage(tc hegel.TestCase) {
	if m.count < 16 {
		m.count++
		if m.call("stage") != m.count {
			panic("stage did not persist")
		}
	}
}
func (m *journalMachine) RuleReload(tc hegel.TestCase) {
	if m.call("reload") != m.count {
		panic("reload changed pending requests")
	}
}
func (m *journalMachine) RuleRetry(tc hegel.TestCase) {
	if m.call("retry") != m.count {
		panic("retry duplicated intent")
	}
}
func (m *journalMachine) RuleAcknowledge(tc hegel.TestCase) {
	if m.count > 0 {
		m.count--
	}
	if m.call("ack") != m.count {
		panic("ack changed another request")
	}
}
func (m *journalMachine) RuleRejectChangedIntent(tc hegel.TestCase) {
	if m.call("changed") != m.count {
		panic("changed intent overwrote retry")
	}
}
func (m *journalMachine) InvariantStableJournal(tc hegel.TestCase) {
	if m.call("inspect") != m.count {
		panic("journal count diverged")
	}
}
func TestHegelPendingSaveStateSwarm(t *testing.T) {
	script := `const {Journal}=require('../ui/pending-saves.mjs'),rl=require('node:readline').createInterface({input:process.stdin});let values=new Map(),s={getItem:k=>values.get(k)||null,setItem:(k,v)=>values.set(k,v)},j=new Journal(s),sequence=0,expected=new Map();rl.on('line',line=>{try{const {action}=JSON.parse(line);if(action==='reset'){values=new Map();j=new Journal(s);expected.clear();sequence=0}if(action==='stage'){const key=(++sequence).toString(16).padStart(64,'0'),body={requestID:key,resultToken:'a'.repeat(64),indexes:[0],tags:['review'],note:'draft '+sequence};j.begin('',body);expected.set(key,JSON.stringify(body));body.note='mutated caller'}if(action==='reload')j=new Journal(s);if(action==='retry')for(const e of j.list())j.begin(e.path,e.body);if(action==='ack'){const key=expected.keys().next().value;if(key){j.complete(key);expected.delete(key)}}if(action==='changed')for(const e of j.list()){let rejected=false;try{j.begin(e.path,{...e.body,note:'foreign'})}catch{rejected=true}if(!rejected)throw Error('changed intent accepted')}const actual=j.list();if(actual.length!==expected.size||actual.some(e=>JSON.stringify(e.body)!==expected.get(e.body.requestID)))throw Error('persisted intent mutated');console.log(JSON.stringify({count:actual.length}))}catch(e){console.log(JSON.stringify({error:e.message}))}});`
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
		m := &journalMachine{in: in, out: scanner}
		m.call("reset")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(18), hegel.WithAlwaysCheckInvariants("InvariantStableJournal"))
	}, options(60)...)
}
