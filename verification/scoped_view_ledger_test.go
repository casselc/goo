package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type scopedViewEvent struct {
	Action     string `json:"action"`
	Reads      int    `json:"reads"`
	Generation int    `json:"generation"`
}

// Replay observed source effects, not clicks. The browser supplies the source-read
// counter and harness ordinal of observed successful applied Runs after each
// operation has settled; this is not the opaque dataset scope-generation token.
func TestQuintScopedViewLedger(t *testing.T) {
	path := os.Getenv("GOO_SCOPED_LEDGER")
	if path == "" {
		t.Skip("observed scoped-view browser ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var events []scopedViewEvent
	if e = json.Unmarshal(raw, &events); e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	for _, event := range events {
		seen[event.Action] = true
	}
	if len(events) < 6 || events[0] != (scopedViewEvent{Action: "run", Reads: 1, Generation: 1}) {
		t.Fatal("missing initial observed run")
	}
	for _, action := range []string{"run", "switch", "select", "stage", "refuse"} {
		if !seen[action] {
			t.Fatalf("missing observed %s", action)
		}
	}
	if events[len(events)-1].Reads < 2 || events[len(events)-1].Generation < 2 {
		t.Fatal("missing explicitly rerun applied dataset")
	}
	spec, e := os.ReadFile("model/scoped_view_261.qnt")
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	if e = os.WriteFile(filepath.Join(dir, "scoped_view_261.qnt"), spec, 0600); e != nil {
		t.Fatal(e)
	}
	binding := `spec: scoped_view_261.qnt
module: scoped_view_261
constants: {LIMIT: '8'}
invariants: [safety]
actions:
  applyRun:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  switch:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  selectRow:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  stage:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  refuse:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
`
	bindingPath := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(bindingPath, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	check := func(trace []scopedViewEvent) bool {
		t.Helper()
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range trace {
			action := event.Action
			if action == "run" {
				action = "applyRun"
			}
			if action == "select" {
				action = "selectRow"
			}
			rec.Record(context.Background(), action, nil, "obs.reads", event.Reads, "obs.generation", event.Generation)
		}
		checker, e := validate.NewChecker(bindingPath, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(events) {
		t.Fatal("observed scope-view effects rejected")
	}
	// One fault at a time, including simultaneous illicit read/scope changes that
	// preserve the simple reads==generation invariant but violate action ownership.
	for _, action := range []string{"switch", "select", "stage", "refuse"} {
		index := -1
		for i, event := range events {
			if event.Action == action {
				index = i
				break
			}
		}
		for _, fault := range []string{"read", "generation", "both"} {
			changed := append([]scopedViewEvent(nil), events...)
			if fault != "generation" {
				changed[index].Reads++
			}
			if fault != "read" {
				changed[index].Generation++
			}
			if check(changed) {
				t.Fatalf("accepted illicit %s effect on %s", fault, action)
			}
		}
	}
}

// Generate actual public-contract operations rather than validating a second Go
// implementation of the same model. Run is host-supplied data replacement; every
// other operation calls the real JS contracts and must preserve the input basis.
func TestHegelScopedViewContractSequences(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(6, 30))
		actions := make([]string, count)
		names := []string{"switch", "select", "stage", "refuse", "run"}
		for i := range actions {
			actions[i] = names[hegel.Draw(tc, hegel.Integers(0, len(names)-1))]
		}
		wire, e := json.Marshal(actions)
		if e != nil {
			t.Fatal(e)
		}
		script := `const fs=require('node:fs'),assert=require('node:assert/strict'),v=require('../ui/view-model.mjs');
const actions=JSON.parse(fs.readFileSync(0,'utf8'));let generation=1;
const make=()=>({version:'goo.dataset/1',id:'ranked',revision:'r'+generation,scope:{id:'q'+generation,generation,signal:'spans',selector:{}},shape:'ranked',fields:[{key:'category',label:'Service',type:'string'},{key:'count',label:'Count',type:'number'}],rows:[{key:'key-'+generation,values:{category:'api',count:'9007199254740993'}}],quality:{extent:'unknown',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'}});
let snapshot=make();for(const action of actions){const before=JSON.stringify(snapshot),previous=snapshot;if(action==='run'){const old=v.selectionIntent(snapshot,snapshot.rows[0].key);generation++;snapshot=make();assert.notEqual(old.revision,snapshot.revision);assert.notEqual(old.scopeId,snapshot.scope.id);assert.notEqual(old.generation,snapshot.scope.generation);assert.throws(()=>v.selectionIntent(snapshot,old.key));continue}if(action==='switch'){for(const kind of v.eligibleKinds(snapshot))assert.equal(v.validateDisplay(snapshot,{kind,category:'category',value:'count'}).ok,true)}else if(action==='refuse'){assert.equal(v.validateDisplay(snapshot,{kind:'bar',category:'category',value:'count',requiresComplete:true}).ok,false)}else{const intent=action==='select'?v.selectionIntent(snapshot,snapshot.rows[0].key):v.scopeIntent(snapshot,'category','api');assert.equal(intent.datasetId,snapshot.id);assert.equal(intent.revision,snapshot.revision);assert.equal(intent.scopeId,snapshot.scope.id);assert.equal(intent.generation,generation)}assert.equal(snapshot,previous);assert.equal(JSON.stringify(snapshot),before)}process.stdout.write('ok');`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(wire)
		out, e := cmd.CombinedOutput()
		if e != nil || string(out) != "ok" {
			t.Fatalf("public scoped contracts: %v: %s", e, out)
		}
	}, options(30)...)
}
