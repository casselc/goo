package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func poolInstrument(t *testing.T) Instrument {
	d := stockDefinition(t)
	d.Source.SchemaURL = PoolConnectionsSchema
	i := Instrument{Kind: "stock", Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Metric: d.Source.Metric, Unit: d.Source.Unit, Temporality: 2, Supported: true, Definition: &d}
	r := stockRow(d, 120000, "3")
	p := PlanInstrument(i, []DiscoveryObservation{{Resource: r.Resource, Attributes: map[string]string{"db.client.connection.pool.name": "primary", "db.client.connection.state": "used"}}})
	i.CapabilityPlan = &p
	return i
}
func TestStockDomainValidationValuesAndReplay(t *testing.T) {
	d := stockDefinition(t)
	d.StockDomain = "nonnegative-integer"
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	for _, v := range []string{"0", "3", "9007199254740991"} {
		r, e := Evaluate(q, []json.RawMessage{raw(stockRow(d, 120000, v))}, time.Unix(200, 0))
		if e != nil {
			t.Fatal(v, e)
		}
		if e = ValidateResult(r); e != nil {
			t.Fatal(e)
		}
	}
	for _, v := range []string{"-1", "1.5", "9007199254740992", "9007199254740991.1", "1.00000000000000001"} {
		if _, e := Evaluate(q, []json.RawMessage{raw(stockRow(d, 120000, v))}, time.Unix(200, 0)); e == nil {
			t.Fatal("invalid domain accepted", v)
		}
	}
	missing := stockRow(d, 120000, "invalid")
	missing.Flags = 1
	if _, e := Evaluate(q, []json.RawMessage{raw(missing)}, time.Unix(200, 0)); e != nil {
		t.Fatal("missing flag domain", e)
	}
	generic := d
	generic.StockDomain = ""
	if _, e := Evaluate(Question{Definition: generic, FromMS: q.FromMS, ToMS: q.ToMS}, []json.RawMessage{raw(stockRow(generic, 120000, "-1.5"))}, time.Unix(200, 0)); e != nil {
		t.Fatal("generic stock changed", e)
	}
	for _, kind := range []string{"gauge", "counter", "histogram", "lifecycle"} {
		bad := d
		bad.Kind = kind
		if bad.Validate() == nil {
			t.Fatal("domain accepted on", kind)
		}
	}
	bad := d
	bad.StockDomain = "future"
	if bad.Validate() == nil {
		t.Fatal("unknown domain")
	}
	b, _ := json.Marshal(generic)
	if strings.Contains(string(b), "stockDomain") {
		t.Fatal("old wire changed")
	}
}
func TestPoolCandidateSameStreamAndBoundedPreview(t *testing.T) {
	i := poolInstrument(t)
	if !PoolConnectionsCandidate(i) || i.CapabilityPlan.Profile != "database-pool-connections" {
		t.Fatal(i)
	}
	base := i.CapabilityPlan.Coverage.Streams[0]
	for _, state := range []string{"used", "idle", "waiting-custom"} {
		copy := i
		p := *i.CapabilityPlan
		stream := base
		stream.Stream.Attributes = map[string]string{"db.client.connection.pool.name": "primary", "db.client.connection.state": state}
		p.Coverage.Streams = []ObservedStream{stream}
		copy.CapabilityPlan = &p
		if !PoolConnectionsCandidate(copy) {
			t.Fatal(state)
		}
	}
	for _, key := range []string{"db.client.connection.pool.name", "db.client.connection.state"} {
		copy := i
		p := *i.CapabilityPlan
		stream := base
		stream.Stream.Attributes = map[string]string{"db.client.connection.pool.name": "primary", "db.client.connection.state": "used"}
		stream.Stream.Attributes[key] = ""
		p.Coverage.Streams = []ObservedStream{stream}
		copy.CapabilityPlan = &p
		if PoolConnectionsCandidate(copy) {
			t.Fatal("empty admitted", key)
		}
	}
	p := *i.CapabilityPlan
	a, b := base, base
	a.Stream.Attributes = map[string]string{"db.client.connection.pool.name": "primary"}
	b.Stream.Attributes = map[string]string{"db.client.connection.state": "used"}
	p.Coverage.Streams = []ObservedStream{a, b}
	copy := i
	copy.CapabilityPlan = &p
	if PoolConnectionsCandidate(copy) {
		t.Fatal("union witness")
	}
	p.Coverage.Streams = make([]ObservedStream, 26)
	for n := range p.Coverage.Streams {
		p.Coverage.Streams[n] = a
	}
	p.Coverage.Streams[25] = base
	if PoolConnectionsCandidate(copy) {
		t.Fatal("outside preview admitted")
	}
	copy = i
	copy.SchemaURL = "https://opentelemetry.io/schemas/1.30.0"
	if PoolConnectionsCandidate(copy) {
		t.Fatal("wrong schema")
	}
}
