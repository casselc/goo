package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const compositionParityJS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';
const {value,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));let actual;try{actual=C.view(value)}catch{}
assert.equal(!!actual,accepted,'Go and JS composition verdict differs');if(actual){assert.deepEqual(actual,value);const before=structuredClone(actual);const reversed=C.reverse(actual.composition,actual.queries[0]);assert.equal(C.role(reversed.children[0]),C.role(actual.composition.children[1]));assert.deepEqual(actual,before,'composition edit mutated input');assert.equal(C.records(actual.composition).kind.name,'record-table');}
`

func compositionWire(seed int) map[string]any {
	signal := []string{"spans", "logs"}[seed%2]
	query := map[string]any{"signal": signal, "environment": "prod", "fromMs": 1000, "toMs": 2000, "limit": 1, "offset": 0, "groupBy": []any{map[string]any{"kind": "column", "key": "service"}}}
	agg := map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": []string{"bar", "table"}[seed%2], "version": 1}, "display": map[string]any{"category": "category", "value": "count", "scale": "population", "order": "value-desc"}}
	records := map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": "record-table", "version": 1}, "display": map[string]any{"columns": []any{"utc", "summary"}, "fieldColumns": []any{map[string]any{"kind": "attribute", "key": fmt.Sprintf("team.%d", seed)}}}}
	a, r := map[string]any{"aggregate": agg}, map[string]any{"records": records}
	if seed%3 == 0 {
		a["hidden"] = true
	}
	if seed%3 == 1 {
		r["hidden"] = true
	}
	children := []any{a, r}
	if seed%2 == 1 {
		children = []any{r, a}
	}
	return map[string]any{"version": 1, "title": fmt.Sprintf("View %d", seed), "queries": []any{query}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}}
}
func compositionCase(t interface{ Fatal(...any) }, seed, mutation int) {
	value := compositionWire(seed)
	comp := value["composition"].(map[string]any)
	children := comp["children"].([]any)
	var a, r map[string]any
	for _, child := range children {
		v := child.(map[string]any)
		if v["aggregate"] != nil {
			a = v
		} else {
			r = v
		}
	}
	switch mutation {
	case 0:
		value["layout"] = nil
	case 1:
		value["columns"] = []any{}
	case 2:
		value["displays"] = nil
	case 3:
		value["tail"] = nil
	case 4:
		value["breakdowns"] = map[string]any{}
	case 5:
		value["fieldColumns"] = nil
	case 6:
		comp["children"] = []any{a, a}
	case 7:
		a["hidden"] = true
		r["hidden"] = true
	case 8:
		r["records"].(map[string]any)["scope"].(map[string]any)["signal"] = map[string]string{"spans": "logs", "logs": "spans"}[value["queries"].([]any)[0].(map[string]any)["signal"].(string)]
	case 9:
		r["records"].(map[string]any)["display"].(map[string]any)["columns"] = []any{"service", "summary"}
	case 10:
		r["records"].(map[string]any)["display"].(map[string]any)["fieldColumns"] = []any{map[string]any{"kind": "attribute", "key": "x"}, map[string]any{"kind": "attribute", "key": "x"}}
	case 11:
		comp["kind"].(map[string]any)["version"] = 2
	case 12:
		comp["children"] = []any{a, r, a}
	case 13:
		r["records"].(map[string]any)["kind"].(map[string]any)["name"] = "table"
	case 14:
		a["query"] = value["queries"].([]any)[0]
	case 15:
		value["composition"] = nil
	case 16:
		comp["children"] = []any{map[string]any{"aggregate": a["aggregate"], "records": r["records"]}, r}
	case 17:
		a["hidden"] = "yes"
	case 18:
		r["records"].(map[string]any)["display"].(map[string]any)["columns"] = []any{"utc", "summary", "utc"}
	case 19:
		value["queries"] = append(value["queries"].([]any), value["queries"].([]any)[0])
	}
	raw, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
		return
	}
	var v workbench.ExplorationView
	e = json.Unmarshal(raw, &v)
	if e == nil {
		e = v.Validate()
	}
	accepted := e == nil
	if accepted != (mutation < 0) {
		t.Fatal("unexpected composition verdict", seed, mutation, e)
		return
	}
	if e := runNativeHistogramJS(map[string]any{"value": value, "accepted": accepted}, compositionParityJS); e != nil {
		t.Fatal(e)
	}
}
func TestHegelViewCompositionValid(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { compositionCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), -1) }, options(50)...)
}
func TestHegelViewCompositionRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		compositionCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), hegel.Draw(tc, hegel.Integers(0, 19)))
	}, options(50)...)
}
