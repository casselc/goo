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
	"hegel.dev/go/hegel"
)

// Generate the actual compiler shape: parallel linked_traces, linked_spans,
// and linked_attributes string-map arrays. Go independently admits each plan.
func TestHegelSpanLink452Projection(t *testing.T) {
	type fixture struct {
		Origin      exploration.Query `json:"origin"`
		Rows        []map[string]any  `json:"rows"`
		RowIndex    int               `json:"rowIndex"`
		LinkIndex   int               `json:"linkIndex"`
		Trace, Span string            `json:"-"`
		Copies      int               `json:"-"`
	}
	fixtures := []fixture{}
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		selected := hegel.Draw(tc, hegel.Integers(0, 1))
		copies := hegel.Draw(tc, hegel.Integers(0, 2))
		start := int64(1700000000000 + hegel.Draw(tc, hegel.Integers(0, 1000000)))
		width := int64(hegel.Draw(tc, hegel.Integers(1, 7200000)))
		traces := []string{fmt.Sprintf("%032x", n), fmt.Sprintf("%032x", n+1)}
		spans := []string{fmt.Sprintf("%016x", n+2), fmt.Sprintf("%016x", n+3)}
		q := exploration.Query{Signal: "spans", Environment: fmt.Sprintf("env-%d", n), FromMS: start, ToMS: start + width, Limit: 17, Offset: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: fmt.Sprint(n)}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}}}}
		if mode > 0 {
			q.Environment = ""
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprint(n)}}
			if mode == 1 {
				q.EnvironmentSelector.Policy = "current-only"
			}
			if mode == 3 {
				q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			}
			if mode == 4 {
				q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		if err := q.Validate(); err != nil {
			t.Fatal(err)
		}
		rows := []map[string]any{{"TraceId": fmt.Sprintf("%032x", n+100), "SpanId": fmt.Sprintf("%016x", n+101), "linked_traces": traces, "linked_spans": spans, "linked_attributes": []map[string]string{{"messaging.message.id": fmt.Sprint(n)}, {"empty": ""}}}}
		for i := 0; i < copies; i++ {
			rows = append(rows, map[string]any{"TraceId": traces[selected], "SpanId": spans[selected]})
		}
		rows = append(rows, map[string]any{"TraceId": traces[0], "SpanId": spans[1]}, map[string]any{"TraceId": traces[1], "SpanId": spans[0]})
		fixtures = append(fixtures, fixture{q, rows, 0, selected, traces[selected], spans[selected], copies})
	}, options(60)...)
	wire, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {plan,inspect} from '../ui/span-link-intent.mjs';globalThis.fetch=()=>{throw Error('pure planner attempted IO')};const results=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(f),out=plan(f),local=inspect(f);if(JSON.stringify(local.local)!==JSON.stringify(out.local))throw Error('passive admission differs');if(JSON.stringify(f)!==before)throw Error('input mutation');results.push(structuredClone(out));out.question.where.all[0].value='mutated';out.origin.question.where.all[0].value='mutated';out.attributes.changed='mutated';out.local.rowIndexes.push(999);if(JSON.stringify(f)!==before)throw Error('aliased output');for(const mutate of [x=>x.rows[0].linked_traces[x.linkIndex]='0'.repeat(32),x=>x.rows[0].linked_spans[x.linkIndex]='A'.repeat(16),x=>x.rows[0].linked_spans[x.linkIndex]='',x=>x.rows[0].linked_attributes.pop(),x=>x.rows[0].linked_attributes[x.linkIndex]={invalid:1},x=>x.rows[0].linked_traces.pop()]){const bad=structuredClone(f);mutate(bad);let refused=false;try{plan(bad)}catch{refused=true}if(!refused)throw Error('invalid admitted')}}process.stdout.write(JSON.stringify(results));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(wire)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node planner: %v %s", err, out)
	}
	var plans []struct {
		Question exploration.Query
		Local    struct {
			State      string
			RowIndexes []int
		}
		Target struct {
			TraceID string
			SpanID  string
		}
		Origin struct{ Question exploration.Query }
	}
	if err := json.Unmarshal(out, &plans); err != nil {
		t.Fatal(err, string(out))
	}
	if len(plans) != len(fixtures) {
		t.Fatal("lost fixture")
	}
	for i, f := range fixtures {
		p := plans[i]
		if err := p.Question.Validate(); err != nil {
			t.Fatal(err)
		}
		expected := exploration.Query{Signal: "spans", Environment: f.Origin.Environment, EnvironmentSelector: f.Origin.EnvironmentSelector, FromMS: f.Origin.FromMS, ToMS: f.Origin.ToMS, Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: f.Trace}, {Field: &exploration.Field{Kind: "column", Key: "span_id"}, Op: "eq", Value: f.Span}}}}
		if !reflect.DeepEqual(p.Question, expected) || !reflect.DeepEqual(p.Origin.Question, f.Origin) {
			t.Fatalf("scope changed fixture %d", i)
		}
		if p.Local.State != []string{"absent", "unique", "ambiguous"}[f.Copies] || len(p.Local.RowIndexes) != f.Copies || p.Target.TraceID != f.Trace || p.Target.SpanID != f.Span {
			t.Fatalf("tuple membership fixture %d", i)
		}
		for j, index := range p.Local.RowIndexes {
			if index != j+1 {
				t.Fatal("crossed tuple selected")
			}
		}
	}
	t.Logf("%d Hegel compiler-shaped link banks; exact Go/JS scope, crossed tuples, ambiguity, six refusal controls each, detached outputs and passive admission", len(fixtures))
}
