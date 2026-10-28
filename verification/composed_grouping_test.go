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

type composedGroupingMachine struct{ worker journalMachine }

func (m *composedGroupingMachine) RuleAddSpan(tc hegel.TestCase)     { m.worker.call("span") }
func (m *composedGroupingMachine) RuleAddLog(tc hegel.TestCase)      { m.worker.call("log") }
func (m *composedGroupingMachine) RuleAddResource(tc hegel.TestCase) { m.worker.call("resource") }
func (m *composedGroupingMachine) RuleRemoveFirst(tc hegel.TestCase) { m.worker.call("first") }
func (m *composedGroupingMachine) RuleRemoveLast(tc hegel.TestCase)  { m.worker.call("last") }
func (m *composedGroupingMachine) RuleReplace(tc hegel.TestCase)     { m.worker.call("replace") }
func (m *composedGroupingMachine) RuleDetached(tc hegel.TestCase)    { m.worker.call("detach") }
func (m *composedGroupingMachine) InvariantScope(tc hegel.TestCase)  { m.worker.call("inspect") }
func TestHegelComposedGroupingStateSwarm(t *testing.T) {
	script := `const a=require('node:assert/strict'),{groupQueries}=require('../ui/population-workspace.mjs'),rl=require('node:readline').createInterface({input:process.stdin}),original=['spans','logs'].map(signal=>({signal,environment:'prod',fromMs:1000,toMs:61000,offset:20,limit:20,bucketMs:1000,groupTimeSeries:true,where:{any:[{field:{kind:'attribute',key:'tenant'},op:'eq'},{field:{kind:'resource',key:'tenant'},op:'missing'}]}})),frozen=JSON.stringify(original);let state,expected,changed;function reset(){state=structuredClone(original);expected=[[],[]];changed=[false,false]}reset();rl.on('line',line=>{try{const{action}=JSON.parse(line);if(action==='reset')reset();if(['span','log','resource','replace','first','last'].includes(action)){const i=action==='log'?1:0,kind=action==='resource'?'resource':'attribute',field={kind,key:action==='log'?'messaging.destination.name':'http.method'},mode=action==='replace'?'replace':['first','last'].includes(action)?'remove':'add',picked=mode==='remove'?(expected[i][action==='first'?0:expected[i].length-1]||field):field,previous=structuredClone(state),duplicate=expected[i].some(f=>f.kind===picked.kind&&f.key===picked.key),refused=mode==='add'&&(duplicate||expected[i].length===2)||mode==='remove'&&!duplicate;if(refused)a.throws(()=>groupQueries(state,original[i].signal,picked,mode));else{state=groupQueries(state,original[i].signal,picked,mode);if(mode==='replace')expected[i]=[structuredClone(picked)];if(mode==='add')expected[i].push(structuredClone(picked));if(mode==='remove')expected[i]=expected[i].filter(f=>f.kind!==picked.kind||f.key!==picked.key);changed[i]=true}a.deepEqual(state[1-i],previous[1-i]);picked.key='callback mutation'}if(action==='detach'){const copy=groupQueries(state,'logs',{kind:'attribute',key:'opaque'});copy[0].where.any[0].field.key='mutated'}a.equal(JSON.stringify(original),frozen);state.forEach((q,i)=>{a.deepEqual(q.groupBy||[],expected[i]);a.ok(expected[i].length<=2);a.deepEqual(q.where,original[i].where);for(const k of ['signal','environment','fromMs','toMs','limit','bucketMs'])a.equal(q[k],original[i][k]);a.equal(q.offset,changed[i]?0:20);a.equal(!!q.groupTimeSeries,!changed[i])});console.log(JSON.stringify({count:0}))}catch(e){console.log(JSON.stringify({error:e.stack}))}});`
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
		m := &composedGroupingMachine{worker: journalMachine{in: in, out: scanner}}
		m.worker.call("reset")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(24), hegel.WithAlwaysCheckInvariants("InvariantScope"))
	}, options(100)...)
}

func TestQuintComposedGroupingLedger(t *testing.T) {
	file := os.Getenv("GOO_COMPOSED_GROUPING_LEDGER")
	if file == "" {
		t.Skip("live browser ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action string
				Reads  int
			}
			Checks []struct {
				Parent, Child exploration.Query
				Fields        []exploration.Field
				Companion     bool
				Groups        []struct {
					G0, G1 *string
					Count  string
				}
			}
			Errors []string
			Oracle []struct {
				Route         *string
				Method, Count string
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[1].Reads++
		case 1:
			l.Checks[0].Child.FromMS++
		case 2:
			l.Checks[0].Child.GroupBy[1].Key = "http.request.method"
		case 3:
			l.Checks[3].Child.Environment = "other"
		case 4:
			l.Checks[1].Child.Where = nil
		case 5:
			l.Checks[1].Child.GroupTimeSeries = true
		case 6:
			l.Checks[0].Child.GroupBy = append(l.Checks[0].Child.GroupBy, l.Checks[0].Child.GroupBy[0])
		case 7:
			l.Checks[1].Child.GroupBy[0], l.Checks[1].Child.GroupBy[1] = l.Checks[1].Child.GroupBy[1], l.Checks[1].Child.GroupBy[0]
		case 8:
			l.Checks[2].Groups[0].Count = "99"
		}
		coherent := len(l.Errors) == 0 && len(l.Checks) == 4 && len(l.Oracle) == 3 && len(l.Events) > 0 && l.Events[len(l.Events)-1].Reads == 7
		for _, c := range l.Checks {
			expected := c.Parent
			if !c.Companion {
				expected.GroupBy = c.Fields
				expected.GroupTimeSeries = false
				expected.Offset = 0
			}
			coherent = coherent && c.Child.Validate() == nil && reflect.DeepEqual(expected, c.Child)
			if c.Companion {
				continue
			}
			coherent = coherent && len(c.Fields) == 2 && len(c.Groups) == 3
			for _, g := range c.Groups {
				found := false
				for _, o := range l.Oracle {
					values := map[string]*string{"http.route": o.Route, "http.method": &o.Method}
					if reflect.DeepEqual(g.G0, values[c.Fields[0].Key]) && reflect.DeepEqual(g.G1, values[c.Fields[1].Key]) && g.Count == o.Count {
						found = true
					}
				}
				coherent = coherent && found
			}
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
		t.Fatal("observed composed grouping trace rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatalf("mutation %d admitted", i)
		}
	}
}
