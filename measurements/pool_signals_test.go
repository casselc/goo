package measurements

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/metrics"
)

func poolSignalsFixture(t *testing.T) PoolSignalsRequest {
	t.Helper()
	i := poolInstrument(t)
	d := *i.Definition
	d.StockDomain = "nonnegative-integer"
	stream := i.CapabilityPlan.Coverage.Streams[0].Stream
	parent := Question{Definition: d, FromMS: 100000, ToMS: 102001, FromNS: "100000000123", ToNS: "102000000123"}
	instruments := []Instrument{}
	for _, kind := range []string{"capacity", "wait"} {
		def := d
		def.StockDomain = ""
		def.Name = "candidate-" + kind
		def.Title = kind
		if kind == "capacity" {
			def.Source.Metric = "db.client.connection.max"
		} else {
			def.Kind = "histogram"
			def.Source.Metric = "db.client.connection.wait_time"
			def.Source.Unit = "s"
		}
		candidateStream := stream
		candidateStream.Name = def.Source.Metric
		candidateStream.Unit = def.Source.Unit
		candidateStream.Attributes = map[string]string{"db.client.connection.pool.name": "primary"}
		candidate := Instrument{Kind: def.Kind, Metric: def.Source.Metric, Unit: def.Source.Unit, Scope: def.Source.Scope, Version: def.Source.Version, SchemaURL: def.Source.SchemaURL, Temporality: 2, Supported: true, Definition: &def, CapabilityPlan: &InstrumentCapabilityPlan{Coverage: InstrumentCoverage{Streams: []ObservedStream{{Stream: candidateStream}}}}}
		instruments = append(instruments, candidate)
	}
	return PoolSignalsRequest{Parent: parent, Stream: stream, Instruments: instruments}
}
func TestPoolSignalsExactPinnedIndependentQuestions(t *testing.T) {
	req := poolSignalsFixture(t)
	before, _ := json.Marshal(req)
	plan, e := PlanPoolSignals(req)
	if e != nil || len(plan.Capacity) != 1 || len(plan.Wait) != 1 {
		t.Fatal(plan, e)
	}
	if !reflect.DeepEqual(plan.Parent, req.Parent) || !sameExactStream(plan.Stream, req.Stream) {
		t.Fatal("parent changed")
	}
	for _, option := range append(plan.Capacity, plan.Wait...) {
		q := option.Question
		if q.FromMS != req.Parent.FromMS || q.ToMS != req.Parent.ToMS || q.FromNS != req.Parent.FromNS || q.ToNS != req.Parent.ToNS || q.Definition.Source.ExactStream == nil || option.Key != Hash(q) {
			t.Fatal(option)
		}
		if _, e := Compile(q); e != nil {
			t.Fatal(e)
		}
		if _, ok := q.Definition.Source.ExactStream.Attributes["db.client.connection.state"]; ok {
			t.Fatal("state leaked")
		}
	}
	if plan.Capacity[0].Question.Definition.StockDomain != "nonnegative-integer" {
		t.Fatal("capacity domain missing")
	}
	after, _ := json.Marshal(req)
	if string(before) != string(after) {
		t.Fatal("input mutated")
	}
	plan.Parent.Definition.Source.Services[0] = "changed"
	plan.Stream.Resource["service.instance.id"] = "changed"
	after, _ = json.Marshal(req)
	if string(before) != string(after) {
		t.Fatal("output aliases input")
	}
}
func TestPoolSignalsIdentityNoBroadening(t *testing.T) {
	for _, edit := range []func(*metrics.Stream){func(s *metrics.Stream) { s.ResourceSchemaURL = "other" }, func(s *metrics.Stream) { s.Resource = map[string]string{} }, func(s *metrics.Stream) { s.Scope = "other" }, func(s *metrics.Stream) { s.Version = "other" }, func(s *metrics.Stream) { s.SchemaURL = "other" }, func(s *metrics.Stream) {
		s.Attributes = map[string]string{"db.client.connection.pool.name": "primary", "db.client.connection.state": "used"}
	}, func(s *metrics.Stream) {
		s.Attributes = map[string]string{"db.client.connection.pool.name": "primary", "extra": "x"}
	}, func(s *metrics.Stream) { s.Attributes = map[string]string{} }, func(s *metrics.Stream) { s.Attributes = map[string]string{"db.client.connection.pool.name": "other"} }} {
		req := poolSignalsFixture(t)
		for n := range req.Instruments {
			edit(&req.Instruments[n].CapabilityPlan.Coverage.Streams[0].Stream)
		}
		plan, e := PlanPoolSignals(req)
		if e != nil || len(plan.Capacity)+len(plan.Wait) != 0 {
			t.Fatal("broadened", plan, e)
		}
	}
	req := poolSignalsFixture(t)
	req.Parent.Definition.StockDomain = ""
	if _, e := PlanPoolSignals(req); e == nil {
		t.Fatal("uncertified parent")
	}
	req = poolSignalsFixture(t)
	pinned := req.Stream
	req.Parent.Definition.Source.ExactStream = &pinned
	different := req.Stream
	different.Attributes = map[string]string{"db.client.connection.pool.name": "different", "db.client.connection.state": "used"}
	req.Stream = different
	if _, e := PlanPoolSignals(req); e == nil {
		t.Fatal("pinned parent replaced")
	}
}
func TestPoolSignalsPreviewBudgetAndDefinitionNameDedup(t *testing.T) {
	req := poolSignalsFixture(t)
	copy := req.Instruments[0]
	d := *copy.Definition
	d.Name = "duplicate-name"
	copy.Definition = &d
	req.Instruments = append(req.Instruments, copy)
	plan, e := PlanPoolSignals(req)
	if e != nil || len(plan.Capacity) != 1 || len(plan.Wait) != 1 {
		t.Fatal("name became identity", plan, e)
	}
	reverse := req
	reverse.Instruments = append([]Instrument{}, req.Instruments...)
	for a, b := 0, len(reverse.Instruments)-1; a < b; a, b = a+1, b-1 {
		reverse.Instruments[a], reverse.Instruments[b] = reverse.Instruments[b], reverse.Instruments[a]
	}
	again, e := PlanPoolSignals(reverse)
	if e != nil || !reflect.DeepEqual(plan, again) {
		t.Fatal("order changed plan", e)
	}
	req = poolSignalsFixture(t)
	for n := range req.Instruments {
		coverage := &req.Instruments[n].CapabilityPlan.Coverage
		match := coverage.Streams[0]
		bad := match
		bad.Stream.Attributes = map[string]string{"db.client.connection.pool.name": "other"}
		coverage.Streams = make([]ObservedStream, 26)
		for k := range coverage.Streams {
			coverage.Streams[k] = bad
		}
		coverage.Streams[25] = match
	}
	plan, e = PlanPoolSignals(req)
	if e != nil || len(plan.Capacity)+len(plan.Wait) != 0 || plan.StreamsOmitted != 2 {
		t.Fatal("preview became absence/fullscan", plan, e)
	}
	req = poolSignalsFixture(t)
	req.Instruments = make([]Instrument, 17)
	if _, e := PlanPoolSignals(req); e == nil {
		t.Fatal("candidate budget")
	}
}

