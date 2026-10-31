package verification

import (
	"encoding/json"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const measureDeclarationParityJS = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {validate,presentation,update} from '../ui/view-declaration.mjs';
const {declaration,question,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));
let actual;try{actual=validate(declaration,question)}catch{}
assert.equal(!!actual,accepted,'Go and JS measure declaration verdict differs');
if(actual){assert.deepEqual(actual,declaration);const p=presentation(actual);assert.equal(p.value,declaration.display.value);assert.equal(p.scale,'max');const next=update(actual,{...p,kind:p.kind==='bar'?'table':'bar',order:'label'},question);assert.equal(next.display.value,declaration.display.value);assert.equal(next.display.scale,'max');assert.deepEqual(actual,declaration,'editing changed checked source');}
`

var savedMeasureKeys = []string{"numeric.mean", "numeric.sum", "numeric.min", "numeric.max", "duration.mean", "duration.min", "duration.max", "duration.p50", "duration.p95", "duration.p99"}

func measureDeclarationCase(t interface{ Fatal(...any) }, seed, mutation int) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}, Summary: "span-duration"}
	key := savedMeasureKeys[seed%len(savedMeasureKeys)]
	d := workbench.ViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.ViewKind{Name: []string{"bar", "table"}[seed%2], Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: key, Scale: "max", Order: "value-desc"}}
	switch mutation {
	case 0:
		q.Numeric = nil
		d.Display.Value = "numeric.mean"
	case 1:
		q.Summary = ""
		d.Display.Value = "duration.p95"
	case 2:
		d.Display.Scale = "population"
	case 3:
		d.Display.Value = "duration.p90"
	case 4:
		d.Display.Value = "status.error_fraction"
	case 5:
		d.Display.Value = "work.delta"
	case 6:
		d.Scope.Kind = "measurement-window"
	case 7:
		q.GroupTimeSeries = true
	case 8:
		q.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
	case 9:
		q.GroupBy = append(q.GroupBy, exploration.Field{Kind: "column", Key: "name"})
	}
	accepted := d.Validate(q) == nil
	if accepted != (mutation < 0) {
		t.Fatal("unexpected measure verdict", key, mutation)
		return
	}
	if e := runNativeHistogramJS(map[string]any{"declaration": d, "question": q, "accepted": accepted}, measureDeclarationParityJS); e != nil {
		t.Fatal(e)
	}
	if accepted {
		raw, _ := json.Marshal(d)
		var round workbench.ViewDeclaration
		if e := json.Unmarshal(raw, &round); e != nil || round.Validate(q) != nil {
			t.Fatal("wire roundtrip refused measure", key, e)
		}
	}
}
func TestHegelViewMeasureDeclarationParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { measureDeclarationCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
}
func TestHegelViewMeasureDeclarationRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		measureDeclarationCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 9)))
	}, options(50)...)
}
