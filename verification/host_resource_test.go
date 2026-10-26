package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"reflect"
	"strings"
	"testing"
	"time"
)

func hrQuestion(host, kind, presence string) measurements.Question {
	i := hdInstrument("host-observation", "system.cpu.utilization", kind, "1", exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}, "unused")
	d := *i.Definition
	d.Source.Services = []string{}
	d.Source.Namespace = nil
	d.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: host}
	d.Source.SchemaURL = "https://opentelemetry.io/schemas/1.37.0"
	if kind == "stock" {
		d.Source.Metric = "system.memory.usage"
		d.Source.Unit = "By"
		d.StockDomain = "nonnegative-integer"
	}
	resource := map[string]string{"host.id": host, "host.name": "shared-label", "deployment.environment.name": "prod"}
	if presence != "missing" {
		resource["service.name"] = ""
		resource["service.instance.id"] = ""
	}
	if presence == "named" {
		resource["service.name"] = "agent"
		resource["service.instance.id"] = "agent-one"
	}
	d.Source.ExactStream = &metrics.Stream{Resource: resource, ResourceSchemaURL: "raw-resource-schema", Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: d.Source.Metric, Unit: d.Source.Unit, Attributes: map[string]string{"reported": "value"}}
	return measurements.Question{Definition: d, FromMS: 100000, ToMS: 160000}
}
func hrRow(q measurements.Question, value string) map[string]any {
	s := q.Definition.Source.ExactStream
	return map[string]any{"resource": s.Resource, "resource_schema_url": s.ResourceSchemaURL, "scope": s.Scope, "version": s.Version, "schema_url": s.SchemaURL, "name": s.Name, "unit": s.Unit, "attributes": s.Attributes, "start_ns": "90000000000", "end_ns": "120000000000", "value": value, "flags": 0, "temporality": 2, "monotonic": false}
}
func TestHegelHostAuthorityAndRawIdentity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		host := fmt.Sprintf("host-%d", hegel.Draw(tc, hegel.Integers(1, 100000)))
		for _, kind := range []string{"gauge", "stock"} {
			for _, presence := range []string{"missing", "empty", "named"} {
				q := hrQuestion(host, kind, presence)
				row := hrRow(q, "0")
				if kind == "gauge" {
					row["value"] = "0.25"
				}
				before, _ := json.Marshal(q)
				raw, _ := json.Marshal(row)
				r, e := measurements.Evaluate(q, []json.RawMessage{raw}, time.Unix(200, 0), "", nil)
				if e != nil || measurements.ValidateResult(r) != nil {
					tc.Errorf("valid raw host %s/%s %v", kind, presence, e)
					return
				}
				var stream metrics.Stream
				if kind == "gauge" {
					stream = r.Gauges.Series[0].Stream
				} else {
					stream = r.Stocks.Series[0].Stream
				}
				if !reflect.DeepEqual(stream, *q.Definition.Source.ExactStream) {
					tc.Errorf("raw identity fabricated/normalized")
					return
				}
				after, _ := json.Marshal(q)
				if string(before) != string(after) {
					tc.Errorf("input mutation")
					return
				}
				if presence == "missing" {
					if _, ok := stream.Resource["service.name"]; ok {
						tc.Errorf("service fabricated")
						return
					}
				}
				for _, field := range []string{"host", "environment", "resource-schema", "scope", "schema", "attributes", "window"} {
					copyRaw, _ := json.Marshal(row)
					var wrong map[string]any
					_ = json.Unmarshal(copyRaw, &wrong)
					switch field {
					case "host":
						wrong["resource"].(map[string]any)["host.id"] = "other"
					case "environment":
						wrong["resource"].(map[string]any)["deployment.environment.name"] = "other"
					case "resource-schema":
						wrong["resource_schema_url"] = "other"
					case "scope":
						wrong["scope"] = "other"
					case "schema":
						wrong["schema_url"] = "other"
					case "attributes":
						wrong["attributes"].(map[string]any)["extra"] = "other"
					case "window":
						wrong["end_ns"] = "160000000000"
					}
					bad, _ := json.Marshal(wrong)
					if _, e := measurements.Evaluate(q, []json.RawMessage{bad}, time.Unix(200, 0), "", nil); e == nil {
						tc.Errorf("raw host mismatch accepted %s", field)
						return
					}
				}
			}
		}
	}, options(10)...)
}
func TestHostAuthorityDoesNotRelaxServiceOrCounter(t *testing.T) {
	for _, kind := range []string{"counter", "histogram"} {
		q := hrQuestion("one", kind, "missing")
		if q.Validate() == nil {
			t.Fatal("host authority expanded", kind)
		}
	}
	for _, presence := range []string{"missing", "empty"} {
		q := hrQuestion("one", "gauge", presence)
		s := *q.Definition.Source.ExactStream
		points := []metrics.GaugePoint{{Stream: s, EndNS: "120000000000", Value: 0.25}}
		if _, e := metrics.EvaluateGauges("100000000000", "160000000000", "60000000000", points); e == nil {
			t.Fatal("legacy service gauge relaxed")
		}
		if _, e := metrics.EvaluateHostGauges("100000000000", "160000000000", "60000000000", "other", points); e == nil {
			t.Fatal("host selection ignored")
		}
	}
	legacy := hdInstrument("legacy", "calls", "counter", "1", exploration.InventoryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000}, "api").Definition
	b, _ := json.Marshal(legacy)
	if strings.Contains(string(b), `"resource":`) {
		t.Fatal("legacy wire changed")
	}
	p, e := measurements.Compile(measurements.Question{Definition: *legacy, FromMS: 100000, ToMS: 160000})
	if e != nil || !strings.Contains(p.SQL, "mapUpdate(ResourceAttributes") || !strings.Contains(p.SQL, "ServiceName IN") {
		t.Fatal("legacy source policy", e)
	}
}

