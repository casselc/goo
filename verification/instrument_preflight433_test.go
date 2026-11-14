package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// The generator chooses the known identity defect independently of the preview
// and then permutes all observations. Admission must not follow the visible25.
func TestHegelInstrumentInstancePreflight433(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(26, 80))
		mode := hegel.Draw(tc, hegel.Integers(0, 8))
		supported := hegel.Draw(tc, hegel.Booleans())
		binding := hegel.Draw(tc, hegel.Booleans())
		kind := hegel.Draw(tc, hegel.SampledFrom([]string{"gauge", "counter", "histogram"}))
		unit := "1"
		if kind == "histogram" {
			unit = "s"
		}
		instrument := measurements.Instrument{Kind: kind, Metric: "custom.preflight", Scope: "ordinary", Unit: unit, Supported: supported, Temporality: 1, Monotonic: true}
		if binding {
			instrument.Definition = &measurements.Definition{Source: measurements.Binding{Services: []string{"api"}}}
		}
		observations := make([]measurements.DiscoveryObservation, count)
		for index := range observations {
			observations[index] = measurements.DiscoveryObservation{Resource: map[string]string{"service.name": "api", "service.instance.id": fmt.Sprint(index)}, Attributes: map[string]string{"case": fmt.Sprint(index)}}
		}
		index := hegel.Draw(tc, hegel.Integers(0, count-1))
		knownInvalid := mode == 1 || mode == 2 || mode == 3 || mode == 6 || mode == 8
		switch mode {
		case 1:
			delete(observations[index].Resource, "service.instance.id")
		case 2:
			observations[index].Resource["service.instance.id"] = ""
		case 3:
			delete(observations[index].Resource, "service.instance.id")
			observations[(index+1)%count].Resource["service.instance.id"] = ""
		case 4:
			observations = nil
		case 5:
			observations = append(observations, measurements.DiscoveryObservation{})
		case 6:
			delete(observations[index].Resource, "service.instance.id")
			observations = append(observations, measurements.DiscoveryObservation{})
		case 7:
			observations[index].Attributes = nil
		case 8:
			observations[index].Resource = map[string]string{}
			observations[index].Attributes = nil
		}
		before, _ := json.Marshal([]any{instrument, observations})
		plan := measurements.PlanInstrument(instrument, observations)
		after, _ := json.Marshal([]any{instrument, observations})
		if string(before) != string(after) {
			tc.Fatal("preflight mutated source identity or binding")
		}
		if plan.Eligible != (supported && binding && !knownInvalid) || plan.Verification != "not-run" {
			tc.Fatal("known identity policy changed", mode, supported, binding, plan)
		}
		if knownInvalid && binding {
			if len(plan.Reasons) == 0 || !strings.HasPrefix(plan.Reasons[0], "Current service evaluation policy requires reported nonempty service.instance.id;") || !strings.HasSuffix(plan.Reasons[0], "This is not an OTLP requirement.") {
				tc.Fatal("policy reason missing or confused with OTel", plan)
			}
		} else {
			for _, reason := range plan.Reasons {
				if strings.Contains(reason, "service.instance.id") {
					tc.Fatal("unknown identity became known invalid", mode, plan)
				}
			}
		}
		reordered := slices.Clone(observations)
		slices.Reverse(reordered)
		if len(reordered) > 1 {
			reordered = append(reordered[1:], reordered[0])
		}
		if other := measurements.PlanInstrument(instrument, reordered); !reflect.DeepEqual(plan, other) {
			tc.Fatal("observation order changed full-population preflight", mode, plan, other)
		}
		if mode == 4 || mode == 5 || mode == 6 || mode == 7 || mode == 8 {
			if plan.Coverage.State != "unavailable" {
				tc.Fatal("unknown coverage invented", plan.Coverage)
			}
		} else if plan.Coverage.TotalStreams != count || len(plan.Coverage.Streams) != 25 || plan.Coverage.OmittedStreams != count-25 {
			tc.Fatal("display cap altered full observations", plan.Coverage)
		}
	}, options(100)...)
}