func TestPoolSignalsPreviewCountsRefuseOverflow(t *testing.T) {
	for _, count := range []int{-1, RowLimit + 1, int(^uint(0) >> 1)} {
		req := poolSignalsFixture(t)
		req.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = count
		if _, err := PlanPoolSignals(req); err == nil {
			t.Fatalf("invalid preview count accepted: %d", count)
		}
	}
	req := poolSignalsFixture(t)
	req.Instruments[0].CapabilityPlan.Coverage.OmittedStreams = RowLimit
	req.Instruments[0].CapabilityPlan.Coverage.Streams = make([]ObservedStream, DiscoveryStreamLimit+1)
	if _, err := PlanPoolSignals(req); err == nil {
		t.Fatal("combined preview count exceeded bound")
	}
	req = poolSignalsFixture(t)
	req.Instruments[0].CapabilityPlan.Coverage.Streams = make([]ObservedStream, RowLimit+1)
	if _, err := PlanPoolSignals(req); err == nil {
		t.Fatal("oversized preview array accepted")
	}
	req = poolSignalsFixture(t)
	for n := range req.Instruments {
		req.Instruments[n].CapabilityPlan.Coverage.OmittedStreams = RowLimit
	}
	result, err := PlanPoolSignals(req)
	if err != nil || result.StreamsOmitted != 2*RowLimit {
		t.Fatal(result, err)
	}
}
