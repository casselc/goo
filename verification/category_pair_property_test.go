package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os/exec"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"hegel.dev/go/hegel"
)

type categoryCell338 struct {
	State string  `json:"state"`
	Count *string `json:"count"`
}
type categoryRow338 struct {
	Value    *string         `json:"value"`
	Selected categoryCell338 `json:"selected"`
	Baseline categoryCell338 `json:"baseline"`
}
type categoryProjection338 struct {
	Rows     []categoryRow338 `json:"rows"`
	Max      string           `json:"max"`
	Complete bool             `json:"complete"`
}

func categoryKey338(value *string) string {
	if value == nil {
		return "0"
	}
	return "1" + *value
}
func categoryValues338(side map[string]any) map[string]string {
	out := map[string]string{}
	if side["state"] != "ready" {
		return out
	}
	for _, item := range side["items"].([]map[string]any) {
		key := "0"
		if item["value"] != nil {
			key = "1" + item["value"].(string)
		}
		out[key] = item["count"].(string)
	}
	return out
}
func categorySide338(seed, side int, field map[string]string) map[string]any {
	state := []string{"ready", "ready", "ready", "error", "loading", "unloaded"}[(seed/3+side)%6]
	if state != "ready" {
		return map[string]any{"state": state, "error": "independent source unavailable"}
	}
	values := []any{nil, "", "null", "0", "雪👀", "__proto__", "1", "literal\\ntext"}
	items := []map[string]any{}
	var total uint64
	for i, value := range values {
		if (seed+i*7+side)%3 == 0 {
			continue
		}
		n := uint64(1 + (seed+i)%3)
		total += n
		items = append(items, map[string]any{"value": value, "count": strconv.FormatUint(n, 10)})
	}
	own := []string{"selected-only", "baseline-only"}[side]
	items = append(items, map[string]any{"value": own, "count": "1"})
	total++
	if seed%5 == 0 {
		n := uint64(18446744073709551580)
		items[0]["count"] = strconv.FormatUint(n, 10)
		old, _ := strconv.ParseUint(strconv.Itoa(1+(seed+firstCategoryIndex338(seed, side))%3), 10, 64)
		total = total - old + n
	}
	if seed%11 == 0 {
		items = []map[string]any{}
		total = 0
	}
	return map[string]any{"state": "ready", "field": field, "items": items, "total": strconv.FormatUint(total, 10), "basis": strings.Repeat("c", 64), "readAt": fmt.Sprintf("2026-11-05T12:00:%02dZ", side+3), "reads": map[string]any{"count": map[string]string{"readAt": fmt.Sprintf("2026-11-05T12:00:%02dZ", side), "basis": strings.Repeat("a", 64)}, "groups": map[string]string{"readAt": fmt.Sprintf("2026-11-05T12:00:%02dZ", side+1), "basis": strings.Repeat("b", 64)}}}
}
func firstCategoryIndex338(seed, side int) int {
	for i := range 8 {
		if (seed+i*7+side)%3 != 0 {
			return i
		}
	}
	panic("no generated category")
}

