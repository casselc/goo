package verification

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

const serviceHandoffJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as H from '../ui/service-handoff.mjs';const {q,row,expected}=JSON.parse(fs.readFileSync(0,'utf8'));const before=structuredClone({q,row}),intent=H.create(q,row);assert.deepEqual(H.scope(intent),expected);assert.deepEqual(intent.origin,q);assert.equal(intent.service,row.ServiceName);assert.deepEqual({q,row},before);assert(Object.isFrozen(intent.origin));for(const policy of ['effective','unknown']){const bad=structuredClone(q);delete bad.environment;bad.environmentSelector={version:'goo/environment-selector/v1',policy,scope:{mode:'named',value:expected.environment}};assert.throws(()=>H.create(bad,row));}for(const change of [r=>delete r.ResourceAttributes,r=>r.ResourceAttributes=null,r=>r.ServiceName='',r=>r.timestamp_ns=String(BigInt(q.toMs)*1000000n),r=>r.ResourceAttributes['service.namespace']=null]){const bad=structuredClone(row);change(bad);assert.throws(()=>H.create(q,bad));}for(const change of [x=>x.extra=true,x=>x.fromMs++,x=>x.environmentSelector.policy='effective',x=>x.namespace={mode:'missing',value:''}]){const bad=structuredClone(intent);change(bad);assert.throws(()=>H.validate(bad));}`

func TestHegelObservedServiceHandoffPreservesExecutableScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		from := int64(seed) * 1000
		env := fmt.Sprintf("prod-%d'雪", seed%13)
		service := fmt.Sprintf("api-%d", seed%17)
		scope := exploration.InventoryQuestion{Environment: env, FromMS: from, ToMS: from + 60000}
		resource := map[string]string{"deployment.environment.name": env, "deployment.environment": "ignored-legacy"}
		switch seed % 3 {
		case 0:
			scope.Namespace = &identity.Namespace{Mode: "missing"}
		case 1:
			scope.Namespace = &identity.Namespace{Mode: "value"}
			resource[identity.NamespaceKey] = ""
		default:
			scope.Namespace = &identity.Namespace{Mode: "value", Value: fmt.Sprintf("team-%d'雪", seed%7)}
			resource[identity.NamespaceKey] = scope.Namespace.Value
		}
		expectedNS := map[string]any{"mode": scope.Namespace.Mode}
		if scope.Namespace.Mode == "value" {
			expectedNS["value"] = scope.Namespace.Value
		}
		expected := map[string]any{"environment": env, "namespace": expectedNS, "fromMs": from, "toMs": from + 60000}
		q := map[string]any{"signal": "logs", "fromMs": from, "toMs": from + 60000, "limit": 100, "offset": 0, "where": map[string]any{"field": map[string]any{"kind": "column", "key": "body"}, "op": "contains", "value": "origin-only"}}
		selected, err := identity.CurrentEnvironment(env)
		if err != nil {
			tc.Fatal(err)
			return
		}
		selector, err := selected.Selector()
		if err != nil {
			tc.Fatal(err)
			return
		}
		if seed%2 == 0 {
			q["environment"] = env
		} else {
			q["environmentSelector"] = selector
		}
		row := map[string]any{"ServiceName": service, "ResourceAttributes": resource, "timestamp_ns": fmt.Sprint((from + int64(seed%60000)) * 1000000)}
		// The JS constructor must produce exactly the scope accepted by existing Go
		// service discovery, with no origin log predicate carried across signals.
		if err := runNativeHistogramJS(map[string]any{"q": q, "row": row, "expected": expected}, serviceHandoffJS); err != nil {
			tc.Fatal(err)
			return
		}
		if _, err := exploration.CompileInventory(scope); err != nil {
			tc.Fatal(err)
			return
		}
		if _, err := exploration.CompileServiceDiscovery(scope); err != nil {
			tc.Fatal(err)
			return
		}
		for _, signal := range []string{"logs", "spans"} {
			legacy := exploration.DashboardRecordQuestion(scope, service, signal)
			typed := legacy
			typed.Environment = ""
			typed.EnvironmentSelector = &selector
			if !legacy.SameEnvironment(typed) {
				tc.Fatal("environment policy changed")
				return
			}
			a, err := exploration.Compile(legacy)
			if err != nil {
				tc.Fatal(err)
				return
			}
			b, err := exploration.Compile(typed)
			if err != nil {
				tc.Fatal(err)
				return
			}
			if !reflect.DeepEqual(a, b) {
				tc.Fatal("equivalent scope changed executable plan")
				return
			}
			if legacy.Where == nil {
				tc.Fatal("service/namespace identity lost")
				return
			}
		}
	}, options(30)...)
}
