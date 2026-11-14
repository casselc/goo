package measurements

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/metrics"
)

func preflightInstrument433() Instrument {
	return Instrument{Kind: "gauge", Metric: "custom.load", Scope: "ordinary", Unit: "1", Supported: true, Definition: &Definition{Source: Binding{Services: []string{"api"}}}}
}
func preflightObservation433(instance *string, key string) DiscoveryObservation {
	resource := map[string]string{"service.name": "api"}
	if instance != nil {
		resource["service.instance.id"] = *instance
	}
	return DiscoveryObservation{Resource: resource, Attributes: map[string]string{"case": key}}
}
func preflightString433(value string) *string { return &value }
func TestInstrumentServiceInstancePreflight433(t *testing.T) {
	instrument := preflightInstrument433()
	good := preflightObservation433(preflightString433("one"), "good")
	missing := preflightObservation433(nil, "missing")
	empty := preflightObservation433(preflightString433(""), "empty")
	cases := []struct {
		name         string
		observations []DiscoveryObservation
		eligible     bool
		reason       string
	}{
		{"all-present", []DiscoveryObservation{good}, true, ""},
		{"missing", []DiscoveryObservation{missing}, false, "missing values were observed"},
		{"empty", []DiscoveryObservation{empty}, false, "empty values were observed"},
		{"mixed-known", []DiscoveryObservation{good, missing, empty}, false, "missing and empty values were observed"},
		{"no-observations", nil, true, ""},
		{"unknown-resource", []DiscoveryObservation{{Attributes: map[string]string{}}}, true, ""},
		{"unknown-dimensions", []DiscoveryObservation{{Resource: map[string]string{"service.instance.id": "one"}}}, true, ""},
		{"known-missing-with-unknown-dimensions", []DiscoveryObservation{{Resource: map[string]string{}}}, false, "missing values were observed"},
		{"unknown-does-not-hide-known-missing", []DiscoveryObservation{{}, missing}, false, "missing values were observed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := PlanInstrument(instrument, tc.observations)
			if p.Eligible != tc.eligible {
				t.Fatal(p)
			}
			if tc.reason != "" {
				if len(p.Reasons) == 0 || !strings.Contains(p.Reasons[0], tc.reason) || !strings.HasPrefix(p.Reasons[0], "Current service evaluation policy requires reported nonempty service.instance.id;") || !strings.HasSuffix(p.Reasons[0], "This is not an OTLP requirement.") {
					t.Fatal("missing decisive policy reason", p.Reasons)
				}
			}
			if tc.reason == "" {
				for _, reason := range p.Reasons {
					if strings.Contains(reason, "service.instance.id") {
						t.Fatal("unknown became missing", p)
					}
				}
			}
			if tc.name == "no-observations" || tc.name == "unknown-resource" || tc.name == "unknown-dimensions" {
				if p.Coverage.State != "unavailable" {
					t.Fatal("unknown coverage invented", p.Coverage)
				}
			}
		})
	}
}
func outsidePreviewObservation433(t *testing.T, i Instrument, good []DiscoveryObservation, instance *string, key string) DiscoveryObservation {
	t.Helper()
	ids := make([]string, 0, len(good))
	hash := func(o DiscoveryObservation) string {
		return Hash(metrics.Stream{Resource: o.Resource, Attributes: o.Attributes, ResourceSchemaURL: o.ResourceSchemaURL, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, Name: i.Metric, Unit: i.Unit})
	}
	for _, o := range good {
		ids = append(ids, hash(o))
	}
	slices.Sort(ids)
	threshold := ids[DiscoveryStreamLimit-1]
	for n := 0; n < 1000; n++ {
		o := preflightObservation433(instance, fmt.Sprintf("%s-%d", key, n))
		if hash(o) > threshold {
			return o
		}
	}
	t.Fatal("could not construct omitted-stream witness")
	return DiscoveryObservation{}
}
func TestInstrumentPreflightIgnoresDisplayCapAndOrder433(t *testing.T) {
	instrument := preflightInstrument433()
	good := []DiscoveryObservation{}
	for n := 0; n < 60; n++ {
		good = append(good, preflightObservation433(preflightString433(fmt.Sprint(n)), fmt.Sprint(n)))
	}
	missing := outsidePreviewObservation433(t, instrument, good, nil, "missing")
	empty := outsidePreviewObservation433(t, instrument, good, preflightString433(""), "empty")
	for _, bad := range [][]DiscoveryObservation{{missing}, {empty}, {missing, empty}} {
		observations := append(slices.Clone(good), bad...)
		baseline := PlanInstrument(instrument, observations)
		if baseline.Eligible || baseline.Coverage.TotalStreams != len(observations) || len(baseline.Coverage.Streams) != DiscoveryStreamLimit || baseline.Coverage.OmittedStreams != len(observations)-DiscoveryStreamLimit {
			t.Fatal("cap changed admission or coverage", baseline)
		}
		for _, stream := range baseline.Coverage.Streams {
			if stream.Instance != "present" {
				t.Fatal("invalid witness was not beyond display cap", stream)
			}
		}
		small := PlanInstrument(instrument, bad)
		if small.Eligible != baseline.Eligible || !reflect.DeepEqual(small.Reasons, baseline.Reasons) {
			t.Fatal("display-only streams changed refusal", small, baseline)
		}
		for turn := 0; turn < 5; turn++ {
			slices.Reverse(observations)
			observations = append(observations[1:], observations[0])
			if p := PlanInstrument(instrument, observations); !reflect.DeepEqual(p, baseline) {
				t.Fatal("permutation changed preflight", p, baseline)
			}
		}
	}
}
func TestInstrumentPreflightPreservesProfilesAndKernelGuards433(t *testing.T) {
	invalid := []DiscoveryObservation{preflightObservation433(nil, "missing")}
	for _, metric := range []string{"custom.load", "http.server.request.duration", CurrentHTTPClientDurationMetric, CurrentRPCDurationMetric, "messaging.process.duration", "go.memory.used", "db.client.connection.count"} {
		instrument := preflightInstrument433()
		instrument.Metric = metric
		instrument.Unit = "s"
		instrument.Kind = "histogram"
		instrument.Temporality = 1
		p := PlanInstrument(instrument, invalid)
		if p.Eligible || len(p.Reasons) == 0 || !strings.HasPrefix(p.Reasons[0], "Current service evaluation policy") {
			t.Fatal("profile returned before preflight", metric, p)
		}
	}
	instrument := preflightInstrument433()
	instrument.Kind = exponentialHistogramKind
	if p := PlanInstrument(instrument, invalid); p.Eligible || !slices.Contains(p.Reasons, exponentialHistogramUnsupportedReason) {
		t.Fatal("kernel refusal was replaced", p)
	}
	instrument = preflightInstrument433()
	instrument.Supported = false
	if p := PlanInstrument(instrument, []DiscoveryObservation{preflightObservation433(preflightString433("one"), "ok")}); p.Eligible {
		t.Fatal("preflight promoted unsupported binding", p)
	}
	instrument = preflightInstrument433()
	instrument.Supported = false
	if p := PlanInstrument(instrument, invalid); p.Eligible || !slices.Contains(p.Reasons, "No valid package binding is available from the current kernels.") {
		t.Fatal("instance preflight hid an existing kernel refusal", p)
	}
	instrument = preflightInstrument433()
	instrument.Definition = nil
	if p := PlanInstrument(instrument, invalid); p.Eligible {
		t.Fatal("preflight invented binding", p)
	}
	instrument = preflightInstrument433()
	instrument.Definition.Source.Resource = &ResourceBinding{Kind: "host", ID: "host-one"}
	if p := PlanInstrument(instrument, invalid); !p.Eligible || strings.Contains(strings.Join(p.Reasons, " "), "service.instance.id") {
		t.Fatal("service policy applied to separate resource identity policy", p)
	}
}
