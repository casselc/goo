package verification

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const recordListPublicJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';import * as P from '../ui/record-page.mjs';import * as L from '../ui/record-list.mjs';
const {query,declaration,accepted,raw,expanded}=JSON.parse(fs.readFileSync(0,'utf8'));let verdict=true;try{C.recordDisplay(declaration,query)}catch{verdict=false}assert.equal(verdict,accepted);if(!accepted)process.exit(0);
assert.deepEqual(C.recordDisplay(declaration,query),declaration);assert.equal(expanded.view.composition.children.find(c=>c.records).records.kind.name,declaration.kind.name);
const before=structuredClone(raw),page=P.fromExploration(raw),model=L.listModel(page,declaration.display);assert.equal(P.adoptPage(page),page);assert.deepEqual(raw,before);assert(Object.isFrozen(page)&&Object.isFrozen(page.rows)&&Object.isFrozen(page.rows[0].Attributes));assert.equal(page.page.returned,1);assert.equal(page.page.hasMore,Object.hasOwn(raw,'hasMore')?raw.hasMore:null);assert.equal(page.quality.extent,'bounded-page');assert(!Object.hasOwn(page,'count'));assert(!Object.hasOwn(page,'completeness'));assert.deepEqual(model.rows.map(r=>r.index),[0]);assert(model.rows[0].fields.some(f=>f.field.kind==='attribute'&&f.field.key==='zone'&&f.state==='empty'));assert(model.rows[0].fields.some(f=>f.field.kind==='resource'&&f.field.key==='zone'&&f.text==='resource-zone'));assert.equal(L.fieldValue(page.rows[0],{kind:'attribute',key:'nullable'}).state,'null');assert.equal(L.fieldValue(page.rows[0],{kind:'attribute',key:'absent'}).state,'missing');
const copied=JSON.parse(JSON.stringify(page)),imported=P.adoptPage(copied);copied.rows[0].Attributes.zone='changed';assert.equal(imported.rows[0].Attributes.zone,'');assert.notEqual(imported,copied);assert.throws(()=>{model.page.rows[0].Attributes.zone='changed'});
for(const mutate of [p=>p.page.returned++,p=>p.page.offset++,p=>p.page.hasMore='false',p=>p.quality.extent='complete',p=>p.count='999',p=>p.completeness={aspects:['records']},p=>p.rows[0].Attributes.accessor=()=>{}]){const bad=structuredClone(page);mutate(bad);assert.throws(()=>P.adoptPage(bad));}
let invoked=0;const accessor=structuredClone(page);Object.defineProperty(accessor.rows[0].Attributes,'danger',{enumerable:true,get(){invoked++;return 'bad'}});assert.throws(()=>P.adoptPage(accessor));assert.equal(invoked,0);assert.deepEqual(raw,before);
`

func recordListCase(tc *hegel.T, seed, bad int) {
	signal := "spans"
	if seed%2 == 0 {
		signal = "logs"
	}
	q := exploration.Query{Signal: signal, Environment: "source", FromMS: 1000, ToMS: 61000, Limit: seed%20 + 1, Offset: 0, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "source-service"}}
	kind := "record-list"
	if seed%3 == 0 {
		kind = "record-table"
	}
	columns := []string{"utc", "summary"}
	if seed%2 == 0 {
		columns = append(columns, "service")
	}
	if seed%3 == 0 {
		columns = append(columns, "context")
	}
	d := workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: signal}, Kind: workbench.RecordViewKind{Name: kind, Version: 1}, Display: workbench.RecordViewDisplay{Columns: columns, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "zone"}, {Kind: "resource", Key: "zone"}}}}
	switch bad {
	case 0:
		d.Kind.Name = "record-card"
	case 1:
		d.Kind.Version = 2
	case 2:
		d.Display.Columns = []string{"summary", "service"}
	case 3:
		d.Display.FieldColumns = append(d.Display.FieldColumns, d.Display.FieldColumns[0])
	case 4:
		d.Display.Columns = []string{"utc", "summary", "query"}
	case 5:
		d.Display.FieldColumns[0].Kind = "physical"
	case 6:
		d.Scope.Signal = map[string]string{"spans": "logs", "logs": "spans"}[signal]
	}
	before, _ := json.Marshal(d)
	err := d.Validate(q)
	accepted := bad < 0
	if (err == nil) != accepted {
		tc.Fatal("record declaration verdict", seed, bad, err)
		return
	}
	after, _ := json.Marshal(d)
	if string(before) != string(after) {
		tc.Fatal("declaration validation mutated input")
		return
	}
	var expanded workbench.ViewTemplateExpansion
	if accepted {
		data, _ := json.Marshal(map[string]any{"version": 1, "title": "Reusable record shape", "queries": []exploration.Query{q}, "template": map[string]any{"version": "goo.template/1", "group": "destination", "profile": "ordinary"}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": "bar", "version": 1}, "display": map[string]any{"category": "category", "value": "count", "scale": "population"}}}, map[string]any{"records": d}}}})
		var source workbench.ExplorationView
		if e := json.Unmarshal(data, &source); e != nil {
			tc.Fatal(e)
			return
		}
		target := q
		target.Environment = "destination"
		target.FromMS = 101000
		target.ToMS = 161000
		target.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "destination-service"}
		target.GroupBy = []exploration.Field{{Kind: "resource", Key: "cohort"}}
		in := workbench.ViewTemplateExpansionRequest{Source: &source, Query: &target}
		original, _ := json.Marshal(in)
		var e error
		expanded, e = workbench.ExpandViewTemplate(in)
		if e != nil {
			tc.Fatal(e)
			return
		}
		unchanged, _ := json.Marshal(in)
		if string(original) != string(unchanged) || !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{target}) {
			tc.Fatal("template changed scope/input")
			return
		}
	}
	row := map[string]any{"ServiceName": "service", "timestamp_ns": "1700000000123456789", "Attributes": map[string]any{"zone": "", "nullable": nil, "rich": map[string]any{"values": []any{1, "two", nil}}}, "ResourceAttributes": map[string]any{"zone": "resource-zone"}}
	if signal == "spans" {
		row["SpanName"] = "span"
		row["duration_ns"] = "1000000"
	} else {
		row["Body"] = "log"
		row["SeverityText"] = "INFO"
	}
	raw := map[string]any{"question": q, "basis": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "capturedAt": "2026-10-31T00:00:00Z", "state": []string{"observed", "changed_between_reads"}[seed%2], "rows": []any{row}, "count": "999"}
	if seed%3 < 2 {
		raw["hasMore"] = seed%3 == 0
	}
	if e := runNativeHistogramJS(map[string]any{"query": q, "declaration": d, "accepted": accepted, "raw": raw, "expanded": expanded}, recordListPublicJS); e != nil {
		tc.Fatal(e)
	}
}
func TestHegelRecordListDeclarationsAndSuppliedPages(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		hegel.Test(t, func(tc *hegel.T) { recordListCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
	})
	t.Run("refused", func(t *testing.T) {
		hegel.Test(t, func(tc *hegel.T) {
			recordListCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 6)))
		}, options(50)...)
	})
}
