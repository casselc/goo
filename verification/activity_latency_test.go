package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelActivityLatencyOwnership(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		script := `const {ReadGate,query,scopeKey}=require('../ui/activity-latency'),assert=require('node:assert/strict');let seed=Number(process.argv[1]);const g=new ReadGate(),tickets=[];let scope=null,frozen=false,pending=null,serial=0;for(let i=0;i<60;i++){seed=(seed*1664525+1013904223)>>>0;const op=seed%6;if(op===0){const s=seed%4?'a':'b';if(s!==scope){scope=s;frozen=false;pending=null;serial++}g.bind(s)}else if(op===1){const token=g.start();if(scope&&!frozen&&pending===null){pending=++serial;tickets.push({token,scope});assert.equal(token,pending)}else assert.equal(token,null)}else if(op===2){g.pause();frozen=true;pending=null;serial++}else if(op===3){g.resume();frozen=false;pending=null;serial++}else if(tickets.length){const ticket=op===4?tickets[0]:tickets.at(-1),expected=pending!==null&&ticket.token===pending&&ticket.scope===scope&&!frozen;assert.equal(g.accept(ticket.token,ticket.scope),expected);if(expected)pending=null}assert.equal(g.generation,serial);assert.equal(g.loading,pending!==null)}const context={selector:{signal:'spans',environment:'test',where:{field:{kind:'attribute',key:'owner'},op:'missing'}},readWindow:{fromMs:1700000000000+seed,toMs:1700000060000+seed}};const q=query(context);assert.equal(scopeKey(context),scopeKey({...context,readWindow:{fromMs:context.readWindow.fromMs+1000,toMs:context.readWindow.toMs+1000}}));process.stdout.write(JSON.stringify(q));`
		raw, e := exec.Command("node", "-e", script, strconv.Itoa(seed)).CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(raw))
		}
		var q exploration.Query
		if e = json.Unmarshal(raw, &q); e != nil {
			tc.Fatal(e)
		}
		if _, e = exploration.Compile(q); e != nil {
			tc.Fatal(e)
		}
		if q.ToMS-q.FromMS != 60000 || q.Limit != 1 {
			tc.Fatal("window/budget drift")
		}
	}, options(60)...)
}

func TestQuintActivityLatencyLedger(t *testing.T) {
	name := os.Getenv("GOO_ACTIVITY_LATENCY_LEDGER")
	if name == "" {
		t.Skip("observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l struct {
		First, Next                                                    exploration.Query
		Pair                                                           json.RawMessage
		FirstRead, BeforeBrush, BeforeReturn, AfterReturn, StagedReads int
		Scope                                                          struct {
			Signal, Environment string
			Where               *exploration.Predicate
		}
	}
	if e = json.Unmarshal(raw, &l); e != nil {
		t.Fatal(e)
	}
	if l.FirstRead != 1 || l.BeforeBrush != 2 || l.StagedReads != l.BeforeBrush || l.Next.ToMS <= l.First.ToMS || l.First.ToMS-l.First.FromMS != 60000 || l.Next.ToMS-l.Next.FromMS != 60000 || l.First.Environment != l.Scope.Environment || l.Next.Environment != l.Scope.Environment || !reflect.DeepEqual(l.First.Where, l.Scope.Where) || !reflect.DeepEqual(l.Next.Where, l.Scope.Where) {
		t.Fatal("rolling read/staging scope drift")
	}
	type event struct {
		Action       string `json:"action"`
		Depth, Reads int
	}
	events := []event{{"history", 0, l.BeforeBrush}, {"stage", 1, l.StagedReads}}
	for i := l.StagedReads + 1; i <= l.BeforeReturn; i++ {
		events = append(events, event{"read", 1, i})
	}
	events = append(events, event{"back", 0, l.AfterReturn})
	replay := map[string]any{"parent": l.Next, "pair": l.Pair, "initial": l.BeforeBrush, "beforeReturn": l.BeforeReturn, "afterReturn": l.AfterReturn, "events": events}
	body, e := json.Marshal(replay)
	if e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(t.TempDir(), "region-ledger.json")
	if e = os.WriteFile(file, body, 0600); e != nil {
		t.Fatal(e)
	}
	t.Setenv("GOO_ACTIVITY_REGION_REPLAY", file)
	checkRegionPartitionLedger(t, "GOO_ACTIVITY_REGION_REPLAY", true)
}
