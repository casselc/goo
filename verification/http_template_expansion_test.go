package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
)

const httpTemplatePublicJS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {validate} from '../ui/view-declaration.mjs';
const {source,expanded}=JSON.parse(fs.readFileSync(0,'utf8'));
assert.deepEqual(expanded.view.queries,[source]);assert.equal(expanded.template.id,'otel.http-server-routes');assert.equal(expanded.template.version,1);
const d=expanded.view.displays.spans;assert.deepEqual(validate(d,source),d);assert.equal(d.kind.name,'bar');assert.equal(d.display.value,'count');assert.equal(d.display.scale,'population');
for(const key of ['rows','snapshot','sourceBasis','completeness','tail','breakdowns'])assert.equal(expanded.view[key],undefined);
`

func httpTemplateCase(t interface{ Fatal(...any) }, seed, mutation int) {
	plan, e := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "http-server", HTTPServerObserved: true, Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: int64(1000 + seed), ToMS: int64(61000 + seed)}, Service: fmt.Sprintf("api-%d", seed), Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}}, nil)
	if e != nil {
		t.Fatal(e)
		return
	}
	in := workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1}
	for _, w := range plan.Widgets {
		if w.Key == "dimension" {
			in.Widget = w
		}
	}
	q := in.Widget.Query
	if q == nil {
		t.Fatal("missing route widget")
		return
	}
	if seed%2 == 0 {
		selected, e := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
		if e != nil {
			t.Fatal(e)
			return
		}
		wire, e := selected.Selector()
		if e != nil {
			t.Fatal(e)
			return
		}
		q.Environment = ""
		q.EnvironmentSelector = &wire
	}
	extra := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: fmt.Sprintf("team-%d", seed)}
	q.Where = &exploration.Predicate{All: []exploration.Predicate{*q.Where, extra}}
	if seed%3 == 0 {
		q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	}
	switch mutation {
	case 0:
		in.Version = 2
	case 1:
		in.Template = "otel.http-client-routes"
	case 2:
		in.Widget.Family = "all-spans"
	case 3:
		in.Widget.Grain = "client-span-record"
	case 4:
		in.Widget.State = "not_observed"
	case 5:
		q.GroupBy = []exploration.Field{{Kind: "resource", Key: "http.route"}}
	case 6:
		q.Where = &exploration.Predicate{Any: []exploration.Predicate{*q.Where, {Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "other"}}}
	case 7:
		q.Where = &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}, {Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Client"}, {Field: &exploration.Field{Kind: "attribute", Key: "http.request.method"}, Op: "ne"}}}
	case 8:
		q.Where = &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}, {Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Server"}, {Field: &exploration.Field{Kind: "attribute", Key: "http.method"}, Op: "ne"}}}
	case 9:
		q.Offset = 1
	}
	before, _ := json.Marshal(in)
	out, e := workbench.ExpandViewTemplate(in)
	if (e == nil) != (mutation < 0) {
		t.Fatal("template verdict", seed, mutation, e)
		return
	}
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("expansion mutated input")
		return
	}
	if e != nil {
		return
	}
	if len(out.View.Queries) != 1 || !reflect.DeepEqual(out.View.Queries[0], *q) {
		t.Fatal("template changed checked query")
		return
	}
	if e = runNativeHistogramJS(map[string]any{"source": q, "expanded": out}, httpTemplatePublicJS); e != nil {
		t.Fatal(e)
		return
	}
	out.View.Queries[0].Where.All[0].All[0].Value = "mutated"
	final, _ := json.Marshal(in)
	if string(final) != string(before) {
		t.Fatal("expanded query aliases source")
	}
}
func TestHegelHTTPTemplateExpansion(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { httpTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
}
func TestHegelHTTPTemplateRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		httpTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 9)))
	}, options(50)...)
}
