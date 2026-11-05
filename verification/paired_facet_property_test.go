package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// The browser planner's actual output must survive authoritative Go admission,
// SQL compilation and portable YAML. Match supplies a separate semantic oracle:
// original operand AND (any included membership) AND NOT excluded membership.
func TestHegelPairedFacetAtomicScope337(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		signal := []string{"logs", "spans"}[seed%2]
		fieldChoice := (seed / 2) % 2
		field := exploration.Field{Kind: []string{"attribute", "resource"}[fieldChoice], Key: []string{"http.request.method", "cloud.region"}[fieldChoice]}
		scope := map[string]any{"kind": "record-query", "signal": signal}
		leaf := func(kind, key, op, value string) map[string]any {
			return map[string]any{"field": map[string]string{"kind": kind, "key": key}, "op": op, "value": value}
		}
		mode := []string{"named", "empty", "unreported"}[seed%3]
		env := map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": mode}}
		if mode == "named" {
			env["scope"].(map[string]any)["value"] = "prod-'雪"
			if seed%5 == 0 {
				env["policy"] = "current-only"
			}
		}
		from := int64(1790000000000 + seed*1000)
		q := map[string]any{"signal": signal, "environmentSelector": env, "fromMs": from, "toMs": from + 60000, "limit": 1 + seed%100, "offset": 0, "groupBy": []exploration.Field{{Kind: "column", Key: "service"}}, "where": map[string]any{"all": []any{leaf("resource", "service.namespace", "eq", ""), map[string]any{"any": []any{leaf("attribute", "tenant", "eq", "selected-'雪"), map[string]any{"field": map[string]string{"kind": "attribute", "key": "fallback"}, "op": "exists"}}}}}}
		if seed%3 == 1 {
			q["bucketMs"] = 1000
		}
		baseline := maps.Clone(q)
		if seed%3 != 0 {
			baseline["fromMs"] = from - 60000
			baseline["toMs"] = from
		}
		baseline["where"] = map[string]any{"all": []any{map[string]any{"field": map[string]string{"kind": "resource", "key": "service.namespace"}, "op": "missing"}, map[string]any{"any": []any{leaf("attribute", "tenant", "eq", "baseline-'雪"), map[string]any{"field": map[string]string{"kind": "attribute", "key": "fallback"}, "op": "exists"}}}}}
		facet := map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "categorical-facet", "version": 1}, "field": field, "display": map[string]any{"basis": "applied", "title": "Method or region"}}
		comparison := map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "query-count-comparison", "version": 1}, "baselineQuery": baseline, "display": map[string]any{"value": "count", "view": "groups", "order": "change"}}
		composition := map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "bar", "version": 1}, "display": map[string]any{"category": "category", "value": "count"}}}, map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "record-table", "version": 1}, "display": map[string]any{"columns": []string{"utc", "service", "summary"}}}}, map[string]any{"comparison": comparison}, map[string]any{"facet": facet}}}
		if seed%3 == 1 {
			composition["children"] = append(composition["children"].([]any), map[string]any{"timeline": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "time-bar", "version": 1}, "display": map[string]any{"value": "count"}}})
		}
		inputView := map[string]any{"version": 1, "title": "Paired facet scope", "queries": []any{q}, "composition": composition}
		values := []any{nil, "", "POST-'雪"}
		items := []map[string]any{{"signal": signal, "field": field, "value": values[seed%3], "mode": "include"}}
		if seed%2 == 0 {
			items = append(items, map[string]any{"signal": signal, "field": field, "value": values[(seed+1)%3], "mode": "include"})
		}
		items = append(items, map[string]any{"signal": signal, "field": field, "value": values[(seed+2)%3], "mode": "exclude"})
		raw, err := json.Marshal(map[string]any{"view": inputView, "items": items})
		if err != nil {
			tc.Fatal(err)
			return
		}
		var before workbench.ExplorationView
		viewRaw, _ := json.Marshal(inputView)
		if err = json.Unmarshal(viewRaw, &before); err != nil {
			tc.Fatal(err)
			return
		}
		if err = before.Validate(); err != nil {
			tc.Fatal(err)
			return
		}
		cmd := exec.Command("node", "--input-type=module", "-e", pairedFacetJS337)
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var after workbench.ExplorationView
		if err = json.Unmarshal(out, &after); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if err = after.Validate(); err != nil {
			tc.Fatal(err)
			return
		}
		for _, sides := range [][2]exploration.Query{{before.Queries[0], after.Queries[0]}, {*comparison337(before.Composition).BaselineQuery, *comparison337(after.Composition).BaselineQuery}} {
			original, narrowed := sides[0], sides[1]
			if narrowed.Where == nil || len(narrowed.Where.All) < 2 || !reflect.DeepEqual(narrowed.Where.All[0], *original.Where) {
				tc.Fatal("whole original predicate was not preserved")
				return
			}
			expected := original
			expected.Where = narrowed.Where
			if !reflect.DeepEqual(expected, narrowed) {
				tc.Fatal("operand window, typed policy or query profile changed")
				return
			}
			plan, e := exploration.Compile(narrowed)
			if e != nil || plan.Rows == "" || plan.Count == "" || plan.Parameters["from"] != fmt.Sprint(original.FromMS) || plan.Parameters["to"] != fmt.Sprint(original.ToMS) {
				tc.Fatal("record/count SQL lost exact operand bounds", e)
				return
			}
			for _, namespace := range []*string{nil, text337(""), text337("other")} {
				for _, tenant := range []string{"selected-'雪", "baseline-'雪", "other"} {
					for _, present := range []bool{false, true} {
						for _, v := range []*string{nil, text337(""), text337("POST-'雪"), text337("other")} {
							row := map[exploration.Field]string{{Kind: "attribute", Key: "tenant"}: tenant}
							if namespace != nil {
								row[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *namespace
							}
							if present {
								row[exploration.Field{Kind: "attribute", Key: "fallback"}] = "yes"
							}
							if v != nil {
								row[field] = *v
							}
							included, hasIncludes, excluded := false, false, false
							for _, item := range items {
								wanted := item["value"]
								matches := wanted == nil && v == nil || wanted != nil && v != nil && wanted.(string) == *v
								if item["mode"] == "include" {
									hasIncludes = true
									included = included || matches
								} else {
									excluded = excluded || matches
								}
							}
							want := exploration.Match(original.Where, row) && (!hasIncludes || included) && !excluded
							if exploration.Match(narrowed.Where, row) != want {
								tc.Fatal("common membership truth table differs", field, row, items)
								return
							}
						}
					}
				}
			}
		}
		a, b := after.Queries[0].Where.All, comparison337(after.Composition).BaselineQuery.Where.All
		if !reflect.DeepEqual(a[1:], b[1:]) {
			tc.Fatal("common delta differs across operands")
			return
		}
		artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "team", Name: "paired-facet", Revision: fmt.Sprint(seed), View: after}
		encoded, err := workbench.EncodeViewArtifact(artifact)
		if err != nil {
			tc.Fatal(err)
			return
		}
		decoded, err := workbench.DecodeViewArtifact(encoded.YAML)
		if err != nil || !reflect.DeepEqual(decoded.Artifact, artifact) {
			tc.Fatal("portable pair/facet definition changed", err)
			return
		}
	}, options(40)...)
}
func text337(s string) *string { return &s }
func comparison337(c *workbench.ViewComposition) *workbench.CountComparisonViewDeclaration {
	for _, child := range c.Children {
		if child.Comparison != nil {
			return child.Comparison
		}
	}
	return nil
}

