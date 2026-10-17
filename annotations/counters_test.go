package annotations

import (
	"math"
	"strconv"
	"testing"
)

func TestNativeRateSourceContract(t *testing.T) {
	p, e := LoadOperationPackage("../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	contract := p.Definition.Rates[0]
	row := CounterRow{Resource: map[string]string{"service.name": p.Source.Service, "service.instance.id": p.Source.Instance, "deployment.environment.name": p.Source.Environment}, Scope: p.Manifest.Scope, Version: p.Manifest.ScopeVersion, SchemaURL: p.Manifest.SchemaURL, Name: contract.Metric, Unit: contract.Unit, Attributes: map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": "ok"}, StartNS: "1000000", EndNS: "100000000", Value: "2", Temporality: 2, Monotonic: true}
	later := row
	later.EndNS = "200000000"
	later.Value = "5"
	result, e := p.EvaluateRateRows(contract, 100, 200, []CounterRow{row, later})
	if e != nil || result.Series[0].State != "ok" || result.Series[0].Intervals[0].Delta != "3" {
		t.Fatal(result, e)
	}
	for _, mutate := range []func(*CounterRow){func(r *CounterRow) { r.Value = "9007199254740992" }, func(r *CounterRow) { r.Value = "1.5" }, func(r *CounterRow) { r.Value = strconv.FormatFloat(math.Inf(1), 'g', -1, 64) }, func(r *CounterRow) { r.SchemaURL = "" }, func(r *CounterRow) { r.Temporality = 1 }, func(r *CounterRow) { r.Flags = 2 }, func(r *CounterRow) { r.Monotonic = false }, func(r *CounterRow) { r.StartNS = "0" }} {
		bad := row
		mutate(&bad)
		if _, e = p.EvaluateRateRows(contract, 100, 200, []CounterRow{bad}); e == nil {
			t.Fatal("accepted invalid source counter")
		}
	}
	missing := later
	missing.Flags = 1
	missing.Value = "ignored"
	r, e := p.EvaluateRateRows(contract, 100, 200, []CounterRow{row, missing})
	if e != nil || r.Series[0].Intervals[0].State != "unknown" {
		t.Fatal("flagged interval became known", e)
	}
}
