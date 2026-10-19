package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExactNativeBoundsPlanAndReplay(t *testing.T) {
	d := definition(t, "checkout-calls")
	q := Question{Definition: d, FromMS: 100000, ToMS: 102001, FromNS: "100000000123", ToNS: "102000000123"}
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(p.SQL, "fromUnixTimestamp64Nano") || p.Parameters["p1"] != q.ToNS {
		t.Fatal(p)
	}
	row := metricRow{Resource: map[string]string{"service.name": d.Source.Services[0], "service.instance.id": "sdk-one", "deployment.environment.name": d.Environment}, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: d.Source.Metric, Unit: d.Source.Unit, Attributes: d.Source.Attributes, StartNS: q.FromNS, EndNS: q.ToNS, Value: "6", Temporality: 1, Monotonic: true}
	r, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	if r.Counters.FromNS != q.FromNS || r.Counters.ToNS != q.ToNS || r.Counters.Series[0].State != "ok" {
		t.Fatal(r.Counters)
	}
	b, _ := json.Marshal(r)
	var restored Result
	if e = json.Unmarshal(b, &restored); e != nil {
		t.Fatal(e)
	}
	if e = ValidateResult(&restored); e != nil {
		t.Fatal(e)
	}
	for _, bounds := range [][2]string{{"0100000000123", q.ToNS}, {q.FromNS, ""}, {q.FromNS, q.FromNS}, {"-1", q.ToNS}, {"100001000123", q.ToNS}, {q.FromNS, "4102444800000000000"}} {
		bad := q
		bad.FromNS, bad.ToNS = bounds[0], bounds[1]
		if bad.Validate() == nil {
			t.Fatalf("accepted %v", bounds)
		}
	}
	legacy := q
	legacy.FromNS = ""
	legacy.ToNS = ""
	lp, e := Compile(legacy)
	if e != nil || !strings.Contains(lp.SQL, "fromUnixTimestamp64Milli") {
		t.Fatal(lp, e)
	}
	b, _ = json.Marshal(legacy)
	if strings.Contains(string(b), "fromNS") {
		t.Fatal("legacy hash shape changed")
	}
}
