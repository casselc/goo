package metrics

import (
	"math"
	"reflect"
	"testing"
)

func TestExplicitHostKernelsDoNotRelaxServiceKernels(t *testing.T) {
	stream := Stream{Resource: map[string]string{"host.id": "one"}, Scope: "hostmetrics", Name: "utilization", Unit: "1"}
	gauges := []GaugePoint{{Stream: stream, EndNS: "10", Value: 0.5}}
	if _, e := EvaluateGauges("0", "20", "20", gauges); e == nil {
		t.Fatal("service gauge identity relaxed")
	}
	r, e := EvaluateHostGauges("0", "20", "20", "one", gauges)
	if e != nil || r.Series[0].Samples[0].Value == nil || *r.Series[0].Samples[0].Value != 0.5 {
		t.Fatal(r, e)
	}
	if _, e := EvaluateHostGauges("0", "20", "20", "other", gauges); e == nil {
		t.Fatal("host identity relaxed")
	}
	stocks := []StockPoint{{Stream: stream, Temporality: "cumulative", StartNS: "0", EndNS: "10", Value: 1024}}
	if _, e := EvaluateStocks("0", "20", "20", stocks); e == nil {
		t.Fatal("service stock identity relaxed")
	}
	s, e := EvaluateHostStocks("0", "20", "20", "one", stocks)
	if e != nil || *s.Series[0].Samples[0].Value != 1024 {
		t.Fatal(s, e)
	}
	if _, e := EvaluateHostStocks("0", "20", "20", "", stocks); e == nil {
		t.Fatal("empty host selector")
	}
}

func TestHostGaugeCanonicalIdentityAndArrivalOrder(t *testing.T) {
	stream := Stream{Resource: map[string]string{"host.id": "one"}, Scope: "hostmetrics", Name: "utilization", Unit: "1"}
	empty := stream
	empty.Attributes = map[string]string{}
	for _, conflict := range []bool{false, true} {
		secondValue := 0.5
		if conflict {
			secondValue = 0.75
		}
		a := GaugePoint{Stream: stream, EndNS: "10", Value: 0.5}
		b := GaugePoint{Stream: empty, EndNS: "10", Value: secondValue}
		forward, e := EvaluateHostGauges("0", "20", "20", "one", []GaugePoint{a, b})
		if e != nil {
			t.Fatal(e)
		}
		reverse, e := EvaluateHostGauges("0", "20", "20", "one", []GaugePoint{b, a})
		if e != nil || !reflect.DeepEqual(forward, reverse) {
			t.Fatal("arrival order changed host result", forward, reverse, e)
		}
		if len(forward.Series) != 1 || len(forward.Series[0].Samples) != 1 {
			t.Fatal("nil/empty attributes split same stream", forward)
		}
		sample := forward.Series[0].Samples[0]
		if conflict && (sample.State != "unknown" || sample.Value != nil) {
			t.Fatal("conflicting exactstream observations gained winner", sample)
		}
		if !conflict && (sample.State != "observed" || sample.Value == nil || *sample.Value != 0.5) {
			t.Fatal("duplicate observation changed", sample)
		}
	}
	zero, e := EvaluateHostGauges("0", "20", "20", "one", []GaugePoint{{Stream: stream, EndNS: "10", Value: math.Copysign(0, -1)}})
	if e != nil || math.Signbit(*zero.Series[0].Samples[0].Value) {
		t.Fatal("host negativezero not canonical", zero, e)
	}
	explicitEmptyService := stream
	explicitEmptyService.Resource = map[string]string{"host.id": "one", "service.name": ""}
	distinct, e := EvaluateHostGauges("0", "20", "20", "one", []GaugePoint{{Stream: stream, EndNS: "10", Value: 0.5}, {Stream: explicitEmptyService, EndNS: "10", Value: 0.5}})
	if e != nil || len(distinct.Series) != 2 {
		t.Fatal("missing and emptyservice merged", distinct, e)
	}
	if stream.Attributes != nil || len(stream.Resource) != 1 {
		t.Fatal("caller identity mutated")
	}
}
