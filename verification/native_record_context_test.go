package verification

import (
	"encoding/json"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

func nrcInputs(ns *identity.Namespace, kind string) (measurements.Question, metrics.Stream) {
	scope := exploration.InventoryQuestion{Environment: "prod", FromMS: 1790000000000, ToMS: 1790000006000, Namespace: ns}
	in := hdInstrument("context", "custom.work", kind, "1", scope, "api")
	p := measurements.Question{Definition: *in.Definition, FromMS: scope.FromMS, ToMS: scope.ToMS}
	s := metrics.Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "instance-one", "deployment.environment.name": "prod", "host.name": "do-not-copy"}, Scope: in.Scope, Version: in.Version, SchemaURL: in.SchemaURL, Name: in.Metric, Unit: in.Unit, Attributes: map[string]string{"metric.only": "not-a-record-selector"}}
	if ns != nil && ns.Mode == "value" {
		s.Resource[identity.NamespaceKey] = ns.Value
	}
	return p, s
}
func nrcPredicate(p *exploration.Predicate, values map[exploration.Field]string) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for i := range p.All {
			if !nrcPredicate(&p.All[i], values) {
				return false
			}
		}
		return true
	}
	if p.Field != nil && p.Field.Kind == "column" && p.Field.Key == "timestamp_ns" {
		a, ok := new(big.Int).SetString(values[*p.Field], 10)
		if !ok {
			return false
		}
		b, ok := new(big.Int).SetString(p.Value, 10)
		if !ok {
			return false
		}
		switch p.Op {
		case "gte":
			return a.Cmp(b) >= 0
		case "lte":
			return a.Cmp(b) <= 0
		}
		return false
	}
	return hdPredicate(p, values)
}
func TestHegelNativeRecordContextExactMembership(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(0, 2000000))
		width := hegel.Draw(tc, hegel.Integers(1, 2000001))
		// Epoch nanoseconds exceed2^53. Membership is checked with integer facts,
		// independently of implementation division and timestamp predicate assembly.
		from := int64(1790000000000000000) + int64(offset)
		to := from + int64(width)
		for _, ns := range []*identity.Namespace{nil, {Mode: "value", Value: ""}, {Mode: "value", Value: "commerce-π"}} {
			for _, kind := range []string{"counter", "histogram"} {
				for _, signal := range []string{"logs", "spans"} {
					p, s := nrcInputs(ns, kind)
					before, _ := json.Marshal([]any{p, s})
					got, e := exploration.PlanNativeRecordContext(p, s, signal, strconv.FormatInt(from, 10), strconv.FormatInt(to, 10))
					after, _ := json.Marshal([]any{p, s})
					if e != nil || string(before) != string(after) {
						tc.Errorf("mapping/input %v", e)
						return
					}
					divisor := big.NewInt(1000000)
					lo := new(big.Int).Div(big.NewInt(from), divisor)
					upper, rem := new(big.Int).QuoRem(big.NewInt(to), divisor, new(big.Int))
					if rem.Sign() != 0 {
						upper.Add(upper, big.NewInt(1))
					}
					if got.Question.FromMS != lo.Int64() || got.Question.ToMS != upper.Int64() || got.Question.Signal != signal || got.Question.Environment != "prod" || got.Question.Offset != 0 || got.Question.Limit != 100 || got.Grain != "interval-context" {
						tc.Errorf("context geometry/scope")
						return
					}
					var leaves []exploration.Field
					var walk func(*exploration.Predicate)
					walk = func(x *exploration.Predicate) {
						if x == nil {
							return
						}
						if x.Field != nil {
							leaves = append(leaves, *x.Field)
						}
						for i := range x.All {
							walk(&x.All[i])
						}
					}
					walk(got.Question.Where)
					wantLeaves := []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "resource", Key: identity.NamespaceKey}, {Kind: "resource", Key: "service.instance.id"}, {Kind: "column", Key: "timestamp_ns"}, {Kind: "column", Key: "timestamp_ns"}}
					if !reflect.DeepEqual(leaves, wantLeaves) {
						tc.Errorf("copied metric-only identity/dimensions")
						return
					}
					for _, at := range []int64{from - 1, from, from + 1, to - 1, to, to + 1} {
						for _, instance := range []string{"instance-one", "other", ""} {
							for _, resourceNS := range []*string{nil, lpcPointer(""), lpcPointer("commerce-π"), lpcPointer("other")} {
								fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "resource", Key: "service.instance.id"}: instance, {Kind: "column", Key: "timestamp_ns"}: strconv.FormatInt(at, 10)}
								if resourceNS != nil {
									fields[exploration.Field{Kind: "resource", Key: identity.NamespaceKey}] = *resourceNS
								}
								want := at >= from && at < to && instance == "instance-one"
								if ns == nil {
									want = want && resourceNS == nil
								} else {
									want = want && resourceNS != nil && *resourceNS == ns.Value
								}
								if nrcPredicate(got.Question.Where, fields) != want {
									tc.Errorf("exact contextual membership")
									return
								}
							}
						}
					}
				}
			}
		}
	}, options(24)...)
}
func TestNativeRecordContextRefusalsAndPoints(t *testing.T) {
	p, s := nrcInputs(nil, "counter")
	from := "1790000000000000001"
	to := "1790000000000000002"
	for _, bounds := range [][2]string{{"01790000000000000001", to}, {"+1790000000000000001", to}, {"1790000000000000001.0", to}, {from, from}, {to, from}, {"1789999999999999999", to}, {from, "1790000006000000001"}, {from, "18446744073709551615"}} {
		if _, e := exploration.PlanNativeRecordContext(p, s, "spans", bounds[0], bounds[1]); e == nil {
			t.Fatalf("invalid bounds accepted %v", bounds)
		}
	}
	for _, signal := range []string{"metrics", "", "Spans"} {
		if _, e := exploration.PlanNativeRecordContext(p, s, signal, from, to); e == nil {
			t.Fatal("invalid signal")
		}
	}
	for _, key := range []string{"service.name", "service.instance.id"} {
		bad := s
		bad.Resource = map[string]string{}
		for k, v := range s.Resource {
			bad.Resource[k] = v
		}
		delete(bad.Resource, key)
		if _, e := exploration.PlanNativeRecordContext(p, bad, "spans", from, to); e == nil {
			t.Fatal("missing established identity", key)
		}
	}
	pinned := p
	pinned.Definition.Source.ExactStream = &s
	other := s
	other.Attributes = map[string]string{"metric.only": "different"}
	if _, e := exploration.PlanNativeRecordContext(pinned, other, "logs", from, to); e == nil {
		t.Fatal("pinned stream replacement")
	}
	gauge := p
	gauge.Definition.Kind = "gauge"
	point, e := exploration.PlanNativeRecordContext(gauge, s, "logs", from, to)
	if e != nil || point.Grain != "point-context" {
		t.Fatal("pure exact gauge point", e)
	}
	if _, e := exploration.PlanNativeRecordContext(gauge, s, "logs", from, "1790000000000000003"); e == nil {
		t.Fatal("gauge lookaround inferred")
	}
	coarse := p
	coarse.ToMS = coarse.FromMS + 7200001
	if _, e := exploration.PlanNativeRecordContext(coarse, s, "spans", from, to); e == nil {
		t.Fatal("coarse >2h accepted")
	}
}
