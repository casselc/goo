package annotations

import "testing"

func TestHistogramSourceContract(t *testing.T) {
	p, e := LoadOperationPackage("../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	c := p.Definition.Histograms[0]
	base := CounterRow{Resource: map[string]string{"service.name": p.Source.Service, "service.instance.id": p.Source.Instance, "deployment.environment.name": p.Source.Environment}, Scope: p.Manifest.Scope, Version: p.Manifest.ScopeVersion, SchemaURL: p.Manifest.SchemaURL, Name: c.Metric, Unit: c.Unit, Attributes: map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": "ok"}, StartNS: "1000000", EndNS: "100000000", Temporality: 2}
	buckets := make([]string, len(c.Bounds)+1)
	for i := range buckets {
		buckets[i] = "0"
	}
	buckets[1] = "9007199254740993"
	row := HistogramRow{CounterRow: base, Bounds: c.Bounds, BucketCounts: buckets, Count: "9007199254740993"}
	next := row
	next.EndNS = "200000000"
	next.Count = "9007199254740994"
	next.BucketCounts = append([]string(nil), buckets...)
	next.BucketCounts[1] = next.Count
	result, e := p.EvaluateHistogramRows(c, 100, 200, []HistogramRow{row, next})
	if e != nil || result.Series[0].Intervals[0].Distribution.Count != "1" {
		t.Fatal("histogram UInt64 source precision lost", e)
	}
	bad := row
	bad.Unit = "ms"
	if _, e = p.EvaluateHistogramRows(c, 100, 200, []HistogramRow{bad}); e == nil {
		t.Fatal("accepted unit mismatch")
	}
	bad = row
	bad.Bounds = append([]float64{}, c.Bounds...)
	bad.Bounds[1] = .002
	if _, e = p.EvaluateHistogramRows(c, 100, 200, []HistogramRow{bad}); e == nil {
		t.Fatal("accepted bucket mismatch")
	}
	bad = row
	bad.Count = "1"
	if _, e = p.EvaluateHistogramRows(c, 100, 200, []HistogramRow{bad}); e == nil {
		t.Fatal("accepted invalid population")
	}
}
