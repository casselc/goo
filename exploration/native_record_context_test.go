package exploration

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

type nativeContextFixture struct {
	Name     string                `json:"name"`
	Parent   measurements.Question `json:"parent"`
	Stream   metrics.Stream        `json:"stream"`
	Signal   string                `json:"signal"`
	FromNS   string                `json:"fromNS"`
	ToNS     string                `json:"toNS"`
	Expected *NativeRecordContext  `json:"expected,omitempty"`
	Refused  bool                  `json:"refused,omitempty"`
}

func TestNativeRecordContextFixtures(t *testing.T) {
	d := dashboardTestInstrument("context", "calls", "counter", "1").Definition
	parent := measurements.Question{Definition: *d, FromMS: 100000, ToMS: 102001, FromNS: "100000000123", ToNS: "102000000123"}
	stream := metrics.Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "instance-1", "deployment.environment.name": "prod", "host.name": "not-copied"}, Scope: d.Source.Scope, Name: d.Source.Metric, Unit: d.Source.Unit, Attributes: map[string]string{"metric.only": "not-copied"}}
	fixtures := []nativeContextFixture{}
	for _, ns := range []*identity.Namespace{nil, {Mode: "value", Value: ""}, {Mode: "value", Value: "payments"}} {
		p := parent
		s := stream
		s.Resource = map[string]string{}
		for k, v := range stream.Resource {
			s.Resource[k] = v
		}
		if ns != nil {
			s.Resource[identity.NamespaceKey] = ns.Value
		}
		p.Definition.Source.Namespace = ns
		for _, signal := range []string{"logs", "spans"} {
			fixtures = append(fixtures, nativeContextFixture{Name: signal + "-namespace-" + func() string {
				if ns == nil {
					return "missing"
				}
				return ns.Value
			}(), Parent: p, Stream: s, Signal: signal, FromNS: "100000000124", ToNS: "101000000001"})
		}
	}
	histogram := parent
	histogram.Definition.Kind = "histogram"
	histogram.Definition.Source.Unit = "s"
	histogramStream := stream
	histogramStream.Unit = "s"
	fixtures = append(fixtures, nativeContextFixture{Name: "histogram-interval", Parent: histogram, Stream: histogramStream, Signal: "spans", FromNS: "100000000124", ToNS: "101000000001"})
	gauge := parent
	gauge.Definition.Kind = "gauge"
	fixtures = append(fixtures, nativeContextFixture{Name: "gauge-point", Parent: gauge, Stream: stream, Signal: "logs", FromNS: "100000000124", ToNS: "100000000125"})
	for _, bounds := range [][2]string{{"0100000000124", "101000000001"}, {"100000000122", "101000000001"}, {"100000000124", "102000000124"}, {"100000000124", "100000000124"}, {"100000000124", "18446744073709551615"}} {
		fixtures = append(fixtures, nativeContextFixture{Name: "invalid-bounds", Parent: parent, Stream: stream, Signal: "logs", FromNS: bounds[0], ToNS: bounds[1], Refused: true})
	}
	fixtures = append(fixtures, nativeContextFixture{Name: "gauge-no-inferred-window", Parent: gauge, Stream: stream, Signal: "logs", FromNS: "100000000124", ToNS: "101000000001", Refused: true})
	mismatch := stream
	mismatch.Resource = map[string]string{"service.name": "api", "service.instance.id": "instance-1", "deployment.environment.name": "other"}
	fixtures = append(fixtures, nativeContextFixture{Name: "environment-mismatch", Parent: parent, Stream: mismatch, Signal: "spans", FromNS: "100000000124", ToNS: "101000000001", Refused: true})
	for _, key := range []string{"service.instance.id", "service.name"} {
		bad := stream
		bad.Resource = map[string]string{}
		for k, v := range stream.Resource {
			bad.Resource[k] = v
		}
		delete(bad.Resource, key)
		fixtures = append(fixtures, nativeContextFixture{Name: "missing-" + key, Parent: parent, Stream: bad, Signal: "logs", FromNS: "100000000124", ToNS: "101000000001", Refused: true})
	}
	pinned := parent
	original := stream
	original.Attributes = map[string]string{"metric.only": "pinned"}
	pinned.Definition.Source.ExactStream = &original
	fixtures = append(fixtures, nativeContextFixture{Name: "pinned-stream-mismatch", Parent: pinned, Stream: stream, Signal: "logs", FromNS: "100000000124", ToNS: "101000000001", Refused: true})
	// A too-wide parent is refused before deriving any record query.
	wide := parent
	wide.FromMS = 100000
	wide.ToMS = 7300001
	wide.FromNS = ""
	wide.ToNS = "" // Parent itself refuses >2h.
	fixtures = append(fixtures, nativeContextFixture{Name: "wide-parent", Parent: wide, Stream: stream, Signal: "spans", FromNS: "100000000001", ToNS: "7300000000001", Refused: true})
	for i := range fixtures {
		f := &fixtures[i]
		got, e := PlanNativeRecordContext(f.Parent, f.Stream, f.Signal, f.FromNS, f.ToNS)
		if f.Refused {
			if e == nil {
				t.Fatalf("%s accepted", f.Name)
			}
			continue
		}
		if e != nil {
			t.Fatalf("%s: %v", f.Name, e)
		}
		f.Expected = got
		if _, e := Compile(got.Question); e != nil {
			t.Fatal(e)
		}
		v := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "resource", Key: "service.instance.id"}: "instance-1", {Kind: "column", Key: "timestamp_ns"}: f.FromNS}
		if ns, ok := f.Stream.Resource[identity.NamespaceKey]; ok {
			v[Field{Kind: "resource", Key: identity.NamespaceKey}] = ns
		}
		if !Match(got.Question.Where, v) {
			t.Fatalf("%s boundary not included", f.Name)
		}
		v[Field{Kind: "column", Key: "timestamp_ns"}] = f.ToNS
		if Match(got.Question.Where, v) {
			t.Fatal("exclusive endpoint included")
		}
	}
	path := "../tests/fixtures/native-record-context.json"
	if os.Getenv("GOO_UPDATE_NATIVE_CONTEXT_FIXTURES") == "1" {
		b, _ := json.MarshalIndent(fixtures, "", "  ")
		if e := os.WriteFile(path, append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var frozen []nativeContextFixture
	if e = json.Unmarshal(b, &frozen); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(fixtures, frozen) {
		t.Fatal("shared context fixtures drift")
	}
}

func TestNativeRecordContextRejectsResourceBinding(t *testing.T) {
	d := dashboardTestInstrument("host-context", "cpu", "gauge", "1").Definition
	d.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "one"}
	_, e := PlanNativeRecordContext(measurements.Question{Definition: *d, FromMS: 1000, ToMS: 2000}, metrics.Stream{}, "spans", "1000000000", "1000000001")
	if e == nil || !strings.Contains(e.Error(), "resource-bound") {
		t.Fatal("resource measurement gained service context", e)
	}
}
