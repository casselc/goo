package verification

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const httpMethodFacetExpansionJS = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as C from '../ui/view-composition.mjs';import * as T from '../ui/view-template-handoff.mjs';const p=JSON.parse(fs.readFileSync(0,'utf8')),{widget,factory,source,target,expanded,numericWidget,numeric,invalid}=p;
assert.deepEqual(T.request(widget),{template:'otel.http-server-routes',version:1,widget});assert.deepEqual(T.stagedView(factory.view,widget.query),factory.view);assert.deepEqual(C.facet(factory.view.composition).field,{kind:'attribute',key:'http.request.method'});assert.equal(C.facet(factory.view.composition).display.title,'HTTP methods');T.templateView(source);assert.deepEqual(T.authoredRequest(source,target),{source,query:target});assert.deepEqual(T.authoredResponse(expanded,source,target),expanded.view);const expected={...target,offset:0};assert.deepEqual(expanded.view.queries,[expected]);assert.deepEqual(C.facet(expanded.view.composition),C.facet(source.composition));assert.notDeepEqual(source.queries[0].groupBy,expected.groupBy);assert.deepEqual(C.records(expanded.view.composition),C.records(source.composition));assert.deepEqual(numeric.view.queries,[numericWidget.query]);assert.equal(C.facet(numeric.view.composition),undefined);assert.equal(numeric.view.template,undefined);T.stagedView(numeric.view,numericWidget.query);for(const widget of invalid)assert.throws(()=>T.templateView({...source,queries:[widget.query]}));const wrong=structuredClone(expanded);C.facet(wrong.view.composition).field.key='changed.binding';assert.throws(()=>T.authoredResponse(wrong,source,target));const changed=structuredClone(expanded);changed.view.queries[0].where=source.queries[0].where;assert.throws(()=>T.authoredResponse(changed,source,target));const before=structuredClone(source);C.facet(expanded.view.composition).display.title='changed output';assert.deepEqual(source,before);`

func TestHegelHTTPMethodFacetTemplateScopeAndLiteralBinding(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		sourceQ := authoredTemplateQuery(seed, "otel.http-server", "source-http", false)
		sourceQ.GroupBy = []exploration.Field{{Kind: "attribute", Key: "http.route"}}
		widget := exploration.DashboardWidget{Family: "http-server", Key: "dimension", Title: "HTTP routes", State: "supported", Grain: "server-span-record", Query: &sourceQ}
		factory, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: widget})
		if e != nil {
			tc.Fatal(e)
			return
		}
		raw, _ := json.Marshal(factory.View)
		var source workbench.ExplorationView
		if e = json.Unmarshal(raw, &source); e != nil {
			tc.Fatal(e)
			return
		}
		for i := range source.Composition.Children {
			child := &source.Composition.Children[i]
			if child.Facet != nil {
				if seed%3 == 0 {
					child.Facet.Field = exploration.Field{Kind: "resource", Key: "cloud.region"}
				}
				child.Facet.Display.Title = "Literal team field"
				child.Hidden = seed%2 == 0
			}
			if child.Records != nil {
				child.Records.Kind.Name = "record-list"
				child.Records.Display.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "resource", Key: "service.namespace"}}
			}
		}
		target := authoredTemplateQuery(seed, "otel.http-server", "destination-http", true)
		if seed%3 == 0 {
			target.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		}
		before, _ := json.Marshal(source)
		targetBefore, _ := json.Marshal(target)
		expanded, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &source, Query: &target})
		if e != nil {
			tc.Fatal(e)
			return
		}
		expected := target
		expected.Offset = 0
		if !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{expected}) {
			tc.Fatal("destination query changed")
			return
		}
		after, _ := json.Marshal(source)
		targetAfter, _ := json.Marshal(target)
		if string(before) != string(after) || string(targetBefore) != string(targetAfter) {
			tc.Fatal("input alias mutation")
			return
		}
		numericQ := sourceQ
		numericQ.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "bytes"}, Type: "uint64", Parse: "canonical-decimal", Unit: "By"}
		numericWidget := widget
		numericWidget.Query = &numericQ
		numeric, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: numericWidget})
		if e != nil {
			tc.Fatal(e)
			return
		}
		if numeric.View.Template != nil || len(numeric.View.Composition.Children) != 2 || !reflect.DeepEqual(numeric.View.Queries, []exploration.Query{numericQ}) {
			tc.Fatal("specialized HTTP scope/compatibility changed")
			return
		}
		invalid := []exploration.DashboardWidget{}
		for n := 0; n < 3; n++ {
			b, _ := json.Marshal(sourceQ)
			var q exploration.Query
			json.Unmarshal(b, &q)
			switch n {
			case 0:
				q.Where.All[1].Value = "Client"
			case 1:
				q.Where.All = append(q.Where.All[:2], q.Where.All[3:]...)
			case 2:
				q.Where.All[0].Value = ""
			}
			bad := widget
			bad.Query = &q
			if _, e = workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: bad}); e == nil {
				tc.Fatal("HTTP owner profile admitted", n)
				return
			}
			invalid = append(invalid, bad)
		}
		if e = runNativeHistogramJS(map[string]any{"widget": widget, "factory": factory, "source": source, "target": target, "expanded": expanded, "numericWidget": numericWidget, "numeric": numeric, "invalid": invalid}, httpMethodFacetExpansionJS); e != nil {
			tc.Fatal(e)
			return
		}
		encoded, e := workbench.EncodeViewArtifact(workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "team", Name: "http-methods", Revision: "v1", View: expanded.View})
		if e != nil {
			tc.Fatal(e)
			return
		}
		decoded, e := workbench.DecodeViewArtifact(encoded.YAML)
		if e != nil || !reflect.DeepEqual(decoded.Artifact.View, expanded.View) {
			tc.Fatal("portable facet template changed", e)
			return
		}
	}, options(50)...)
}
