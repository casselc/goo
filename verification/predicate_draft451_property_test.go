package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

func TestHegelPredicateDraft451State(t *testing.T) {
	type action struct {
		Kind  string                 `json:"kind"`
		Where *exploration.Predicate `json:"where,omitempty"`
	}
	type snapshot struct {
		Applied, Draft *exploration.Predicate
		Error          *string
		Pending, Dirty bool
		Revision       int
	}
	type observation struct {
		State       snapshot
		Accepted    *bool
		CandidateOK bool
	}
	type fixture struct {
		Actions  []action      `json:"actions"`
		Expected []observation `json:"-"`
	}
	fixtures := []fixture{}
	clone := func(p *exploration.Predicate) *exploration.Predicate {
		if p == nil {
			return nil
		}
		raw, _ := json.Marshal(p)
		var copy exploration.Predicate
		_ = json.Unmarshal(raw, &copy)
		return &copy
	}
	valid := func(p *exploration.Predicate) bool {
		return (exploration.Query{Signal: "logs", Environment: "ordinary", FromMS: 0, ToMS: 60000, Limit: 100, Where: p}).Validate() == nil
	}
	hegel.Test(t, func(tc *hegel.T) {
		f := fixture{}
		state := snapshot{}
		for range 25 {
			kind := hegel.Draw(tc, hegel.Integers(0, 8))
			n := hegel.Draw(tc, hegel.Integers(0, 100000))
			p := &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: fmt.Sprint(n)}
			a := action{Kind: []string{"stage", "error", "pending", "invalid", "accept", "stale", "discard", "retire", "reset"}[kind], Where: p}
			var accepted *bool
			switch a.Kind {
			case "stage":
				state.Draft = clone(p)
				state.Error = nil
				state.Pending = false
				state.Revision++
			case "error":
				state.Draft = clone(p)
				message := "invalid draft"
				state.Error = &message
				state.Pending = false
				state.Revision++
			case "pending":
				state.Draft = clone(p)
				state.Error = nil
				state.Pending = true
				state.Revision++
			case "invalid":
				a.Where.Field.Key = "not_a_column"
				state.Draft = clone(a.Where)
				message := "invalid predicate"
				state.Error = &message
				state.Pending = false
				state.Revision++
			case "accept", "stale":
				ok := a.Kind == "accept" && state.Error == nil && !state.Pending && valid(state.Draft)
				accepted = &ok
				if ok {
					state.Applied = clone(state.Draft)
					state.Error = nil
					state.Pending = false
					state.Revision++
				}
			case "discard":
				state.Draft = clone(state.Applied)
				state.Error = nil
				state.Pending = false
				state.Revision++
			case "retire":
				state.Revision++
			case "reset":
				state.Applied = clone(p)
				state.Draft = clone(p)
				state.Error = nil
				state.Pending = false
				state.Revision++
			}
			state.Dirty = state.Error != nil || state.Pending || !reflect.DeepEqual(state.Applied, state.Draft)
			observed := state
			if state.Error != nil {
				message := "draft error"
				observed.Error = &message
			}
			observed.Applied = clone(state.Applied)
			observed.Draft = clone(state.Draft)
			f.Actions = append(f.Actions, a)
			f.Expected = append(f.Expected, observation{observed, accepted, state.Error == nil && !state.Pending && valid(state.Draft)})
		}
		fixtures = append(fixtures, f)
	}, options(50)...)
	script := `import fs from 'node:fs';import{PredicateDraft}from'../ui/predicate-draft.mjs';import{validate}from'../ui/predicate-validation.mjs';globalThis.fetch=()=>{throw Error('pure draft attempted IO')};const out=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const state=new PredicateDraft({validate:p=>validate(p,'logs')}),rows=[];for(const a of f.actions){let accepted;switch(a.kind){case'stage':state.stage(a.where);break;case'error':state.stage(a.where,{error:'invalid draft'});break;case'pending':state.stage(a.where,{pending:true});break;case'invalid':state.stage(a.where);break;case'accept':accepted=state.accept(state.token());break;case'stale':accepted=state.accept(state.token()-1);break;case'discard':state.discard();break;case'retire':state.retire();break;case'reset':state.reset(a.where);break;}let candidateOK=true;try{state.candidate()}catch{candidateOK=false}const before=JSON.stringify(state.snapshot()),snapshot=state.snapshot();const record=structuredClone(snapshot);if(record.error!=null)record.error='draft error';rows.push({state:record,...(accepted===undefined?{}:{accepted}),candidateOK});if(snapshot.draft?.field)snapshot.draft.field.key='caller-mutated';if(JSON.stringify(state.snapshot())!==before)throw Error('snapshot aliases state');if(a.where?.field)a.where.field.key='caller-mutated';if(JSON.stringify(state.snapshot())!==before)throw Error('caller input aliases state')}out.push(rows)}process.stdout.write(JSON.stringify(out));`
	raw, _ := json.Marshal(fixtures)
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	output, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("draft batch: %v %s", e, output)
	}
	var actual [][]observation
	if e = json.Unmarshal(output, &actual); e != nil {
		t.Fatal(e, string(output))
	}
	if len(actual) != len(fixtures) {
		t.Fatal("lost draft sequence")
	}
	for i, f := range fixtures {
		if !reflect.DeepEqual(actual[i], f.Expected) {
			for j, want := range f.Expected {
				if !reflect.DeepEqual(actual[i][j], want) {
					a, _ := json.Marshal(actual[i][j])
					w, _ := json.Marshal(want)
					t.Fatalf("sequence%d step%d %s actual%s expected%s", i, j, f.Actions[j].Kind, a, w)
				}
			}
		}
	}
	t.Logf("%d generated draft sequences; %d independently checked transitions", len(fixtures), len(fixtures)*25)
}
