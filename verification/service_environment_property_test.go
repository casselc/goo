package verification

import (
	"fmt"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

const serviceEnvironment318JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import * as H from '../ui/service-handoff.mjs';import * as E from '../ui/record-environment.mjs';
const {q,row,resources,matches,expected,nativeName}=JSON.parse(fs.readFileSync(0,'utf8'));
const before=structuredClone({q,row}),intent=H.create(q,row);
assert.deepEqual(H.scope(intent),expected);assert.deepEqual(intent.origin,q);
assert.deepEqual(resources.map(r=>E.matches(intent.environmentSelector,r)),matches);
assert.equal(E.currentName(q),nativeName);assert.deepEqual({q,row},before);
for(const [i,resource]of resources.entries()){
 const candidate={...row,ResourceAttributes:{...resource,...(Object.hasOwn(row.ResourceAttributes,'service.namespace')?{'service.namespace':row.ResourceAttributes['service.namespace']}:{} )}};
 if(matches[i])assert.equal(H.create(q,candidate).service,row.ServiceName);else assert.throws(()=>H.create(q,candidate));
}
for(const mutate of [x=>x.version='unknown',x=>x.origin.toMs++,x=>x.environmentSelector.policy='unknown',x=>x.origin.environment='ambiguous']){
 const invalid=structuredClone(intent);mutate(invalid);assert.throws(()=>H.validate(invalid));
}
assert(Object.isFrozen(intent)&&Object.isFrozen(intent.origin)&&Object.isFrozen(intent.environmentSelector));
`

// Generated policies cross the browser intent, Go service scope and dashboard
// boundary. Resource membership uses the checked semantic model as its oracle;
// live ClickHouse cases separately execute current/legacy precedence and absence.
func TestHegelTypedServiceEnvironment318(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"named", "empty", "unreported"}))
		name := fmt.Sprintf("prod-%d'雪", seed%17)
		semanticScope := identity.EnvironmentScope{Mode: mode}
		if mode == "named" {
			semanticScope.Value = name
		}
		selection, err := identity.EffectiveEnvironment(semanticScope)
		if mode == "named" && hegel.Draw(tc, hegel.Booleans()) {
			selection, err = identity.CurrentEnvironment(name)
		}
		if err != nil {
			tc.Fatal(err)
			return
		}
		wire, err := selection.Selector()
		if err != nil {
			tc.Fatal(err)
			return
		}
		from := int64(seed) * 1000
		service := fmt.Sprintf("api-%d", seed%19)
		scope := exploration.InventoryQuestion{EnvironmentSelector: &wire, FromMS: from, ToMS: from + 60000}
		resources := []map[string]string{
			{}, {identity.EnvironmentKey: name}, {identity.LegacyEnvironmentKey: name},
			{identity.EnvironmentKey: "other", identity.LegacyEnvironmentKey: name},
			{identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: name},
			{identity.LegacyEnvironmentKey: ""},
			{identity.EnvironmentKey: name, identity.LegacyEnvironmentKey: "other"},
		}
		matches := make([]bool, len(resources))
		var rowResource map[string]string
		for i, resource := range resources {
			matches[i] = selection.Matches(resource)
			if matches[i] && rowResource == nil {
				rowResource = map[string]string{}
				for k, v := range resource {
					rowResource[k] = v
				}
			}
		}
		if rowResource == nil {
			tc.Fatal("no matching generated resource")
			return
		}
		ns := map[string]any{"mode": "missing"}
		scope.Namespace = &identity.Namespace{Mode: "missing"}
		if seed%3 != 0 {
			value := ""
			if seed%3 == 2 {
				value = fmt.Sprintf("team-%d'雪", seed%11)
			}
			scope.Namespace = &identity.Namespace{Mode: "value", Value: value}
			ns = map[string]any{"mode": "value", "value": value}
			rowResource[identity.NamespaceKey] = value
		}
		question := map[string]any{"signal": "logs", "environmentSelector": wire, "fromMs": from, "toMs": from + 60000, "limit": 100, "where": map[string]any{"field": map[string]any{"kind": "column", "key": "body"}, "op": "contains", "value": "origin-only"}}
		expected := map[string]any{"environmentSelector": wire, "namespace": ns, "fromMs": from, "toMs": from + 60000}
		var nativeName any
		if selection.CurrentOnly() {
			nativeName = name
		}
		row := map[string]any{"ServiceName": service, "ResourceAttributes": rowResource, "timestamp_ns": fmt.Sprint((from + 1) * 1000000)}
		if err := runNativeHistogramJS(map[string]any{"q": question, "row": row, "resources": resources, "matches": matches, "expected": expected, "nativeName": nativeName}, serviceEnvironment318JS); err != nil {
			tc.Fatal(err)
			return
		}
		inventory, err := exploration.CompileInventory(scope)
		if err != nil {
			tc.Fatal(err)
			return
		}
		discovery, err := exploration.CompileServiceDiscovery(scope)
		if err != nil {
			tc.Fatal(err)
			return
		}
		if strings.Contains(inventory.Groups+discovery.Count+discovery.Groups, name) {
			tc.Fatal("environment interpolated into executable SQL")
			return
		}
		plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Scope: scope, Service: service, Signals: []string{"spans", "logs", "metrics"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}, HTTPServerObserved: true}, nil)
		if err != nil {
			tc.Fatal(err)
			return
		}
		if !selection.CurrentOnly() && (plan.NativeCapability == nil || plan.NativeCapability.State != "gauge-only" || plan.NativeCapability.Reason == "") {
			tc.Fatal("unsupported native evaluation disguised as absent observations")
			return
		}
		queries := []exploration.Query{scope.Query(), exploration.DashboardRecordQuestion(scope, service, "logs"), exploration.DashboardRecordQuestion(scope, service, "spans")}
		for _, widget := range plan.Widgets {
			if widget.Query != nil {
				queries = append(queries, *widget.Query)
			}
			if !selection.CurrentOnly() && widget.Measurement != nil {
				tc.Fatal("unsupported environment reached a native evaluator")
				return
			}
		}
		// Returned questions and plans must own the selector, not alias the caller.
		scope.EnvironmentSelector.Policy = "broken"
		scope.Namespace.Mode = "broken"
		for _, query := range append(queries, plan.Scope.Query()) {
			actual, err := query.EnvironmentSelection()
			if err != nil || !actual.Equal(selection) || query.Validate() != nil {
				tc.Fatal("service continuation lost or aliased its scope", err)
				return
			}
			if _, err := exploration.Compile(query); err != nil {
				tc.Fatal(err)
				return
			}
		}
	}, options(30)...)
}
