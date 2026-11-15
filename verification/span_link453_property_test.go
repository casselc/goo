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

type link453Row struct {
	TraceID    string              `json:"TraceId"`
	SpanID     string              `json:"SpanId"`
	Traces     []string            `json:"linked_traces"`
	Spans      []string            `json:"linked_spans"`
	Attributes []map[string]string `json:"linked_attributes"`
}

// Pure Go/JS membership is compared per observation; duplicate link entries
// cannot inflate counts, while duplicate observations remain separate rows.
func TestHegelSpanLink453IncomingProjection(t *testing.T) {
	type fixture struct {
		Origin   exploration.Query `json:"origin"`
		Rows     []link453Row      `json:"rows"`
		RowIndex int               `json:"rowIndex"`
		Expected []bool            `json:"-"`
	}
	fixtures := []fixture{}
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		duplicates := hegel.Draw(tc, hegel.Integers(1, 3))
		copies := hegel.Draw(tc, hegel.Integers(1, 3))
		start := int64(1700000000000 + hegel.Draw(tc, hegel.Integers(0, 1000000)))
		width := int64(hegel.Draw(tc, hegel.Integers(1, 7200000)))
		target := exploration.SpanLinkTarget{TraceID: fmt.Sprintf("%032x", n), SpanID: fmt.Sprintf("%016x", n+1)}
		otherTrace, otherSpan := fmt.Sprintf("%032x", n+2), fmt.Sprintf("%016x", n+3)
		origin := exploration.Query{Signal: "spans", Environment: fmt.Sprintf("env-%d", n), FromMS: start, ToMS: start + width, Limit: 17, Offset: 100, Where: &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: fmt.Sprint(n)}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}}}}
		if mode > 0 {
			origin.Environment = ""
			origin.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprint(n)}}
			if mode == 1 {
				origin.EnvironmentSelector.Policy = "current-only"
			}
			if mode == 3 {
				origin.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			}
			if mode == 4 {
				origin.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		if err := origin.Validate(); err != nil {
			t.Fatal(err)
		}
		rows := []link453Row{{TraceID: target.TraceID, SpanID: target.SpanID}}
		linked := link453Row{TraceID: otherTrace, SpanID: otherSpan}
		for range duplicates {
			linked.Traces = append(linked.Traces, target.TraceID)
			linked.Spans = append(linked.Spans, target.SpanID)
			linked.Attributes = append(linked.Attributes, map[string]string{"message.id": fmt.Sprint(n), "empty": ""})
		}
		for range copies {
			rows = append(rows, linked)
		}
		rows = append(rows,
			link453Row{otherTrace, otherSpan, []string{target.TraceID, otherTrace}, []string{otherSpan, target.SpanID}, []map[string]string{{}, {}}}, // crossed, never a tuple
			link453Row{otherTrace, otherSpan, []string{target.TraceID}, []string{target.SpanID}, []map[string]string{nil}},
			link453Row{otherTrace, otherSpan, []string{target.TraceID}, []string{target.SpanID}, nil},
			link453Row{otherTrace, otherSpan, []string{target.TraceID}, nil, []map[string]string{{}}},
			link453Row{otherTrace, otherSpan, []string{"invalid", target.TraceID}, []string{"", target.SpanID}, []map[string]string{nil, {}}}) // unrelated bad position does not erase match
		want := make([]bool, len(rows))
		for i, row := range rows {
			if len(row.Traces) == len(row.Spans) && len(row.Traces) == len(row.Attributes) {
				for j, trace := range row.Traces {
					if trace == target.TraceID && row.Spans[j] == target.SpanID && row.Attributes[j] != nil {
						want[i] = true
					}
				}
			}
			if exploration.MatchSpanLinkTarget(&target, row.Traces, row.Spans, row.Attributes) != want[i] {
				t.Fatal("Go tuple membership differs from independent position oracle")
			}
		}
		fixtures = append(fixtures, fixture{origin, rows, 0, want})
	}, options(70)...)
	raw, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {planIncoming} from '../ui/span-link-intent.mjs';import {matches,validateTarget} from '../ui/span-link-membership.mjs';globalThis.fetch=()=>{throw Error('pure incoming planner attempted IO')};const result=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(f),p=planIncoming(f),membership=f.rows.map(row=>matches(row,p.target));result.push({question:structuredClone(p.question),origin:structuredClone(p.origin.question),membership});p.question.linksTo.traceId='changed';p.target.spanId='changed';p.origin.question.where.all[0].value='changed';if(JSON.stringify(f)!==before)throw Error('input/output alias');for(const value of ['',undefined,'0'.repeat(32),'A'.repeat(32),'a'.repeat(31)]){const bad=structuredClone(f);bad.rows[0].TraceId=value;let refused=false;try{planIncoming(bad)}catch{refused=true}if(!refused)throw Error('invalid selected target admitted')}const target={traceId:f.rows[0].TraceId,spanId:f.rows[0].SpanId};for(const bad of [{...target,spanId:''},{...target,extra:'x'},{traceId:target.traceId}]){let refused=false;try{validateTarget(bad)}catch{refused=true}if(!refused)throw Error('invalid target admitted')}for(const mutate of [r=>r.linked_attributes[0]=null,r=>r.linked_attributes[0]={invalid:1},r=>delete r.linked_spans[0]]){const row={linked_traces:[target.traceId],linked_spans:[target.spanId],linked_attributes:[{}]};mutate(row);if(matches(row,target))throw Error('malformed membership admitted')}}process.stdout.write(JSON.stringify(result));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node incoming planner: %v %s", err, out)
	}
	var results []struct {
		Question, Origin exploration.Query
		Membership       []bool
	}
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatal(err, string(out))
	}
	if len(results) != len(fixtures) {
		t.Fatal("lost fixture")
	}
	for i, f := range fixtures {
		r := results[i]
		target := &exploration.SpanLinkTarget{TraceID: f.Rows[0].TraceID, SpanID: f.Rows[0].SpanID}
		expected := exploration.Query{Signal: "spans", Environment: f.Origin.Environment, EnvironmentSelector: f.Origin.EnvironmentSelector, FromMS: f.Origin.FromMS, ToMS: f.Origin.ToMS, Limit: 100, LinksTo: target}
		if r.Question.Validate() != nil || !reflect.DeepEqual(r.Question, expected) || !reflect.DeepEqual(r.Origin, f.Origin) || !reflect.DeepEqual(r.Membership, f.Expected) {
			t.Fatalf("independent Go/JS projection or membership differs fixture %d", i)
		}
	}
	t.Logf("%d Hegel banks: independent same-index membership, crossed pairs, duplicate links/rows, malformed arrays/maps, exact scope projection, detached zero-IO plans", len(fixtures))
}
