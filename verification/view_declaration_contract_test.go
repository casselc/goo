package verification

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const declarationParityScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {validate,presentation,update} from '../ui/view-declaration.mjs';
const {declaration,question,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));
let actual;try{actual=validate(declaration,question)}catch{}
assert.equal(!!actual,accepted,'Go and JS declaration acceptance differs');
if(actual){assert.deepEqual(actual,declaration);const p=presentation(actual);assert.equal(p.kind,declaration.kind.name);assert.equal(p.value,'count');assert.equal(p.query,declaration.display.query||'');const changed=update(actual,{...p,kind:p.kind==='bar'?'table':'bar',order:'label'},question);assert.equal(changed.kind.name,p.kind==='bar'?'table':'bar');assert.equal(changed.display.order,'label');assert.deepEqual(actual,declaration,'property edit mutated input');}
`

func declarationCase(t interface{ Fatal(...any) }, seed, mutation int) {
	signal := []string{"spans", "logs"}[seed%2]
	q := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
	d := workbench.ViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: signal}, Kind: workbench.ViewKind{Name: []string{"bar", "table"}[seed%2], Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count", Title: fmt.Sprintf("Services %d", seed), Scale: []string{"max", "population"}[seed%2], Order: []string{"source", "label", "value-desc", "value-asc"}[seed%4], Query: []string{"", "worker", "é", "服务"}[seed%4], RequiresComplete: seed%2 == 0}}
	switch mutation {
	case 0:
		d.Version = "goo.view/2"
	case 1:
		d.Scope.Signal = map[string]string{"spans": "logs", "logs": "spans"}[signal]
	case 2:
		d.Scope.Kind = "measurement-window"
	case 3:
		d.Kind.Name = "line"
	case 4:
		d.Kind.Version = 2
	case 5:
		d.Display.Value = "p95"
	case 6:
		d.Display.Query = strings.Repeat("é", 129)
	case 7:
		d.Display.Title = strings.Repeat("é", 61)
	case 8:
		d.Display.Query = "a\nb"
	case 9:
		q.GroupBy = nil
	case 10:
		q.GroupBy = append(q.GroupBy, exploration.Field{Kind: "attribute", Key: "region"})
	case 11:
		q.GroupTimeSeries = true
	case 12:
		d.Display.Scale = "valid-share"
	case 13:
		d.Display.Order = "random"
	}
	accepted := d.Validate(q) == nil
	if accepted != (mutation < 0) {
		t.Fatal("unexpected Go declaration verdict", mutation)
		return
	}
	if err := runNativeHistogramJS(map[string]any{"declaration": d, "question": q, "accepted": accepted}, declarationParityScript); err != nil {
		t.Fatal(err)
	}
	if !accepted {
		return
	}
	v := workbench.ExplorationView{Version: 1, Title: "Authored services", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}, Displays: map[string]workbench.ViewDeclaration{signal: d}}
	// Public configuration remains valid after wire serialization.
	raw, _ := json.Marshal(v)
	var clone workbench.ExplorationView
	json.Unmarshal(raw, &clone)
	if err := clone.Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestHegelViewDeclarationParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { declarationCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(40)...)
}
func TestHegelViewDeclarationRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		declarationCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 13)))
	}, options(40)...)
}

func TestViewDeclarationImmutableEditSequence(t *testing.T) {
	s := workbench.Store{Dir: t.TempDir()}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
	v := workbench.ExplorationView{Version: 1, Title: "Reusable services", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}, Displays: map[string]workbench.ViewDeclaration{"spans": {Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.ViewKind{Name: "bar", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count"}}}}
	revisions := map[string]string{}
	for i := 0; i < 20; i++ {
		d := v.Displays["spans"]
		d.Display.Title = fmt.Sprintf("Revision %d", i)
		d.Display.Query = fmt.Sprintf("team-%d", i)
		d.Kind.Name = []string{"bar", "table"}[i%2]
		v.Displays["spans"] = d
		saved, e := s.CreateExplorationView(v)
		if e != nil {
			t.Fatal(e)
		}
		if _, ok := revisions[saved.ID]; ok {
			t.Fatal("distinct authored revision reused identity")
		}
		revisions[saved.ID] = d.Display.Title
		for id, title := range revisions {
			old, e := s.LoadExplorationView(id)
			if e != nil || old.Displays["spans"].Display.Title != title {
				t.Fatal("old config mutated", id, e)
			}
			if old.Queries[0].Limit != 1 || old.Queries[0].GroupBy[0].Key != "service" {
				t.Fatal("presentation changed typed source")
			}
		}
	}
}