const pairedFacetJS337 = `
import fs from 'node:fs';import assert from 'node:assert/strict';import * as C from '../ui/view-composition.mjs';
const {view,items}=JSON.parse(fs.readFileSync(0,'utf8')),q=view.queries[0],before=JSON.stringify({view,items});
globalThis.fetch=()=>{throw Error('pure planner attempted IO')};
const next=C.narrowFixedComparison(view.composition,q,items),answer={...view,queries:[next.query],composition:next.composition};
assert.equal(JSON.stringify({view,items}),before);assert.deepEqual(C.facet(next.composition),C.facet(view.composition));assert.deepEqual(C.timeline(next.composition),C.timeline(view.composition));
const duplicate=structuredClone(view.composition);duplicate.children.push({facet:structuredClone(C.facet(duplicate))});assert.throws(()=>C.validate(duplicate,q));
for(const badItems of [[],[...items,...Array.from({length:33},()=>items[0])],items.map(i=>({...i,origin:'old-owner'})),items.map(i=>({...i,field:{...i.field,key:'different.field'}})),items.map(i=>({...i,signal:i.signal==='logs'?'spans':'logs'})),items.map(i=>({...i,value:{not:'membership'}}))]){assert.throws(()=>C.narrowFixedComparison(view.composition,q,badItems));assert.equal(JSON.stringify({view,items}),before)}
const absent=structuredClone(view.composition);absent.children=absent.children.filter(c=>!c.facet);assert.throws(()=>C.narrowFixedComparison(absent,q,items));
const captured=structuredClone(view.composition),pair=C.comparison(captured);pair.kind.name='captured-count-comparison';pair.baselineCapture='a'.repeat(64);delete pair.baselineQuery;assert.throws(()=>C.narrowFixedComparison(captured,q,items));
// Only the baseline is at the old predicate budget. Selected scope could fit,
// but failure must not mutate either question or the author composition.
const budget=structuredClone(view.composition);C.comparison(budget).baselineQuery.where={any:Array.from({length:31},()=>({field:{kind:'attribute',key:'tenant'},op:'exists'}))};C.validate(budget,q);const old=JSON.stringify(budget);assert.throws(()=>C.narrowFixedComparison(budget,q,items));assert.equal(JSON.stringify(budget),old);assert.equal(JSON.stringify({view,items}),before);
const wire=JSON.stringify(answer);next.query.fromMs++;C.comparison(next.composition).baselineQuery.where={field:{kind:'attribute',key:'edited'},op:'exists'};assert.equal(JSON.stringify({view,items}),before);process.stdout.write(wire);
`
