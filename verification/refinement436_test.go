package verification

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"reflect"
	"strconv"
	"testing"
)

// Inspect all nested mutable allocations without serialization, which would
// erase nil/empty distinctions and can hide accidental integer normalization.
func refinement436Allocations(v reflect.Value, seen map[uintptr]bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			seen[v.Pointer()] = true
			refinement436Allocations(v.Elem(), seen)
		}
	case reflect.Slice:
		if v.Len() > 0 {
			seen[v.Pointer()] = true
		}
		for i := 0; i < v.Len(); i++ {
			refinement436Allocations(v.Index(i), seen)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			refinement436Allocations(v.Field(i), seen)
		}
	}
}
func refinement436Detached(tc *hegel.T, output, input any) {
	a, b := map[uintptr]bool{}, map[uintptr]bool{}
	refinement436Allocations(reflect.ValueOf(output), a)
	refinement436Allocations(reflect.ValueOf(input), b)
	for p := range a {
		if b[p] {
			tc.Fatal("mutable allocation retained")
		}
	}
}
func TestHegelRefinement436Conservation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"logs", "spans"}))
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		typ := hegel.Draw(tc, hegel.SampledFrom([]string{"int64", "uint64", "decimal64"}))
		binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "amount"}, Type: typ, Unit: "items", Parse: "canonical-decimal"}
		threshold := "1"
		if typ == "decimal64" {
			binding.Scale = 2
			threshold = "1.25"
		}
		condition := &exploration.Predicate{All: []exploration.Predicate{{Numeric: &binding, Op: "gte", Value: threshold}, {Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "region"}, Op: "missing"}, {Field: &exploration.Field{Kind: "resource", Key: "region"}, Op: "eq", Value: "west"}}}}}
		column := "severity_number"
		if signal == "spans" {
			column = "duration_ns"
		}
		parent := exploration.Query{Signal: signal, Environment: "prod", FromMS: int64(seed), ToMS: int64(seed + 60000), Limit: hegel.Draw(tc, hegel.Integers(1, 100)), Offset: hegel.Draw(tc, hegel.Integers(1, 1000)), GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Numeric: &binding, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}, Where: &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}, {All: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "route"}, Op: "exists"}, {Field: &exploration.Field{Kind: "column", Key: column}, Op: "gte", Value: "01"}}}}}}
		if signal == "spans" {
			parent.Summary = "span-duration"
			parent.SpanStatus = true
		}
		if hegel.Draw(tc, hegel.Booleans()) {
			parent.Environment = ""
			scope := hegel.Draw(tc, hegel.SampledFrom([]identity.EnvironmentScope{{Mode: "named", Value: "prod"}, {Mode: "empty"}, {Mode: "unreported"}}))
			parent.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: scope}
		}
		if signal == "spans" && hegel.Draw(tc, hegel.Booleans()) {
			parent.ReportedEdge = &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: "api", Namespace: &identity.Namespace{Mode: "value", Value: "shop"}}, Child: exploration.EdgeService{Service: "worker", Namespace: &identity.Namespace{Mode: "value", Value: "shop"}}, DiscoveryNamespace: &identity.Namespace{Mode: "value", Value: "shop"}}
		}
		before, err := exploration.CloneQuery(parent)
		if err != nil {
			tc.Fatal(err)
		}
		beforeCondition, err := exploration.ClonePredicate(condition, signal)
		if err != nil {
			tc.Fatal(err)
		}
		out, err := exploration.Refine(parent, condition)
		if err != nil {
			tc.Fatal(err)
		}
		if !reflect.DeepEqual(parent, before) || !reflect.DeepEqual(condition, beforeCondition) {
			tc.Fatal("inputs mutated")
		}
		policy := out
		policy.Where = parent.Where
		if !reflect.DeepEqual(policy, parent) {
			tc.Fatal("caller policy rewritten")
		}
		if len(out.Where.All) != 2 || !reflect.DeepEqual(out.Where.All[0], *parent.Where) || !reflect.DeepEqual(out.Where.All[1], *condition) {
			tc.Fatal("whole parent OR or condition changed")
		}
		refinement436Detached(tc, out, parent)
		refinement436Detached(tc, out, condition)
		refinement436Detached(tc, before, parent)
		refinement436Detached(tc, beforeCondition, condition)
		identity, err := exploration.Refine(parent, nil)
		if err != nil || !reflect.DeepEqual(identity, parent) {
			tc.Fatal("nil refinement changed identity", err)
		}
		refinement436Detached(tc, identity, parent)
		for _, service := range []string{"api", "worker"} {
			for _, raw := range []string{"", "invalid", "0", "1", "1.25", "2", "9007199254740993"} {
				for _, region := range []string{"<absent>", "", "west", "east"} {
					values := map[exploration.Field]string{{Kind: "column", Key: "service"}: service, {Kind: "column", Key: column}: strconv.Itoa(seed % 3), {Kind: "attribute", Key: "amount"}: raw}
					if seed%2 == 0 {
						values[exploration.Field{Kind: "attribute", Key: "route"}] = ""
					}
					if region != "<absent>" {
						values[exploration.Field{Kind: "resource", Key: "region"}] = region
					}
					if exploration.Match(out.Where, values) != (exploration.Match(parent.Where, values) && exploration.Match(condition, values)) {
						tc.Fatal("membership differs from parent AND condition")
					}
				}
			}
		}
	}, options(160)...)
}
