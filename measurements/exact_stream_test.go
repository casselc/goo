package measurements

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
)

func TestExactStreamBindingPlanReplayAndDrift(t *testing.T) {
	d := definition(t, "checkout-calls")
	row := gaugeRow(d, "one", 102000, "10")
	row.StartNS = ns(100000)
	row.Temporality = 1
	row.Monotonic = true
	row.ResourceSchemaURL = "resource/v1"
	row.Resource["service.namespace"] = ""
	row.Attributes = map[string]string{"route": "", "method": "GET' OR 1=1"}
	d.Source.Attributes = map[string]string{"route": ""}
	stream := metrics.Stream{Resource: row.Resource, ResourceSchemaURL: row.ResourceSchemaURL, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}
	d.Source.ExactStream = &stream
	q := Question{Definition: d, FromMS: 100000, ToMS: 102000}
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(p.SQL, "JSONExtractKeysAndValues") || strings.Contains(p.SQL, "OR 1=1") || len(p.Parameters) > 128 {
		t.Fatal("selector not bounded/parameterized", p)
	}
	for _, key := range []string{"service.namespace", "method"} {
		found := false
		for _, v := range p.Parameters {
			found = found || strings.Contains(v, key)
		}
		if !found {
			t.Fatal("lost identity", key)
		}
	}
	encoded, e := yaml.Marshal(d)
	if e != nil {
		t.Fatal(e)
	}
	restored, e := Read(string(encoded))
	if e != nil || !reflect.DeepEqual(d, restored) {
		t.Fatal("YAML identity changed", e, string(encoded))
	}
	result, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(103, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	if e = ValidateResult(result); e != nil {
		t.Fatal("selected evidence cannot replay", e)
	}
	for _, mutation := range []func(*metricRow){
		func(r *metricRow) { r.Resource["service.instance.id"] = "two" },
		func(r *metricRow) { delete(r.Resource, "service.namespace") },
		func(r *metricRow) { r.Resource["new.resource"] = "" },
		func(r *metricRow) { delete(r.Attributes, "route") },
		func(r *metricRow) { r.Attributes["new.dimension"] = "" },
		func(r *metricRow) { r.ResourceSchemaURL = "resource/v2" },
		func(r *metricRow) { r.SchemaURL = "scope/v2" },
	} {
		var changed metricRow
		json.Unmarshal(raw(row), &changed)
		mutation(&changed)
		if _, e = Evaluate(q, []json.RawMessage{raw(changed)}, time.Unix(103, 0), "", nil); e == nil {
			t.Fatal("off-stream evidence accepted", changed)
		}
	}
	// Empty maps and nil maps both denote no attributes, and order is immaterial.
	a := stream
	b := stream
	b.Attributes = map[string]string{"method": "GET' OR 1=1", "route": ""}
	if !sameExactStream(a, b) {
		t.Fatal("map insertion order became identity")
	}
	a.Attributes = nil
	b.Attributes = map[string]string{}
	if !sameExactStream(a, b) {
		t.Fatal("empty map identity drift")
	}
	// Exact bindings are reusable typed definitions, not permissive extra hints.
	for _, mutation := range []func(*Definition){
		func(d *Definition) { d.Source.ExactStream.Unit = "different" },
		func(d *Definition) { d.Source.ExactStream.Resource["deployment.environment.name"] = "other" },
		func(d *Definition) { d.Source.ExactStream.Attributes["route"] = "other" },
		func(d *Definition) { d.Source.ExactStream.Resource["service.instance.id"] = "" },
	} {
		var changed Definition
		json.Unmarshal(raw(d), &changed)
		mutation(&changed)
		if changed.Validate() == nil {
			t.Fatal("conflicting selector accepted")
		}
	}
	unselected := d
	unselected.Source.ExactStream = nil
	if Hash(unselected) == Hash(d) {
		t.Fatal("selector not part of revision")
	}
	empty, e := Evaluate(q, nil, time.Unix(103, 0), "", nil)
	if e != nil || empty.State != "no_data" || len(empty.Counters.Series) != 0 {
		t.Fatal("missing stream became zero", e, empty)
	}
}

func TestExactStreamAcrossNativeKinds(t *testing.T) {
	for _, kind := range []string{"gauge", "counter", "histogram"} {
		t.Run(kind, func(t *testing.T) {
			d := definition(t, "checkout-calls")
			d.Kind = kind
			if kind == "histogram" {
				d.Source.Unit = "ms"
			}
			row := gaugeRow(d, "one", 101000, "3")
			row.StartNS = ns(100000)
			row.Monotonic = true
			row.Temporality = 1
			row.Count = "3"
			row.Bounds = []float64{1}
			row.BucketCounts = []string{"1", "2"}
			d.Source.ExactStream = &metrics.Stream{Resource: row.Resource, Scope: row.Scope, Version: row.Version, SchemaURL: row.SchemaURL, Name: row.Name, Unit: row.Unit, Attributes: row.Attributes}
			q := Question{Definition: d, FromMS: 100000, ToMS: 102000}
			result, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(103, 0), "", nil)
			if e != nil {
				t.Fatal(e)
			}
			if e = ValidateResult(result); e != nil {
				t.Fatal(e)
			}
			row.ResourceSchemaURL = "other"
			if _, e = Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(103, 0), "", nil); e == nil {
				t.Fatal("resource schema changed selected identity")
			}
		})
	}
}
