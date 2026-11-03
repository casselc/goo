package verification

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

const nativeEnvironment319JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import * as E from '../ui/record-environment.mjs';import * as C from '../ui/native-record-context.mjs';
import {sameNativeStream} from '../ui/measurement-scope.mjs';
const {result,contexts,point}=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(result);
const norm=c=>{c=structuredClone(c);c.question.offset??=0;if(c.question.environmentSelector&&c.question.environment==='')delete c.question.environment;return c};
for(const [index,series]of result.gauges.series.entries()){
 assert(E.matches(E.selection(result.question.definition),series.stream.resource));
 assert.deepEqual(norm(C.plan(result.question,series.stream,'logs',point,String(BigInt(point)+1n))),norm(contexts[index]));
 const bad=structuredClone(result.question);bad.definition.kind='counter';assert.throws(()=>C.plan(bad,series.stream,'logs',point,String(BigInt(point)+1n)));
}
if(result.gauges.series.length===2)assert(!sameNativeStream(result.gauges.series[0].stream,result.gauges.series[1].stream),'raw environment maps are distinct stream identity');
assert.deepEqual(result,before);
`

func TestHegelNativeGaugeEnvironment319(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"named", "empty", "unreported"}))
		name := fmt.Sprintf("prod-%d'雪", seed%13)
		scope := identity.EnvironmentScope{Mode: mode}
		if mode == "named" {
			scope.Value = name
		}
		selected, err := identity.EffectiveEnvironment(scope)
		if mode == "named" && hegel.Draw(tc, hegel.Booleans()) {
			selected, err = identity.CurrentEnvironment(name)
		}
		if err != nil {
			tc.Fatal(err)
			return
		}
		wire, _ := selected.Selector()
		ns := &identity.Namespace{Mode: "missing"}
		if seed%3 != 0 {
			ns.Mode = "value"
			if seed%3 == 2 {
				ns.Value = "team'雪"
			}
		}
		question, template := nrcInputs(ns, "gauge")
		question.Definition.Environment = ""
		question.Definition.EnvironmentSelector = &wire
		point := strconv.FormatInt(question.FromMS*1000000+1000000000+int64(seed), 10)
		variants := []map[string]string{
			{identity.EnvironmentKey: name}, {identity.LegacyEnvironmentKey: name},
			{identity.EnvironmentKey: "other", identity.LegacyEnvironmentKey: name},
			{identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: name}, {},
		}
		var rows, foreign []json.RawMessage
		expected := map[string]float64{}
		for index, labels := range variants {
			// Everything except the raw environment labels is deliberately identical.
			resource := map[string]string{"service.name": "api", "service.instance.id": "same-instance"}
			if ns.Mode == "value" {
				resource[identity.NamespaceKey] = ns.Value
			}
			for key, value := range labels {
				resource[key] = value
			}
			value := seed%97 + 100*index + 1
			raw, _ := json.Marshal(map[string]any{"resource_schema_url": "", "resource": resource, "scope": template.Scope, "version": template.Version, "schema_url": template.SchemaURL, "name": template.Name, "unit": template.Unit, "attributes": map[string]string{}, "start_ns": "0", "end_ns": point, "value": strconv.Itoa(value), "flags": 0})
			if selected.Matches(resource) {
				rows = append(rows, raw)
				expected[measurements.Hash(resource)] = float64(value)
			} else {
				foreign = append(foreign, raw)
			}
		}
		result, err := measurements.Evaluate(question, rows, time.UnixMilli(question.ToMS+1000))
		if err != nil || result.Gauges == nil || len(result.Gauges.Series) != len(expected) {
			tc.Fatal("typed gauge values were rejected or distinct raw streams merged", err)
			return
		}
		if err := measurements.ValidateResult(result); err != nil {
			tc.Fatal(err)
			return
		}
		if strings.Contains(result.Plan.SQL, name) {
			tc.Fatal("environment interpolated into native SQL")
			return
		}
		contexts := []*exploration.NativeRecordContext{}
		for _, series := range result.Gauges.Series {
			value, exists := expected[measurements.Hash(series.Stream.Resource)]
			if !exists || len(series.Samples) != 1 || series.Samples[0].Value == nil || *series.Samples[0].Value != value || series.Samples[0].AtNS != point {
				tc.Fatal("native values or complete stream identity changed")
				return
			}
			context, err := exploration.PlanNativeRecordContext(result.Question, series.Stream, "logs", point, strconv.FormatInt(question.FromMS*1000000+1000000001+int64(seed), 10))
			if err != nil {
				tc.Fatal(err)
				return
			}
			contexts = append(contexts, context)
		}
		for _, raw := range foreign {
			mixed := append(append([]json.RawMessage{}, rows...), raw)
			if _, err := measurements.Evaluate(question, mixed, time.UnixMilli(question.ToMS+1000)); err == nil {
				tc.Fatal("decoder accepted a source row outside the checked policy")
				return
			}
		}
		for _, kind := range []string{"counter", "histogram", "stock", "state", "lifecycle"} {
			unsupported := question
			unsupported.Definition.Kind = kind
			if _, err := measurements.Compile(unsupported); err == nil {
				tc.Fatal("unsupported typed family reached a source plan", kind)
				return
			}
		}
		if err := runNativeHistogramJS(map[string]any{"result": result, "contexts": contexts, "point": point}, nativeEnvironment319JS); err != nil {
			tc.Fatal(err)
			return
		}
		question.Definition.EnvironmentSelector.Policy = "changed"
		actual, err := result.Question.Definition.EnvironmentSelection()
		if err != nil || !actual.Equal(selected) {
			tc.Fatal("retained native result aliases its caller", err)
			return
		}
	}, options(30)...)
}
