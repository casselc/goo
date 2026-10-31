package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
)

const reuseParityJS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {reuse} from '../ui/view-declaration.mjs';
const {declaration,source,destination,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));
const before=JSON.stringify({declaration,source,destination});let actual;try{actual=reuse(declaration,source,destination)}catch{}
assert.equal(!!actual,accepted);assert.equal(JSON.stringify({declaration,source,destination}),before);
if(actual){assert.deepEqual(actual,declaration);actual.display.title='changed';assert.notEqual(declaration.display.title,'changed');}
`

func reuseCase(t interface{ Fatal(...any) }, seed, mutation int) {
	s := exploration.Query{Signal: "spans", Environment: "source", FromMS: 1000, ToMS: 2000, Limit: 10, GroupBy: []exploration.Field{{Kind: "attribute", Key: "http.route"}}, Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}, Summary: "span-duration"}
	raw, _ := json.Marshal(s)
	var dest exploration.Query
	json.Unmarshal(raw, &dest)
	dest.Environment = "destination"
	dest.FromMS = 3000
	dest.ToMS = 4000
	dest.Limit = 1
	key := []string{"count", "numeric.mean", "duration.p95"}[seed%3]
	d := workbench.ViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.ViewKind{Name: "table", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: key, Scale: "max", Title: "Reusable requests"}}
	switch mutation {
	case 0:
		dest.Signal = "logs"
		dest.Summary = ""
	case 1:
		dest.GroupBy[0].Kind = "resource"
	case 2:
		dest.GroupBy[0].Key = "other"
	case 3:
		d.Display.Value = "numeric.mean"
		dest.Numeric.Unit = "bytes"
	case 4:
		d.Display.Value = "numeric.mean"
		dest.Numeric.Field.Key = "other"
	case 5:
		d.Display.Value = "numeric.mean"
		dest.Numeric = nil
	case 6:
		d.Display.Value = "duration.p95"
		dest.Summary = ""
	case 7:
		d.Kind.Name = "line"
	case 8:
		dest.GroupTimeSeries = true
	case 9:
		dest.ToMS = dest.FromMS
	}
	before, _ := json.Marshal([]any{s, dest, d})
	got, e := workbench.ReuseViewDeclaration(d, s, dest)
	accepted := e == nil
	if accepted != (mutation < 0) {
		t.Fatal("reuse verdict", seed, mutation, e)
		return
	}
	after, _ := json.Marshal([]any{s, dest, d})
	if string(before) != string(after) {
		t.Fatal("reuse mutated source inputs")
	}
	if accepted && !reflect.DeepEqual(got, d) {
		t.Fatal("reuse changed declaration")
	}
	if e := runNativeHistogramJS(map[string]any{"declaration": d, "source": s, "destination": dest, "accepted": accepted}, reuseParityJS); e != nil {
		t.Fatal(e)
	}
}
func TestHegelViewReuseParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { reuseCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
}
func TestHegelViewReuseRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		reuseCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 9)))
	}, options(50)...)
}
