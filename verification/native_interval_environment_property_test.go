package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

const nativeInterval320JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import * as C from '../ui/native-record-context.mjs';import * as H from '../ui/measurement-scope.mjs';
const {result,cases}=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(result);
const norm=x=>{x=structuredClone(x);x.question.offset??=0;if(x.question.environmentSelector&&x.question.environment==='')delete x.question.environment;return x};
for(const c of cases){assert.deepEqual(norm(C.plan(result.question,c.stream,'logs',c.from,c.to)),norm(c.context));if(result.question.definition.kind==='histogram'){const s={kind:'measurement-interval',question:result.question,stream:c.stream,fromNS:c.from,toNS:c.to};H.validateMeasurementIntervalSelector(s);const pinned=structuredClone(s);pinned.question.definition.source.exactStream=structuredClone(c.stream);H.validateMeasurementIntervalSelector(pinned);const changed=structuredClone(pinned);changed.stream.attributes['unexpected']='changed';assert.throws(()=>H.validateMeasurementIntervalSelector(changed));const bad=structuredClone(s);bad.stream.resource['deployment.environment.name']='foreign';assert.throws(()=>H.validateMeasurementIntervalSelector(bad));}}
const series=result.counters?.series??result.histograms.series;if(series.length===2)assert(!H.sameNativeStream(series[0].stream,series[1].stream));assert.deepEqual(result,before);
`

func TestHegelNativeCounterEnvironment320(t *testing.T) { nativeIntervalEnvironment320(t, "counter") }
func TestHegelNativeHistogramEnvironment320(t *testing.T) {
	nativeIntervalEnvironment320(t, "histogram")
}
func nativeIntervalEnvironment320(t *testing.T, kind string) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"current", "named", "empty", "unreported"}))
		name := fmt.Sprintf("prod-%d'雪", seed%17)
		scope := identity.EnvironmentScope{Mode: mode}
		if mode == "current" || mode == "named" {
			scope.Mode = "named"
			scope.Value = name
		}
		selection, e := identity.EffectiveEnvironment(scope)
		if mode == "current" {
			selection, e = identity.CurrentEnvironment(name)
		}
		if e != nil {
			tc.Fatal(e)
			return
		}
		selector, _ := selection.Selector()
		ns := &identity.Namespace{Mode: "missing"}
		if seed%3 != 0 {
			ns.Mode = "value"
			if seed%3 == 2 {
				ns.Value = "team'雪"
			}
		}
		q, stream := nrcInputs(ns, kind)
		q.Definition.Environment = ""
		q.Definition.EnvironmentSelector = &selector
		if kind == "histogram" {
			q.Definition.Source.Unit = hegel.Draw(tc, hegel.SampledFrom([]string{"s", "ms"}))
			stream.Unit = q.Definition.Source.Unit
		}
		start := q.FromMS * 1000000
		delta := seed%37 + 1
		temporality := 1
		if kind == "counter" || seed%2 == 0 {
			temporality = 2
		}
		variants := []map[string]string{{identity.EnvironmentKey: name}, {identity.LegacyEnvironmentKey: name}, {identity.EnvironmentKey: "other", identity.LegacyEnvironmentKey: name}, {identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: name}, {}}
		var rows, foreign []json.RawMessage
		want := map[string]bool{}
		for _, labels := range variants {
			resource := map[string]string{"service.name": "api", "service.instance.id": "same-instance"}
			if ns.Mode == "value" {
				resource[identity.NamespaceKey] = ns.Value
			}
			for k, v := range labels {
				resource[k] = v
			}
			for j := 1; j <= 3; j++ {
				r := map[string]any{"resource": resource, "scope": stream.Scope, "version": stream.Version, "schema_url": stream.SchemaURL, "name": stream.Name, "unit": stream.Unit, "attributes": map[string]string{}, "start_ns": strconv.FormatInt(start, 10), "end_ns": strconv.FormatInt(start+int64(j)*1000000000, 10), "temporality": temporality, "monotonic": true, "flags": 0}
				if kind == "counter" {
					v := 100 + delta*(j-1)
					if j == 3 {
						v = 1
					}
					r["value"] = strconv.Itoa(v)
				} else {
					n := j
					if temporality == 1 {
						r["start_ns"] = strconv.FormatInt(start+int64(j-1)*1000000000, 10)
						n = 1
					}
					r["count"] = strconv.Itoa(2 * n)
					r["bounds"] = []float64{1}
					r["bucket_counts"] = []string{strconv.Itoa(n), strconv.Itoa(n)}
				}
				raw, _ := json.Marshal(r)
				if selection.Matches(resource) {
					rows = append(rows, raw)
					want[measurements.Hash(resource)] = true
				} else {
					foreign = append(foreign, raw)
				}
			}
		}
		before, _ := json.Marshal(q)
		result, e := measurements.Evaluate(q, rows, time.UnixMilli(q.ToMS+1000))
		if e != nil {
			tc.Fatal(e)
			return
		}
		after, _ := json.Marshal(q)
		if string(before) != string(after) || measurements.ValidateResult(result) != nil {
			tc.Fatal("question mutated or invalid retained result")
			return
		}
		cases := []map[string]any{}
		seen := map[string]bool{}
		add := func(s metrics.Stream, from, to string) {
			pinned := result.Question
			pinned.Definition.Source.ExactStream = &s
			if err := pinned.Validate(); err != nil {
				tc.Fatal(err)
				return
			}
			altered := s
			altered.Attributes = map[string]string{"unexpected": "changed"}
			if _, err := exploration.PlanNativeRecordContext(pinned, altered, "logs", from, to); err == nil {
				tc.Fatal("pinned full stream was replaced")
				return
			}
			ctx, err := exploration.PlanNativeRecordContext(result.Question, s, "logs", from, to)
			if err != nil {
				tc.Fatal(err)
				return
			}
			cases = append(cases, map[string]any{"stream": s, "from": from, "to": to, "context": ctx})
		}
		if kind == "counter" {
			if result.Counters == nil {
				tc.Fatal("missing counters")
				return
			}
			for _, s := range result.Counters.Series {
				seen[measurements.Hash(s.Stream.Resource)] = true
				valid, unknown := false, false
				for _, i := range s.Intervals {
					if i.FromNS == strconv.FormatInt(start+1000000000, 10) && i.ToNS == strconv.FormatInt(start+2000000000, 10) {
						if i.State != "ok" || i.Delta != strconv.Itoa(delta) || i.PerSecond == nil || *i.PerSecond != float64(delta) {
							tc.Fatal("exact one-second delta/rate changed")
							return
						}
						valid = true
						add(s.Stream, i.FromNS, i.ToNS)
					}
					if i.FromNS == strconv.FormatInt(start+2000000000, 10) {
						if i.State == "ok" || i.PerSecond != nil || i.Delta != "" {
							tc.Fatal("decrease fabricated rate")
							return
						}
						unknown = true
					}
				}
				if !valid || !unknown {
					tc.Fatal("missing valid/decrease interval")
					return
				}
			}
		} else {
			if result.Histograms == nil {
				tc.Fatal("missing histograms")
				return
			}
			for _, s := range result.Histograms.Series {
				seen[measurements.Hash(s.Stream.Resource)] = true
				valid := 0
				for _, i := range s.Intervals {
					if i.State != "ok" {
						continue
					}
					d := i.Distribution
					if d == nil || d.Count != "2" || d.Unit != stream.Unit || !reflect.DeepEqual(d.Bounds, []float64{1}) || !reflect.DeepEqual(d.BucketCounts, []string{"1", "1"}) {
						tc.Fatal("distribution arithmetic/unit changed")
						return
					}
					valid++
					add(s.Stream, i.FromNS, i.ToNS)
				}
				if valid == 0 {
					tc.Fatal("no histogram intervals")
					return
				}
			}
		}
		if !reflect.DeepEqual(seen, want) {
			tc.Fatal("raw environment stream identity merged/changed")
			return
		}
		for _, raw := range foreign {
			mixed := append(append([]json.RawMessage{}, rows...), raw)
			if _, err := measurements.Evaluate(q, mixed, time.UnixMilli(q.ToMS+1000)); err == nil {
				tc.Fatal("foreign policy row admitted")
				return
			}
		}
		if e = runNativeHistogramJS(map[string]any{"result": result, "cases": cases}, nativeInterval320JS); e != nil {
			tc.Fatal(e)
			return
		}
		q.Definition.EnvironmentSelector.Policy = "changed"
		actual, e := result.Question.Definition.EnvironmentSelection()
		if e != nil || !actual.Equal(selection) {
			tc.Fatal("result aliases caller policy")
			return
		}
	}, options(30)...)
}
