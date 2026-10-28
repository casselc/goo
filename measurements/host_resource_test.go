package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
)

func hostResourceQuestion(t *testing.T, kind string) Question {
	t.Helper()
	d := definition(t, "checkout-calls")
	d.Kind = kind
	d.Source.Services = []string{}
	d.Source.Namespace = nil
	d.Source.Resource = &ResourceBinding{Kind: "host", ID: "host-one"}
	d.Source.Metric = "system.cpu.utilization"
	d.Source.Unit = "1"
	if kind == "stock" {
		d.Source.Metric = "system.memory.usage"
		d.Source.Unit = "By"
		d.StockDomain = "nonnegative-integer"
	}
	d.Source.ExactStream = &metrics.Stream{Resource: map[string]string{"host.id": "host-one", "host.name": "same-label", "deployment.environment.name": d.Environment}, ResourceSchemaURL: "resource-schema", Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: d.Source.Metric, Unit: d.Source.Unit, Attributes: map[string]string{"cpu.mode": "user", "cpu.logical_number": "0"}}
	return Question{Definition: d, FromMS: 100000, ToMS: 160000}
}
func hostResourceRow(q Question, value string) metricRow {
	s := q.Definition.Source.ExactStream
	return metricRow{Resource: s.Resource, ResourceSchemaURL: s.ResourceSchemaURL, Scope: s.Scope, Version: s.Version, SchemaURL: s.SchemaURL, Name: s.Name, Unit: s.Unit, Attributes: s.Attributes, StartNS: ns(90000), EndNS: ns(120000), Temporality: 2, Value: value}
}
func TestHostResourceCompileEvaluateReplayWithoutService(t *testing.T) {
	for _, kind := range []string{"gauge", "stock"} {
		q := hostResourceQuestion(t, kind)
		plan, e := Compile(q)
		if e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(plan.SQL, "ResourceAttributes AS resource") || strings.Contains(plan.SQL, "mapUpdate") || strings.Contains(plan.SQL, "ServiceName IN") || !strings.Contains(plan.SQL, "mapContains(ResourceAttributes,'host.id')") {
			t.Fatal(plan.SQL)
		}
		value := "0.25"
		if kind == "stock" {
			value = "4096"
		}
		r, e := Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, value))}, time.Unix(200, 0))
		if e != nil {
			t.Fatal(e)
		}
		if e = ValidateResult(r); e != nil {
			t.Fatal(e)
		}
		var stream metrics.Stream
		if kind == "gauge" {
			stream = r.Gauges.Series[0].Stream
		} else {
			stream = r.Stocks.Series[0].Stream
		}
		if _, present := stream.Resource["service.name"]; present {
			t.Fatal("synthetic service added")
		}
		if _, present := stream.Resource["service.instance.id"]; present {
			t.Fatal("synthetic instance added")
		}
		q.Definition.Source.ExactStream.Resource["service.name"] = ""
		r, e = Evaluate(q, []json.RawMessage{raw(hostResourceRow(q, value))}, time.Unix(200, 0))
		if e != nil {
			t.Fatal(e)
		}
		if e = ValidateResult(r); e != nil {
			t.Fatal(e)
		}
	}
}
func TestHostResourceRefusesIdentityBroadeningAndOtherKinds(t *testing.T) {
	for _, edit := range []func(*Question){func(q *Question) { q.Definition.Source.Services = []string{"fake"} }, func(q *Question) { q.Definition.Source.ExactStream = nil }, func(q *Question) { q.Definition.Source.Resource.Kind = "future" }, func(q *Question) { q.Definition.Source.Resource.ID = "other" }, func(q *Question) { q.Definition.Source.Resource.ID = "" }, func(q *Question) { q.Definition.Kind = "counter" }, func(q *Question) { q.Definition.Kind = "histogram" }} {
		q := hostResourceQuestion(t, "gauge")
		edit(&q)
		if q.Validate() == nil {
			t.Fatal("host authority relaxed", q)
		}
	}
	for _, key := range []string{"host.id", "deployment.environment.name"} {
		q := hostResourceQuestion(t, "gauge")
		row := hostResourceRow(q, "0.5")
		resource := map[string]string{}
		for k, v := range row.Resource {
			resource[k] = v
		}
		resource[key] = "other"
		row.Resource = resource
		if _, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0)); e == nil {
			t.Fatal("source identity mismatch", key)
		}
	}
	q := hostResourceQuestion(t, "gauge")
	row := hostResourceRow(q, "0.5")
	row.ResourceSchemaURL = "other"
	if _, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0)); e == nil {
		t.Fatal("resource schema broadened")
	}
}
func TestHostResourceOmissionPreservesLegacyDefinition(t *testing.T) {
	d := definition(t, "checkout-calls")
	b, _ := json.Marshal(d)
	if strings.Contains(string(b), `"resource":`) {
		t.Fatal("optional resource binding leaked into old bytes")
	}
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.SQL, "mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName)))") || !strings.Contains(p.SQL, "ServiceName IN") {
		t.Fatal("legacy source changed", p, e)
	}
}
