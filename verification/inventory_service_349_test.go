package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

func TestHegelInventoryServiceScope349(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		serviceMode := hegel.Draw(tc, hegel.Integers(0, 2))
		namespaceMode := hegel.Draw(tc, hegel.Integers(0, 3))
		environmentMode := hegel.Draw(tc, hegel.Integers(0, 4))
		service := fmt.Sprintf("api-'\\雪-%d", seed)
		if serviceMode == 1 {
			service = ""
		}
		q := exploration.InventoryQuestion{Environment: "prod-'雪", FromMS: 1800000000000 + int64(seed)*1000, ToMS: 1800000060000 + int64(seed)*1000}
		if serviceMode != 0 {
			q.Service = new(service)
		}
		switch namespaceMode {
		case 1:
			q.Namespace = &identity.Namespace{Mode: "missing"}
		case 2:
			q.Namespace = &identity.Namespace{Mode: "value"}
		case 3:
			q.Namespace = &identity.Namespace{Mode: "value", Value: "team-'雪"}
		}
		if environmentMode > 0 {
			q.Environment = ""
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-'雪"}}
			switch environmentMode {
			case 1:
				q.EnvironmentSelector.Policy = "current-only"
			case 3:
				q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			case 4:
				q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		raw, err := json.Marshal(map[string]any{"scope": q, "service": service, "name": "operation-'雪"})
		if err != nil {
			tc.Fatal(err)
			return
		}
		script := `import fs from 'node:fs';import {inventoryQuestion,inventoryMatches,question} from '../ui/inventory.mjs';const input=JSON.parse(fs.readFileSync(0,'utf8')),scope=structuredClone(input.scope),request=inventoryQuestion(scope),row={service:input.service,name:input.name},result={question:structuredClone(request),rows:[row]},continuation=question(result,row),accepted=inventoryMatches(request,result),changed=structuredClone(result);if(Object.hasOwn(changed.question,'service'))delete changed.question.service;else changed.question.service='';const rejectedPresence=!inventoryMatches(request,changed),foreign={...result,rows:[{...row,service:'foreign-service'}]},foreignAccepted=inventoryMatches(request,foreign);scope.service='mutated';if(scope.namespace)scope.namespace.value='mutated';if(scope.environmentSelector)scope.environmentSelector.scope={mode:'named',value:'mutated'};const bad=structuredClone(result);bad.question.toMs++;const rejectedWindow=!inventoryMatches(request,bad);process.stdout.write(JSON.stringify({request,continuation,accepted,rejectedPresence,foreignAccepted,rejectedWindow}));`
		cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var got struct {
			Request                                                     exploration.InventoryQuestion
			Continuation                                                exploration.Query
			Accepted, RejectedPresence, ForeignAccepted, RejectedWindow bool
		}
		if err = json.Unmarshal(out, &got); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if !reflect.DeepEqual(got.Request, q) || !got.Accepted || !got.RejectedPresence || !got.RejectedWindow || got.ForeignAccepted != (q.Service == nil) {
			tc.Fatal("UI lost service selection presence, exact echo, scope isolation or caller detachment", got, q)
			return
		}
		plan, err := exploration.CompileInventory(got.Request)
		if err != nil {
			tc.Fatal(err)
			return
		}
		if !strings.Contains(plan.Groups, "LIMIT 1001") || strings.Contains(plan.Groups, service) && service != "" {
			tc.Fatal("service cap or typed literal safety changed")
			return
		}
		if q.Service != nil && !strings.Contains(plan.Groups, "ServiceName = {") {
			tc.Fatal("service applied after aggregate or not at all")
			return
		}
		if q.Service == nil && strings.Contains(plan.Groups, "ServiceName = {") {
			tc.Fatal("omitted service became selected")
			return
		}
		// A single inventory operation always becomes an exact-service/name record
		// question, even when its parent inventory intentionally covers every service.
		where := []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: service}}
		if q.Namespace != nil {
			p := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: identity.NamespaceKey}, Op: "eq", Value: q.Namespace.Value}
			if q.Namespace.Mode == "missing" {
				p.Op = "missing"
			}
			where = append(where, p)
		}
		where = append(where, exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "name"}, Op: "eq", Value: "operation-'雪"})
		wanted := exploration.Query{Signal: "spans", Environment: q.Environment, EnvironmentSelector: q.EnvironmentSelector, FromMS: q.FromMS, ToMS: q.ToMS, Where: &exploration.Predicate{All: where}, GroupBy: []exploration.Field{}, Limit: 100, Offset: 0}
		if !reflect.DeepEqual(got.Continuation, wanted) {
			tc.Fatal("operation continuation lost exact service/namespace/window/policy", got.Continuation, wanted)
			return
		}
		if _, err = exploration.Compile(got.Continuation); err != nil {
			tc.Fatal(err)
			return
		}
		if q.Service != nil {
			original := *q.Service
			copy := exploration.CaptureQuestion{Kind: "inventory", Inventory: &q}.Normalized()
			*q.Service = "caller mutation"
			if copy.Inventory.Service == nil || *copy.Inventory.Service != original {
				tc.Fatal("normalized custody question aliases service pointer")
				return
			}
		}
	}, options(40)...)
}
