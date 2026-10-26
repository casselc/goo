package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"testing"
)

func psFixture(label string) measurements.PoolSignalsRequest {
	i := hdInstrument("count", "db.client.connection.count", "stock", "{connection}", exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 102001}, "api")
	i.Definition.StockDomain = "nonnegative-integer"
	i.Definition.Source.SchemaURL = measurements.PoolConnectionsSchema
	s := metrics.Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": label, "deployment.environment.name": "prod"}, ResourceSchemaURL: "resource-schema", Scope: i.Scope, Version: i.Version, SchemaURL: measurements.PoolConnectionsSchema, Name: i.Metric, Unit: i.Unit, Attributes: map[string]string{"db.client.connection.pool.name": label, "db.client.connection.state": "used", "tenant": "one"}}
	q := measurements.PoolSignalsRequest{Parent: measurements.Question{Definition: *i.Definition, FromMS: 100000, ToMS: 102001, FromNS: "100000000123", ToNS: "102000000123"}, Stream: s}
	for _, kind := range []string{"stock", "histogram"} {
		d := *i.Definition
		d.StockDomain = ""
		d.Kind = kind
		d.Name = "candidate-" + kind
		d.Source.Metric = "db.client.connection.max"
		if kind == "histogram" {
			d.Source.Metric = "db.client.connection.wait_time"
			d.Source.Unit = "s"
		}
		cs := s
		cs.Name = d.Source.Metric
		cs.Unit = d.Source.Unit
		cs.Attributes = map[string]string{"db.client.connection.pool.name": label, "tenant": "one"}
		q.Instruments = append(q.Instruments, measurements.Instrument{Supported: true, Kind: kind, Metric: d.Source.Metric, Unit: d.Source.Unit, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Temporality: 2, Definition: &d, CapabilityPlan: &measurements.InstrumentCapabilityPlan{Coverage: measurements.InstrumentCoverage{Streams: []measurements.ObservedStream{{Stream: cs}}}}})
	}
	return q
}
func psCopy(q measurements.PoolSignalsRequest) measurements.PoolSignalsRequest {
	b, _ := json.Marshal(q)
	var c measurements.PoolSignalsRequest
	_ = json.Unmarshal(b, &c)
	return c
}
func TestHegelPoolSignalsIdentityAndScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		q := psFixture(fmt.Sprintf("pool-%d", hegel.Draw(tc, hegel.Integers(1, 100000))))
		before, _ := json.Marshal(q)
		p, e := measurements.PlanPoolSignals(q)
		if e != nil || len(p.Capacity) != 1 || len(p.Wait) != 1 {
			tc.Errorf("valid pair %v", e)
			return
		}
		for _, o := range append(p.Capacity, p.Wait...) {
			v := o.Question
			if v.FromMS != q.Parent.FromMS || v.ToMS != q.Parent.ToMS || v.FromNS != q.Parent.FromNS || v.ToNS != q.Parent.ToNS || o.Key != measurements.Hash(v) || v.Definition.Source.ExactStream == nil {
				tc.Errorf("scope/hash")
				return
			}
			if _, ok := v.Definition.Source.ExactStream.Attributes["db.client.connection.state"]; ok {
				tc.Errorf("state leaked")
				return
			}
		}
		if p.Capacity[0].Question.Definition.StockDomain != "nonnegative-integer" || p.Wait[0].Question.Definition.StockDomain != "" {
			tc.Errorf("domain")
			return
		}
		edits := []func(*metrics.Stream){func(s *metrics.Stream) { s.Resource["service.instance.id"] += "x" }, func(s *metrics.Stream) { s.Resource["extra"] = "x" }, func(s *metrics.Stream) { s.ResourceSchemaURL += "x" }, func(s *metrics.Stream) { s.Scope += "x" }, func(s *metrics.Stream) { s.Version += "x" }, func(s *metrics.Stream) { s.SchemaURL += "x" }, func(s *metrics.Stream) { s.Attributes["tenant"] = "other" }, func(s *metrics.Stream) { delete(s.Attributes, "tenant") }, func(s *metrics.Stream) { s.Attributes["db.client.connection.state"] = "used" }, func(s *metrics.Stream) { s.Attributes["db.client.connection.pool.name"] += "x" }}
		for n, edit := range edits {
			bad := psCopy(q)
			for k := range bad.Instruments {
				edit(&bad.Instruments[k].CapabilityPlan.Coverage.Streams[0].Stream)
			}
			r, e := measurements.PlanPoolSignals(bad)
			if e != nil || len(r.Capacity)+len(r.Wait) != 0 {
				tc.Errorf("identity broadening %d %v", n, e)
				return
			}
		}
		p.Stream.Resource["service.instance.id"] = "changed"
		p.Capacity[0].Question.Definition.Source.ExactStream.Attributes["tenant"] = "changed"
		after, _ := json.Marshal(q)
		if string(before) != string(after) {
			tc.Errorf("input alias/mutation")
		}
	}, options(12)...)
}
func TestPoolSignalsPreviewAndPin(t *testing.T) {
	for _, at := range []int{24, 25} {
		q := psFixture("primary")
		for n := range q.Instruments {
			match := q.Instruments[n].CapabilityPlan.Coverage.Streams[0]
			bad := psCopy(q).Instruments[n].CapabilityPlan.Coverage.Streams[0]
			bad.Stream.Attributes["tenant"] = "other"
			c := &q.Instruments[n].CapabilityPlan.Coverage
			c.Streams = make([]measurements.ObservedStream, 26)
			for k := range c.Streams {
				c.Streams[k] = bad
			}
			c.Streams[at] = match
		}
		p, e := measurements.PlanPoolSignals(q)
		want := 0
		if at == 24 {
			want = 2
		}
		if e != nil || len(p.Capacity)+len(p.Wait) != want || p.StreamsOmitted != 2 {
			t.Fatal("preview", at, p, e)
		}
	}
	q := psFixture("primary")
	pin := psCopy(q).Stream
	q.Parent.Definition.Source.ExactStream = &pin
	q.Stream.Attributes["db.client.connection.state"] = "idle"
	if _, e := measurements.PlanPoolSignals(q); e == nil {
		t.Fatal("pin replaced")
	}
	for _, n := range []int{-1, 10001, int(^uint(0) >> 1)} {
		q := psFixture("primary")
		q.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = n
		if _, e := measurements.PlanPoolSignals(q); e == nil {
			t.Fatal("omission", n)
		}
	}
	q = psFixture("primary")
	dup := psCopy(q).Instruments[0]
	dup.Definition.Name = "renamed"
	dup.Definition.Title = "renamed"
	q.Instruments = append(q.Instruments, dup)
	p, e := measurements.PlanPoolSignals(q)
	if e != nil || len(p.Capacity) != 1 {
		t.Fatal("name identity", e)
	}
	q.Instruments[0], q.Instruments[2] = q.Instruments[2], q.Instruments[0]
	r, e := measurements.PlanPoolSignals(q)
	a, _ := json.Marshal(p)
	b, _ := json.Marshal(r)
	if e != nil || string(a) != string(b) {
		t.Fatal("order nondeterminism", e)
	}
}
