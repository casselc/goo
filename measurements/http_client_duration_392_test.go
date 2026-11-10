package measurements

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestHTTPClientDuration392CapabilityAndCoverage(t *testing.T) {
	for _, schema := range []string{CurrentHTTPClientSchema, "", "https://custom.invalid/schema"} {
		i := Instrument{Metric: CurrentHTTPClientDurationMetric, Kind: "histogram", Unit: "s", SchemaURL: schema, Supported: true, Temporality: 1, Definition: &Definition{}}
		observations := []DiscoveryObservation{{Resource: map[string]string{"service.instance.id": "one"}, Attributes: map[string]string{"http.request.method": "GET", "server.address": "api", "server.port": "443"}}, {Resource: map[string]string{}, Attributes: map[string]string{"http.request.method": "", "server.address": "", "server.port": ""}}, {Resource: map[string]string{"service.instance.id": "three"}, Attributes: map[string]string{}}}
		before, _ := json.Marshal(observations)
		p := PlanInstrument(i, observations)
		after, _ := json.Marshal(observations)
		if string(before) != string(after) || p.Profile != "http-client-request-duration" || !p.ShapeMatches || !p.Eligible || p.RuleSet != CurrentHTTPClientRuleSet || p.Verification != "not-run" || p.Coverage.TotalStreams != 3 {
			t.Fatal(p)
		}
		for _, dimension := range p.Coverage.Dimensions {
			if strings.Contains("|http.request.method|server.address|server.port|", "|"+dimension.Key+"|") && (dimension.Present != 1 || dimension.Empty != 1 || dimension.Missing != 1) {
				t.Fatal(dimension)
			}
		}
		for _, stream := range p.Coverage.Streams {
			if stream.Stream.SchemaURL != schema {
				t.Fatal("schema migrated", stream)
			}
		}
		observations[0], observations[2] = observations[2], observations[0]
		if other := PlanInstrument(i, observations); !reflect.DeepEqual(p, other) {
			t.Fatal("discovery order changed capability")
		}
	}
	for _, i := range []Instrument{{Metric: CurrentHTTPClientDurationMetric, Kind: "histogram", Unit: "ms"}, {Metric: CurrentHTTPClientDurationMetric, Kind: "gauge", Unit: "s"}, {Metric: "http.client.duration", Kind: "histogram", Unit: "s"}} {
		if CurrentHTTPClientDurationShape(i) || PlanInstrument(i, nil).Profile == "http-client-request-duration" {
			t.Fatal("unsupported shape qualified", i)
		}
	}
}
