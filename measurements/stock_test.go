package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func stockDefinition(t *testing.T) Definition {
	d := definition(t, "checkout-calls")
	d.Kind = "stock"
	d.Source.Metric = "db.client.connection.count"
	d.Source.Unit = "{connection}"
	return d
}
func stockRow(d Definition, at int64, value string) metricRow {
	r := gaugeRow(d, "ordinary-python", at, value)
	r.StartNS = ns(90000)
	r.Temporality = 2
	r.Monotonic = false
	return r
}
func TestStockPackageSourceAndReplay(t *testing.T) {
	d := stockDefinition(t)
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	plan, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(plan.SQL, "otel_metrics_sum") || !strings.Contains(plan.SQL, "IsMonotonic AS monotonic") || plan.Parameters["p0"] != "100000" || strings.Contains(plan.SQL, "TimeUnix<=") {
		t.Fatal(plan)
	}
	a := stockRow(d, 110000, "3.5")
	b := stockRow(d, 120000, "-1.25")
	b.StartNS = ns(115000)
	result, e := Evaluate(q, []json.RawMessage{raw(a), raw(b)}, time.Unix(200, 0))
	if e != nil || result.Stocks == nil || result.Gauges != nil || result.Counters != nil {
		t.Fatal(result, e)
	}
	samples := result.Stocks.Series[0].Samples
	if len(samples) != 2 || *samples[1].Value != -1.25 || samples[1].StartNS != b.StartNS {
		t.Fatal(samples)
	}
	if e := ValidateResult(result); e != nil {
		t.Fatal(e)
	}
	encoded, _ := json.Marshal(result)
	var restored Result
	if e = json.Unmarshal(encoded, &restored); e != nil {
		t.Fatal(e)
	}
	if e = ValidateResult(&restored); e != nil {
		t.Fatal(e)
	}
	restored.Stocks.Series[0].Samples[1].StartNS = ns(90000)
	if ValidateResult(&restored) == nil {
		t.Fatal("forged epoch replay accepted")
	}
}
func TestStockSourceRefusalsAndFlags(t *testing.T) {
	d := stockDefinition(t)
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	good := stockRow(d, 120000, "1.5")
	for _, edit := range []func(*metricRow){func(r *metricRow) { r.Temporality = 1 }, func(r *metricRow) { r.Monotonic = true }, func(r *metricRow) { r.Value = "NaN" }, func(r *metricRow) { r.Flags = 2 }, func(r *metricRow) { r.StartNS = ns(130000) }, func(r *metricRow) { r.Resource["service.instance.id"] = "" }} {
		r := good
		r.Resource = map[string]string{}
		for k, v := range good.Resource {
			r.Resource[k] = v
		}
		edit(&r)
		if _, e := Evaluate(q, []json.RawMessage{raw(r)}, time.Unix(200, 0)); e == nil {
			t.Fatal("invalid stock accepted", r)
		}
	}
	missing := good
	missing.Flags = 1
	missing.Value = "not-a-value"
	r, e := Evaluate(q, []json.RawMessage{raw(missing)}, time.Unix(200, 0))
	if e != nil || r.State != "unknown" || r.Stocks.Series[0].Samples[0].Value != nil {
		t.Fatal(r, e)
	}
	empty, e := Evaluate(q, nil, time.Unix(200, 0))
	if e != nil || empty.State != "no_data" || len(empty.Stocks.Series) != 0 {
		t.Fatal(empty, e)
	}
}
func TestStockDiscoveryClassificationAndCapabilities(t *testing.T) {
	q := DiscoveryQuestion{Environment: "local", Service: "independent", FromMS: 1000, ToMS: 2000}
	base := `{"kind":"counter","scope":"ordinary","version":"","schemaURL":"","metric":"stock","unit":"1","temporality":2,"monotonic":false,"dimensions":[]}`
	rows := []json.RawMessage{json.RawMessage(base), json.RawMessage(strings.Replace(base, `"temporality":2`, `"temporality":1`, 1))}
	r, e := discoverRows(q, rows)
	if e != nil || len(r.Instruments) != 2 {
		t.Fatal(r, e)
	}
	found := false
	for _, i := range r.Instruments {
		if i.Kind == "stock" {
			found = true
			if !i.Supported || i.Definition == nil || i.Definition.Kind != "stock" || !i.CapabilityPlan.Eligible || !strings.Contains(strings.Join(i.CapabilityPlan.Hints, " "), "decreases") {
				t.Fatal(i)
			}
		} else if i.Supported {
			t.Fatal("delta nonmono admitted")
		}
	}
	if !found {
		t.Fatal("cumulative stock missing")
	}
}
func TestStockAdditionDoesNotChangeLegacyResultWire(t *testing.T) {
	d := definition(t, "checkout-calls")
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	r, e := Evaluate(q, nil, time.Unix(200, 0))
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(r)
	if strings.Contains(string(b), `"stocks"`) {
		t.Fatal("legacy stock wire present")
	}
}
