package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
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

const observationRadius388 int64 = 60000000000

type observationFixture388 struct {
	Result *measurements.Result
	Stream metrics.Stream
	AtNS   string
	Plans  []*exploration.NativeObservationContext
}

func observationEnv388(selection *identity.EnvironmentSelector, resource map[string]string) bool {
	value, present := resource[identity.EnvironmentKey]
	if !present && selection.Policy == "effective" {
		value, present = resource[identity.LegacyEnvironmentKey]
	}
	switch selection.Scope.Mode {
	case "named":
		return present && value == selection.Scope.Value
	case "empty":
		return present && value == ""
	case "unreported":
		return !present
	}
	return false
}
func observationBounds388(lo, hi, at int64) (int64, int64) {
	return max(lo, at-observationRadius388), min(hi, at+observationRadius388)
}
func observationResult388(t *testing.T, seed, mode, namespace int, lo, hi, at int64, unknown bool) (*measurements.Result, metrics.Stream) {
	t.Helper()
	var ns *identity.Namespace
	if namespace == 0 {
		ns = &identity.Namespace{Mode: "missing"}
	} else {
		ns = &identity.Namespace{Mode: "value"}
		if namespace == 2 {
			ns.Value = "team'雪"
		}
	}
	q, s := nrcInputs(ns, "gauge")
	if seed%3 == 0 {
		q.Definition.Source.Namespace = nil
	} // Projection still pins the selected stream namespace.
	q.FromMS = lo / 1000000
	q.ToMS = (hi + 999999) / 1000000
	q.FromNS = strconv.FormatInt(lo, 10)
	q.ToNS = strconv.FormatInt(hi, 10)
	q.Definition.MaxGapMS = 3000
	q.Definition.Environment = ""
	selection := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
	switch mode {
	case 1:
		selection.Policy = "effective"
	case 2:
		selection.Policy = "effective"
		selection.Scope = identity.EnvironmentScope{Mode: "empty"}
	case 3:
		selection.Policy = "effective"
		selection.Scope = identity.EnvironmentScope{Mode: "empty"}
	case 4:
		selection.Policy = "effective"
		selection.Scope = identity.EnvironmentScope{Mode: "unreported"}
	case 5:
		selection.Policy = "effective"
	}
	q.Definition.EnvironmentSelector = selection
	delete(s.Resource, identity.EnvironmentKey)
	delete(s.Resource, identity.LegacyEnvironmentKey)
	switch mode {
	case 0:
		s.Resource[identity.EnvironmentKey] = "prod"
		s.Resource[identity.LegacyEnvironmentKey] = "foreign"
	case 1:
		s.Resource[identity.LegacyEnvironmentKey] = "prod"
	case 2:
		s.Resource[identity.EnvironmentKey] = ""
		s.Resource[identity.LegacyEnvironmentKey] = "prod"
	case 3:
		s.Resource[identity.LegacyEnvironmentKey] = ""
	case 5:
		s.Resource[identity.EnvironmentKey] = "prod"
		s.Resource[identity.LegacyEnvironmentKey] = "foreign"
	}
	s.Resource["host.name"] = "not-a-record-selector"
	s.Attributes["instrument.only"] = "not-a-record-selector"
	q.Definition.Source.Attributes = map[string]string{"metric.only": s.Attributes["metric.only"]}
	if seed%2 == 0 {
		q.Definition.Source.ExactStream = &s
	}
	flags := 0
	if unknown {
		flags = 1
	}
	raw, _ := json.Marshal(map[string]any{"resource_schema_url": s.ResourceSchemaURL, "resource": s.Resource, "scope": s.Scope, "version": s.Version, "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "attributes": s.Attributes, "start_ns": "0", "end_ns": strconv.FormatInt(at, 10), "value": "0", "flags": flags})
	result, e := measurements.Evaluate(q, []json.RawMessage{raw}, time.UnixMilli(q.ToMS+1))
	if e != nil || result.Gauges == nil || len(result.Gauges.Series) != 1 {
		t.Fatal("generated valid gauge rejected", e)
	}
	return result, result.Gauges.Series[0].Stream
}
func TestHegelNativeObservationContext388(t *testing.T) {
	fixtures := []observationFixture388{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.Integers(0, 5))
		namespace := hegel.Draw(tc, hegel.Integers(0, 2))
		width := hegel.Draw(tc, hegel.SampledFrom([]int64{1, 59000000123, 121000000123, 180000000123, 300000000123}))
		lo := int64(1790000000000000000) + int64(seed)
		if seed%7 == 0 {
			lo = 0
		}
		hi := lo + width
		for _, at := range []int64{lo, lo + width/2, hi - 1} {
			result, stream := observationResult388(t, seed, mode, namespace, lo, hi, at, seed%2 == 0)
			before := continuationJSON352(result)
			streamBefore := continuationJSON352(stream)
			from, to := observationBounds388(lo, hi, at)
			fixture := observationFixture388{Result: result, Stream: stream, AtNS: strconv.FormatInt(at, 10)}
			for _, signal := range []string{"logs", "spans"} {
				plan, e := exploration.PlanNativeObservationContext(result, stream, signal, fixture.AtNS)
				if e != nil {
					tc.Fatal(e)
				}
				fixture.Plans = append(fixture.Plans, plan)
				if plan.FromNS != strconv.FormatInt(from, 10) || plan.ToNS != strconv.FormatInt(to, 10) || plan.AtNS != fixture.AtNS || plan.RadiusNS != "60000000000" || plan.Grain != "observation-context" || plan.Relation != "observation-context" || plan.Question.Signal != signal || plan.Question.FromMS != from/1000000 || plan.Question.ToMS != (to+999999)/1000000 || plan.Question.Limit != 100 || plan.Question.Offset != 0 || !reflect.DeepEqual(plan.Question.EnvironmentSelector, result.Question.Definition.EnvironmentSelector) {
					tc.Fatal("lookaround changed integer geometry/identity")
				}
				// Only shared service/namespace/instance identity and exact time are projected.
				leaves := []exploration.Field{}
				var walk func(*exploration.Predicate)
				walk = func(p *exploration.Predicate) {
					if p == nil {
						return
					}
					if p.Field != nil {
						leaves = append(leaves, *p.Field)
					}
					for i := range p.All {
						walk(&p.All[i])
					}
				}
				walk(plan.Question.Where)
				wantLeaves := []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "resource", Key: identity.NamespaceKey}, {Kind: "resource", Key: "service.instance.id"}, {Kind: "column", Key: "timestamp_ns"}, {Kind: "column", Key: "timestamp_ns"}}
				if !reflect.DeepEqual(leaves, wantLeaves) {
					tc.Fatal("metric-only/instrumentation/unrelated resource filters leaked")
				}
				envs := []map[string]string{{identity.EnvironmentKey: "prod"}, {identity.LegacyEnvironmentKey: "prod"}, {identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: "prod"}, {identity.EnvironmentKey: "foreign", identity.LegacyEnvironmentKey: "prod"}, {identity.EnvironmentKey: ""}, {identity.LegacyEnvironmentKey: ""}, {}}
				for _, stamp := range []int64{from - 1, from, at, to - 1, to, to + 1} {
					if stamp < 0 {
						continue
					}
					for _, instance := range []*string{nil, lpcPointer(""), lpcPointer(stream.Resource["service.instance.id"]), lpcPointer("foreign")} {
						for _, ns := range []*string{nil, lpcPointer(""), lpcPointer("team'雪"), lpcPointer("foreign")} {
							for _, environment := range envs {
								fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: stream.Resource["service.name"], {Kind: "column", Key: "timestamp_ns"}: strconv.FormatInt(stamp, 10)}
								if instance != nil {
									fields[exploration.Field{Kind: "resource", Key: "service.instance.id"}] = *instance
								}
								if ns != nil {
									fields[exploration.Field{Kind: "resource", Key: identity.NamespaceKey}] = *ns
								}
								actualNS, presentNS := stream.Resource[identity.NamespaceKey]
								namespaceMatches := ns == nil && !presentNS || ns != nil && presentNS && *ns == actualNS
								want := stamp >= from && stamp < to && instance != nil && *instance == stream.Resource["service.instance.id"] && namespaceMatches && observationEnv388(result.Question.Definition.EnvironmentSelector, environment)
								got := nrcPredicate(plan.Question.Where, fields) && observationEnv388(plan.Question.EnvironmentSelector, environment)
								if got != want {
									tc.Fatal("independent exact record membership differs")
								}
								fields[exploration.Field{Kind: "column", Key: "service"}] = "foreign"
								if nrcPredicate(plan.Question.Where, fields) {
									tc.Fatal("foreign service admitted")
								}
							}
						}
					}
				}
			}
			if continuationJSON352(result) != before || continuationJSON352(stream) != streamBefore {
				tc.Fatal("planning mutated retained evidence")
			}
			for _, mutate := range []func(*measurements.Result){
				func(r *measurements.Result) { r.Question.Definition.Kind = "counter" },
				func(r *measurements.Result) { r.Question.Definition.Grain = "consumer-group" },
				func(r *measurements.Result) {
					r.Question.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "empty"}}
				},
				func(r *measurements.Result) { r.Gauges.Series = append(r.Gauges.Series, r.Gauges.Series[0]) },
				func(r *measurements.Result) {
					r.Gauges.Series[0].Samples = append(r.Gauges.Series[0].Samples, r.Gauges.Series[0].Samples[0])
				},
				func(r *measurements.Result) { r.Gauges.Series[0].Samples[0].State = "invented" },
				func(r *measurements.Result) { r.Gauges.Series[0].Samples[0].AtNS = "0" + fixture.AtNS },
			} {
				var bad measurements.Result
				wire, _ := json.Marshal(result)
				json.Unmarshal(wire, &bad)
				mutate(&bad)
				if _, e := exploration.PlanNativeObservationContext(&bad, stream, "logs", fixture.AtNS); e == nil {
					tc.Fatal("common malformed selected occurrence admitted")
				}
			}
			for _, signal := range []string{"metrics", "", "Spans"} {
				if _, e := exploration.PlanNativeObservationContext(result, stream, signal, fixture.AtNS); e == nil {
					tc.Fatal("unsupported record signal admitted")
				}
			}
			for _, stamp := range []string{"0" + fixture.AtNS, strconv.FormatInt(at+1, 10), "+1", "not-reported"} {
				if _, e := exploration.PlanNativeObservationContext(result, stream, "logs", stamp); e == nil {
					tc.Fatal("fabricated/noncanonical timestamp admitted")
				}
			}
			var foreign metrics.Stream
			wire, _ := json.Marshal(stream)
			json.Unmarshal(wire, &foreign)
			foreign.Resource["service.instance.id"] = "foreign"
			if _, e := exploration.PlanNativeObservationContext(result, foreign, "logs", fixture.AtNS); e == nil {
				tc.Fatal("foreign unretained stream admitted")
			}
			fixtures = append(fixtures, fixture)
			// Full Go admission also verifies evidence replay/integrity, unlike JS's
			// bounded selected-sample projection over the ordinary owner's checked DTO.
			for _, mutate := range []func(*measurements.Result){func(r *measurements.Result) { r.Basis = "invalid" }, func(r *measurements.Result) { r.Count = "2" }, func(r *measurements.Result) { r.Rows[0] = json.RawMessage(`{}`) }} {
				var bad measurements.Result
				wire, _ := json.Marshal(result)
				json.Unmarshal(wire, &bad)
				mutate(&bad)
				if _, e := exploration.PlanNativeObservationContext(&bad, stream, "logs", fixture.AtNS); e == nil {
					tc.Fatal("corrupt replay admitted")
				}
			}
		}
	}, options(32)...)
	wire, _ := json.Marshal(fixtures)
	cmd := exec.CommandContext(t.Context(), "node", "native_observation_context_388_fixture.mjs")
	cmd.Stdin = bytes.NewReader(wire)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("observation projection parity: %v\n%s", e, out)
	}
	var observed struct{ Cases, Plans int }
	if json.Unmarshal(out, &observed) != nil || observed.Cases != 96 || observed.Plans != 192 {
		t.Fatalf("incomplete observation projection %s", out)
	}
}
