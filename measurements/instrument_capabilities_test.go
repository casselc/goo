package measurements

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInstrumentPlanObservedStreamPresenceAndStableBindings(t *testing.T) {
	q := DiscoveryQuestion{Environment: "test", Service: "api", FromMS: 1000, ToMS: 2000}
	row := func(instance string, attrs map[string]string) json.RawMessage {
		b, _ := json.Marshal(map[string]any{"kind": "histogram", "scope": "http", "metric": "http.server.request.duration", "unit": "s", "schemaURL": HTTPSchema, "temporality": 1, "dimensions": func() []string {
			r := []string{}
			for k := range attrs {
				r = append(r, k)
			}
			return r
		}(), "resource": map[string]string{"service.name": "api", "service.instance.id": instance}, "attributes": attrs})
		return b
	}
	rows := []json.RawMessage{row("a", map[string]string{"http.request.method": "GET", "http.response.status_code": "200"}), row("a", map[string]string{"http.request.method": "GET", "http.response.status_code": "200"}), row("b", map[string]string{"http.request.method": ""}), row("c", map[string]string{})}
	r, e := discoverRows(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	i := r.Instruments[0]
	p := i.CapabilityPlan
	if len(r.Instruments) != 1 || p.Profile != "http-server-duration" || !p.ShapeMatches || !p.Eligible || p.Verification != "not-run" || p.Coverage.TotalStreams != 3 {
		t.Fatalf("%+v", p)
	}
	for _, d := range p.Coverage.Dimensions {
		if d.Present+d.Empty+d.Missing != 3 {
			t.Fatal(d)
		}
		if d.Key == "http.request.method" && (d.Present != 1 || d.Empty != 1 || d.Missing != 1) {
			t.Fatal(d)
		}
	}
	reversed := []json.RawMessage{rows[3], rows[2], rows[1], rows[0]}
	other, e := discoverRows(q, reversed)
	if e != nil || !reflect.DeepEqual(r, other) {
		t.Fatal("row order changed plan", e)
	}
	legacy := []json.RawMessage{json.RawMessage(`{"kind":"histogram","scope":"http","metric":"http.server.request.duration","unit":"s","schemaURL":"https://opentelemetry.io/schemas/1.30.0","temporality":1,"dimensions":["http.request.method","http.response.status_code"]}`)}
	old, e := discoverRows(q, legacy)
	if e != nil || !reflect.DeepEqual(i.Definition, old.Instruments[0].Definition) {
		t.Fatal("presence metadata changed binding", e)
	}
	if old.Instruments[0].CapabilityPlan.Coverage.State != "unavailable" {
		t.Fatal("invented stream coverage")
	}
}
func TestInstrumentCoverageCapDoesNotChangeTotals(t *testing.T) {
	i := Instrument{Metric: "http.server.request.duration", Kind: "histogram", Unit: "s"}
	obs := []DiscoveryObservation{}
	for n := 0; n < DiscoveryStreamLimit+8; n++ {
		obs = append(obs, DiscoveryObservation{Resource: map[string]string{"service.instance.id": string(rune('a' + n))}, Attributes: map[string]string{"http.request.method": "GET"}})
	}
	p := PlanInstrument(i, obs)
	if p.Coverage.TotalStreams != 33 || len(p.Coverage.Streams) != 25 || p.Coverage.OmittedStreams != 8 {
		t.Fatal(p.Coverage)
	}
	if p.Coverage.Dimensions[0].Present != 33 {
		t.Fatal("display cap altered coverage totals")
	}
	p.Coverage.Streams[0].Stream.Attributes["http.request.method"] = "changed"
	for _, o := range obs {
		if o.Attributes["http.request.method"] != "GET" {
			t.Fatal("aliased source metadata")
		}
	}
}

func TestInstrumentCoverageMissingAndEmptyInstancesStayDistinct(t *testing.T) {
	observations := []DiscoveryObservation{{Resource: map[string]string{}, Attributes: map[string]string{}}, {Resource: map[string]string{"service.instance.id": ""}, Attributes: map[string]string{}}}
	p := PlanInstrument(Instrument{}, observations)
	if p.Coverage.TotalStreams != 2 {
		t.Fatal("merged missing and empty instances")
	}
	states := map[string]bool{}
	for _, s := range p.Coverage.Streams {
		states[s.Instance] = true
	}
	if !states["missing"] || !states["empty"] {
		t.Fatal(states)
	}
	observations = append(observations, DiscoveryObservation{})
	p = PlanInstrument(Instrument{}, observations)
	if p.Coverage.State != "unavailable" || p.Coverage.TotalStreams != 0 {
		t.Fatal("partial identity became complete coverage")
	}
}

func TestMessagingProcessingProfilePreservesGenericBindingsAndConditionalPresence(t *testing.T) {
	q := DiscoveryQuestion{Environment: "test", Service: "consumer", FromMS: 1000, ToMS: 2000}
	for _, unit := range []string{"s", "ms"} {
		rows := []json.RawMessage{}
		for _, attrs := range []map[string]string{{"messaging.system": "kafka", "messaging.operation.type": "process"}, {"messaging.system": "", "messaging.destination.name": "orders"}, {}} {
			b, _ := json.Marshal(map[string]any{"kind": "histogram", "scope": "ordinary.messaging", "metric": "messaging.process.duration", "unit": unit, "temporality": 1, "resource": map[string]string{"service.name": "consumer"}, "attributes": attrs})
			rows = append(rows, b)
		}
		result, err := discoverRows(q, rows)
		if err != nil {
			t.Fatal(err)
		}
		i := result.Instruments[0]
		p := i.CapabilityPlan
		if p.Profile != "messaging-process-duration" || p.ShapeMatches != (unit == "s") || !p.Eligible || p.Verification != "not-run" {
			t.Fatalf("%+v", p)
		}
		if p.Coverage.TotalStreams != 3 || len(p.Coverage.Dimensions) != 6 {
			t.Fatal(p.Coverage)
		}
		for _, d := range p.Coverage.Dimensions {
			if d.Present+d.Empty+d.Missing != 3 {
				t.Fatal(d)
			}
			if d.Key == "messaging.system" && (d.Present != 1 || d.Empty != 1 || d.Missing != 1) {
				t.Fatal(d)
			}
		}
		legacy := []json.RawMessage{json.RawMessage(`{"kind":"histogram","scope":"ordinary.messaging","metric":"messaging.process.duration","unit":"` + unit + `","temporality":1}`)}
		old, err := discoverRows(q, legacy)
		if err != nil || !reflect.DeepEqual(i.Definition, old.Instruments[0].Definition) {
			t.Fatal("profile changed binding", err)
		}
	}
	i := Instrument{Metric: "messaging.process.duration", Kind: "counter", Unit: "s"}
	if p := PlanInstrument(i, nil); p.ShapeMatches || p.Eligible {
		t.Fatal(p)
	}
	if p := PlanInstrument(Instrument{Metric: "custom.messaging.process.duration"}, nil); p.Profile != "generic" || len(p.Coverage.Dimensions) != 0 {
		t.Fatal(p)
	}
}
