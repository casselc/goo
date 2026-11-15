package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Cross-language admission is checked against ordinary Go queries and the real
// content-addressed configuration store. No browser or telemetry read is used.
func TestHegelPairedFilters450Projection(t *testing.T) {
	type fixture struct {
		Selector   workbench.ActivitySelector        `json:"selector"`
		Conditions map[string]*exploration.Predicate `json:"conditions"`
		Queries    []exploration.Query               `json:"-"`
	}
	fixtures := []fixture{}
	leaf := func(kind, key, op, value string) *exploration.Predicate {
		return &exploration.Predicate{Field: &exploration.Field{Kind: kind, Key: key}, Op: op, Value: value}
	}
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(0, 100000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		local := hegel.Draw(tc, hegel.Integers(0, 3))
		selector := workbench.ActivitySelector{Signal: "both", Environment: fmt.Sprintf("ordinary-%d", n), Where: leaf("resource", "fixture.bank", "eq", fmt.Sprint(n))}
		if mode > 0 {
			selector.Environment = ""
			selector.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprint(n)}}
			if mode == 1 {
				selector.EnvironmentSelector.Policy = "current-only"
			}
			if mode == 3 {
				selector.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			}
			if mode == 4 {
				selector.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		logs := &exploration.Predicate{All: []exploration.Predicate{*leaf("column", "severity_number", "gte", "17"), {Any: []exploration.Predicate{*leaf("attribute", "tenant", "missing", ""), *leaf("attribute", "tenant", "eq", "")}}}}
		spans := &exploration.Predicate{All: []exploration.Predicate{*leaf("column", "duration_ns", "gte", fmt.Sprint(100000000+n)), *leaf("column", "kind", "eq", "Server")}}
		if local == 1 {
			logs = nil
		}
		if local == 2 {
			spans = nil
		}
		if local == 3 {
			selector.Where = nil
		}
		v := workbench.ActivityView{Version: 1, Title: fmt.Sprintf("Lane predicates %d", n), Selector: selector, Lanes: &workbench.ActivityLanes{Logs: workbench.ActivityLayout{Columns: []string{"utc", "summary"}, Where: logs}, Spans: workbench.ActivityLayout{Columns: []string{"service", "summary", "utc"}, Where: spans}}}
		original := activity446Clone(t, v)
		store := workbench.Store{Dir: t.TempDir()}
		saved, err := store.CreateActivityView(v)
		if err != nil {
			t.Fatal(err)
		}
		qs, err := saved.Queries()
		if err != nil || len(qs) != 2 {
			t.Fatal("ordinary lane query projection", err)
		}
		for i, condition := range []*exploration.Predicate{logs, spans} {
			expected := selector.Where
			if condition != nil {
				expected = condition
				if selector.Where != nil {
					expected = &exploration.Predicate{All: []exploration.Predicate{*selector.Where, *condition}}
				}
			}
			if !reflect.DeepEqual(qs[i].Where, expected) || qs[i].Validate() != nil {
				t.Fatal("combined lane condition changed")
			}
		}
		selector.Signal = "logs"
		fixtures = append(fixtures, fixture{selector, map[string]*exploration.Predicate{"logs": logs, "spans": spans}, qs})
		// Local conditions bind identity and do not mutate the stored shared selector.
		changed := activity446Clone(t, original)
		changed.Lanes.Logs.Where = leaf("column", "severity_number", "gte", "21")
		if logs != nil {
			changed.Lanes.Logs.Where = leaf("column", "severity_number", "gte", "22")
		}
		altered, e := store.CreateActivityView(changed)
		if e != nil || altered.ID == saved.ID {
			t.Fatal("lane predicate omitted from content identity", e)
		}
		returned, _ := saved.Queries()
		returned[0].Where = leaf("column", "body", "contains", "caller-mutated")
		loaded, e := store.LoadActivityView(saved.ID)
		if e != nil {
			t.Fatal(e)
		}
		loaded.ID = ""
		if !reflect.DeepEqual(*loaded, original) {
			t.Fatal("caller query mutated saved lane definition")
		}
	}, options(40)...)
	raw, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {create} from '../ui/paired-activity-plan.mjs';globalThis.fetch=()=>{throw Error('pure planner attempted IO')};const out=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const conditions=Object.fromEntries(Object.entries(f.conditions).filter(([,v])=>v!==null));const before=JSON.stringify(f),plans=create(f.selector,conditions);if(JSON.stringify(f)!==before)throw Error('mutated input');out.push(structuredClone(plans));if(plans.logs.where)plans.logs.where.op='mutated';if(JSON.stringify(f)!==before)throw Error('aliased input');const bad=[{logs:{field:{kind:'column',key:'duration_ns'},op:'gte',value:'1'}},{spans:{field:{kind:'column',key:'body'},op:'contains',value:'x'}}];for(const c of bad){let refused=false;try{create(f.selector,c)}catch{refused=true}if(!refused)throw Error('cross-signal lane admitted')}}process.stdout.write(JSON.stringify(out));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("Node planner: %v %s", e, out)
	}
	var plans []map[string]exploration.TailSelector
	if e = json.Unmarshal(out, &plans); e != nil {
		t.Fatal(e, string(out))
	}
	if len(plans) != len(fixtures) {
		t.Fatal("lost generated fixture")
	}
	for i, f := range fixtures {
		for j, signal := range []string{"logs", "spans"} {
			q := f.Queries[j]
			actual := plans[i][signal].Query(q.ToMS)
			if !reflect.DeepEqual(actual, q) {
				t.Fatalf("Go/JS lane projection differs fixture%d %s: %#v vs %#v", i, signal, actual, q)
			}
		}
	}
	t.Logf("%d generated saved pairs, %d independent Go/JS lane projections", len(fixtures), len(fixtures)*2)
}