// The independent Go oracle checks exact membership and counts emitted by the
// actual browser projection. A failed/loading/unloaded side is never empty data.
func TestHegelCategoryPairProjection338(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		field := map[string]string{"kind": []string{"attribute", "resource"}[seed%2], "key": []string{"http.request.method", "cloud.region"}[seed%2]}
		selected, baseline := categorySide338(seed, 0, field), categorySide338(seed, 1, field)
		input := map[string]any{"selected": selected, "baseline": baseline, "field": field}
		raw, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		cmd := exec.Command("node", "--input-type=module", "-e", categoryPairJS338)
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var result categoryProjection338
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		sides := []map[string]any{selected, baseline}
		members := []map[string]string{categoryValues338(selected), categoryValues338(baseline)}
		union := map[string]bool{}
		var maximum uint64
		for _, m := range members {
			for key, count := range m {
				union[key] = true
				n, _ := strconv.ParseUint(count, 10, 64)
				if n > maximum {
					maximum = n
				}
			}
		}
		if len(result.Rows) != len(union) || result.Max != strconv.FormatUint(maximum, 10) || result.Complete != (selected["state"] == "ready" && baseline["state"] == "ready") {
			tc.Fatal("union/max/readiness disagrees with independent oracle", seed, result)
			return
		}
		seen := map[string]bool{}
		for _, row := range result.Rows {
			key := categoryKey338(row.Value)
			if !union[key] || seen[key] {
				tc.Fatal("invented or duplicate union category", seed, key)
				return
			}
			seen[key] = true
			for side, cell := range []categoryCell338{row.Selected, row.Baseline} {
				if sides[side]["state"] != "ready" {
					if cell.State != "unknown" || cell.Count != nil {
						tc.Fatal("uncertain side converted to zero", seed, key, side, cell)
						return
					}
					continue
				}
				expected, present := members[side][key]
				state := "observed"
				if !present {
					expected = "0"
					state = "absent"
				}
				if cell.State != state || cell.Count == nil || *cell.Count != expected {
					tc.Fatal("exact partition cell changed", seed, key, side, cell)
					return
				}
			}
		}
		if !slices.Equal(slices.Sorted(maps.Keys(seen)), slices.Sorted(maps.Keys(union))) {
			tc.Fatal("union identities changed", seed)
		}
	}, options(40)...)
}
func TestCategoryPairFullUnion338(t *testing.T) {
	cmd := exec.Command("node", "--input-type=module", "-e", categoryPairJS338)
	cmd.Stdin = strings.NewReader(`{"large":true}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	var result categoryProjection338
	if err = json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2000 || result.Max != "1" || !result.Complete {
		t.Fatal("independent 1000-value operands did not retain their full2000-value union")
	}
	for _, row := range result.Rows {
		if !reflect.DeepEqual([]string{row.Selected.State, row.Baseline.State}, []string{"observed", "absent"}) && !reflect.DeepEqual([]string{row.Selected.State, row.Baseline.State}, []string{"absent", "observed"}) {
			t.Fatal("disjoint union membership changed")
		}
	}
}

const categoryPairJS338 = `
import fs from 'node:fs';import assert from 'node:assert/strict';import {projection,countedState} from '../ui/categorical-pair.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8'));globalThis.fetch=()=>{throw Error('pure grouping projection attempted IO')};
const field=input.field||{kind:'attribute',key:'http.request.method'},receipt={readAt:'2026-11-05T00:00:00Z',basis:'a'.repeat(64)},ready=(items)=>({state:'ready',field,items,total:String(items.reduce((n,x)=>n+BigInt(x.count),0n)),readAt:receipt.readAt,basis:'b'.repeat(64),reads:{count:receipt,groups:receipt}});
const a=input.large?ready(Array.from({length:1000},(_,i)=>({value:'selected-'+i,count:'1'}))):input.selected,b=input.large?ready(Array.from({length:1000},(_,i)=>({value:'baseline-'+i,count:'1'}))):input.baseline,before=JSON.stringify([a,b]),answer=projection(a,b,field),reverse=projection(b,a,field);
const key=value=>value===null?'0':'1'+value,reverseRows=new Map(reverse.rows.map(row=>[key(row.value),row]));
assert.equal(answer.max,reverse.max);assert.equal(answer.complete,reverse.complete);for(const row of answer.rows){const other=reverseRows.get(key(row.value));assert.deepEqual(row.selected,other.baseline);assert.deepEqual(row.baseline,other.selected)}
assert.equal(JSON.stringify([a,b]),before);const good=ready([{value:null,count:'1'},{value:'',count:'2'},{value:'null',count:'3'}]);
for(const mutate of [x=>x.state='diagnostic',x=>x.total='7',x=>x.items[0].count='0',x=>x.items[0].count='01',x=>x.items[0].count='18446744073709551616',x=>x.items.push({...x.items[0]}),x=>x.field={...field,key:'other'},x=>x.reads.count.basis='bad',x=>delete x.reads.groups,x=>x.items[0].value='\ud800',x=>x.sourceScope='forged',x=>{x.items=Array.from({length:1001},(_,i)=>({value:String(i),count:'1'}));x.total='1001'}]){const bad=structuredClone(good);mutate(bad);assert.throws(()=>countedState(bad,field));assert.throws(()=>projection(good,bad,field))}
if(answer.rows.length)answer.rows[0].selected.count='999';assert.equal(JSON.stringify([a,b]),before);process.stdout.write(JSON.stringify(projection(a,b,field)));
`
