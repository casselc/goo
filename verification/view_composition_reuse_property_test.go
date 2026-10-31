package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const compositionReuseParityJS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';
const {composition,source,destination,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));const original=structuredClone({composition,source,destination});let actual;try{actual=C.reuse(composition,source,destination)}catch{}
assert.equal(!!actual,accepted,'Go/JS composition reuse differs');assert.deepEqual({composition,source,destination},original,'reuse changed input');
if(actual){assert.deepEqual(actual,composition);const r=C.records(actual);r.display.columns[0]='changed';if(r.display.fieldColumns?.length)r.display.fieldColumns[0].key='changed';C.aggregate(actual).display.title='changed';assert.deepEqual({composition,source,destination},original,'returned intent aliases source');}
`

func compositionReuseCase(t interface{ Fatal(...any) }, seed, mutation int) {
	wire := compositionWire(seed)
	raw, _ := json.Marshal(wire)
	var view workbench.ExplorationView
	if e := json.Unmarshal(raw, &view); e != nil {
		t.Fatal(e)
		return
	}
	source := view.Queries[0]
	var agg *workbench.ViewDeclaration
	var records *workbench.RecordViewDeclaration
	for i := range view.Composition.Children {
		c := &view.Composition.Children[i]
		if c.Aggregate != nil {
			agg = c.Aggregate
		} else {
			records = c.Records
		}
	}
	keys := []string{"count", "numeric.mean", "numeric.sum", "numeric.min", "numeric.max", "duration.mean", "duration.min", "duration.max", "duration.p50", "duration.p95", "duration.p99"}
	key := keys[seed%len(keys)]
	if len(key) > 8 && key[:8] == "duration" {
		source.Signal = "spans"
		source.Summary = "span-duration"
		agg.Scope.Signal = "spans"
		records.Scope.Signal = "spans"
	}
	if key != "count" {
		agg.Display.Value = key
		agg.Display.Scale = "max"
	}
	source.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	srcraw, _ := json.Marshal(source)
	var destination exploration.Query
	_ = json.Unmarshal(srcraw, &destination)
	destination.Environment = "destination"
	destination.FromMS = 3000
	destination.ToMS = 4000
	destination.Limit = []int{1, 10, 100}[seed%3]
	destination.Offset = seed % 5
	destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "current"}
	switch mutation {
	case 0:
		destination.Signal = map[string]string{"spans": "logs", "logs": "spans"}[source.Signal]
		destination.Summary = ""
	case 1:
		destination.GroupBy = []exploration.Field{{Kind: "resource", Key: "service"}}
	case 2:
		agg.Display.Value = "numeric.mean"
		agg.Display.Scale = "max"
		destination.Numeric.Unit = "bytes"
	case 3:
		agg.Display.Value = "numeric.mean"
		agg.Display.Scale = "max"
		destination.Numeric.Field.Key = "other"
	case 4:
		agg.Display.Value = "numeric.mean"
		agg.Display.Scale = "max"
		destination.Numeric.Type = "decimal64"
		destination.Numeric.Scale = 2
	case 5:
		records.Display.Columns = []string{"utc", "utc"}
	case 6:
		for i := range view.Composition.Children {
			view.Composition.Children[i].Hidden = true
		}
	case 7:
		records.Scope.Signal = map[string]string{"spans": "logs", "logs": "spans"}[source.Signal]
	case 8:
		view.Composition.Version = "goo.composition/2"
	case 9:
		records.Display.FieldColumns = []exploration.Field{{Kind: "column", Key: "unsupported"}}
	case 10:
		agg.Display.Value = "numeric.mean"
		agg.Display.Scale = "max"
		destination.Numeric = nil
	case 11:
		view.Composition.Children = append(view.Composition.Children, view.Composition.Children[0])
	}
	before, _ := json.Marshal(map[string]any{"composition": view.Composition, "source": source, "destination": destination})
	out, e := workbench.ReuseViewComposition(*view.Composition, source, destination)
	accepted := e == nil
	if accepted != (mutation < 0) {
		t.Fatal("unexpected reuse verdict", seed, mutation, e)
		return
	}
	if e := runNativeHistogramJS(map[string]any{"composition": view.Composition, "source": source, "destination": destination, "accepted": accepted}, compositionReuseParityJS); e != nil {
		t.Fatal(e)
	}
	if accepted {
		for i := range out.Children {
			if out.Children[i].Records != nil {
				out.Children[i].Records.Display.Columns[0] = "changed"
				if len(out.Children[i].Records.Display.FieldColumns) > 0 {
					out.Children[i].Records.Display.FieldColumns[0].Key = "changed"
				}
			} else {
				out.Children[i].Aggregate.Display.Title = "changed"
			}
		}
	}
	after, _ := json.Marshal(map[string]any{"composition": view.Composition, "source": source, "destination": destination})
	if string(before) != string(after) {
		t.Fatal("reuse aliased or mutated inputs")
	}
}
func TestHegelViewCompositionReuseValid(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { compositionReuseCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), -1) }, options(50)...)
}
func TestHegelViewCompositionReuseRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		compositionReuseCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), hegel.Draw(tc, hegel.Integers(0, 11)))
	}, options(50)...)
}