func TestHostRawUnknownStaleAndConflictingLatest(t *testing.T) {
	s := *hrQuestion("one", "gauge", "missing").Definition.Source.ExactStream
	for _, c := range []struct {
		name, state string
		points      []metrics.GaugePoint
	}{{"zero", "observed", []metrics.GaugePoint{{Stream: s, EndNS: "90", Value: 0}}}, {"stale", "stale", []metrics.GaugePoint{{Stream: s, EndNS: "20", Value: .5}}}, {"missing", "unknown", []metrics.GaugePoint{{Stream: s, EndNS: "90", Value: .5, NoRecordedValue: true}}}, {"conflict", "unknown", []metrics.GaugePoint{{Stream: s, EndNS: "90", Value: .5}, {Stream: s, EndNS: "90", Value: .25}}}} {
		g, e := metrics.EvaluateHostGauges("0", "100", "20", "one", c.points)
		if e != nil || len(g.Series) != 1 {
			t.Fatal(c.name, e)
		}
		v := g.Series[0].Samples[len(g.Series[0].Samples)-1]
		if v.State != c.state || (c.state == "unknown" && v.Value != nil) {
			t.Fatal(c.name, v)
		}
		points := []metrics.StockPoint{}
		for _, p := range c.points {
			points = append(points, metrics.StockPoint{Stream: p.Stream, Temporality: "cumulative", StartNS: "0", EndNS: p.EndNS, Value: p.Value, NoRecordedValue: p.NoRecordedValue})
		}
		r, e := metrics.EvaluateHostStocks("0", "100", "20", "one", points)
		if e != nil || len(r.Series) != 1 {
			t.Fatal(c.name, e)
		}
		stock := r.Series[0].Samples[len(r.Series[0].Samples)-1]
		if stock.State != c.state || (c.state == "unknown" && stock.Value != nil) {
			t.Fatal(c.name, stock)
		}
	}
}
