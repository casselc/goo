package verification

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

type observedGroupingMachine struct{ worker journalMachine }

func (m *observedGroupingMachine) RuleSpan(tc hegel.TestCase)       { m.worker.call("spans") }
func (m *observedGroupingMachine) RuleLog(tc hegel.TestCase)        { m.worker.call("logs") }
func (m *observedGroupingMachine) RuleLegacy(tc hegel.TestCase)     { m.worker.call("legacy") }
func (m *observedGroupingMachine) RuleDetached(tc hegel.TestCase)   { m.worker.call("detach") }
func (m *observedGroupingMachine) RuleUnknown(tc hegel.TestCase)    { m.worker.call("unknown") }
func (m *observedGroupingMachine) InvariantScope(tc hegel.TestCase) { m.worker.call("inspect") }
func TestHegelObservedGroupingStateSwarm(t *testing.T) {
	script := `const a=require('node:assert/strict'),{groupQueries}=require('../ui/population-workspace'),rl=require('node:readline').createInterface({input:process.stdin});const original=['spans','logs'].map(signal=>({signal,environment:'prod',fromMs:1000,toMs:61000,limit:20,offset:20,where:{any:[{field:{kind:'attribute',key:'tenant'},op:'eq',value:''},{field:{kind:'resource',key:'tenant'},op:'missing'}]},groupBy:[{kind:'column',key:'service'}],groupTimeSeries:true,bucketMs:1000})),frozen=JSON.stringify(original);let state=structuredClone(original),expected=[null,null];rl.on('line',line=>{try{const{action}=JSON.parse(line);if(action==='reset'){state=structuredClone(original);expected=[null,null]}if(['spans','logs','legacy'].includes(action)){const i=action==='logs'?1:0,field={kind:'attribute',key:action==='legacy'?'http.method':i?'messaging.destination.name':'http.route'},previous=structuredClone(state);state=groupQueries(state,original[i].signal,field);a.deepEqual(state[1-i],previous[1-i]);expected[i]=field.key;field.key='caller mutation'}if(action==='detach'){const copy=structuredClone(state);copy[0].where.any[0].value='mutated';const out=groupQueries(state,'logs',{kind:'attribute',key:'opaque'});out[0].where.any[0].value='mutated'}if(action==='unknown')a.throws(()=>groupQueries(state,'metrics',{kind:'attribute',key:'x'}));a.equal(JSON.stringify(original),frozen);state.forEach((q,i)=>{a.deepEqual(q.where,original[i].where);for(const key of ['signal','environment','fromMs','toMs','limit','bucketMs'])a.equal(q[key],original[i][key]);a.equal(q.offset,expected[i]===null?20:0);if(expected[i]!==null){a.deepEqual(q.groupBy,[{kind:'attribute',key:expected[i]}]);a.equal(Object.hasOwn(q,'groupTimeSeries'),false)}});console.log(JSON.stringify({count:0}))}catch(e){console.log(JSON.stringify({error:e.stack}))}});`
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
		m := &observedGroupingMachine{worker: journalMachine{in: in, out: scanner}}
		m.worker.call("reset")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24), hegel.WithAlwaysCheckInvariants("InvariantScope"))
	}, options(80)...)
}

func TestQuintObservedGroupingLedger(t *testing.T) {
	name := os.Getenv("GOO_OBSERVED_GROUPING_LEDGER")
	if name == "" {
		t.Skip("live browser ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Pairs []struct {
				Parent, Child exploration.Query
				Companion     bool
			}
			Events []struct {
				Action string
				Reads  int
			}
			Errors []string
			Oracle []struct {
				Value *string
				Count string
			}
			Groups []struct {
				G0    *string
				Count string
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[1].Reads++
		case 1:
			l.Pairs[0].Child.FromMS++
		case 2:
			l.Pairs[1].Child.GroupBy[0].Key = "http.method"
		case 3:
			l.Pairs[3].Child.GroupBy = nil
		case 4:
			l.Pairs[1].Child.GroupTimeSeries = true
		case 5:
			l.Pairs[1].Child.Where = nil
		}
		coherent := len(l.Pairs) == 4 && len(l.Errors) == 0 && len(l.Groups) == 2 && len(l.Oracle) == 2 && len(l.Events) > 0 && l.Events[len(l.Events)-1].Reads == 7
		for _, pair := range l.Pairs {
			expected := pair.Parent
			if !pair.Companion {
				expected.GroupBy = []exploration.Field{{Kind: "attribute", Key: "http.route"}}
				expected.GroupTimeSeries = false
				expected.Offset = 0
			}
			coherent = coherent && pair.Child.Validate() == nil && reflect.DeepEqual(expected, pair.Child)
		}
		for _, group := range l.Groups {
			found := false
			for _, o := range l.Oracle {
				if reflect.DeepEqual(group.G0, o.Value) && group.Count == o.Count {
					found = true
				}
			}
			coherent = coherent && found
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, ev := range l.Events {
			rec.Record(context.Background(), ev.Action, nil, "obs.reads", ev.Reads, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/question_links.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("observed grouping trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatalf("mutation %d admitted", i)
		}
	}
}
