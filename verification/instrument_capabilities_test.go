package verification

import (
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
)

func TestHegelInstrumentCapabilityHints(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		kind := []string{"gauge", "counter", "histogram"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		unit := []string{"s", "ms", "By", ""}[hegel.Draw(tc, hegel.Integers(0, 3))]
		n := hegel.Draw(tc, hegel.Integers(1, 70))
		i := measurements.Instrument{Metric: "http.server.request.duration", Kind: kind, Unit: unit, Scope: "ordinary", Supported: true, Definition: &measurements.Definition{}}
		observations := []measurements.DiscoveryObservation{}
		for k := 0; k < n; k++ {
			attrs := map[string]string{}
			switch k % 3 {
			case 0:
				attrs["http.request.method"] = "GET"
			case 1:
				attrs["http.request.method"] = ""
			}
			observations = append(observations, measurements.DiscoveryObservation{Resource: map[string]string{"service.instance.id": fmtInt(k)}, Attributes: attrs})
		}
		p := measurements.PlanInstrument(i, observations)
		if p.ShapeMatches != (kind == "histogram" && unit == "s") || p.Verification != "not-run" || p.Coverage.TotalStreams != n {
			tc.Fatal(p)
		}
		for _, d := range p.Coverage.Dimensions {
			if d.Present+d.Empty+d.Missing != n {
				tc.Fatal(d)
			}
		}
		for a, b := 0, len(observations)-1; a < b; a, b = a+1, b-1 {
			observations[a], observations[b] = observations[b], observations[a]
		}
		if !reflect.DeepEqual(p, measurements.PlanInstrument(i, observations)) {
			tc.Fatal("order changed plan")
		}
		if len(p.Coverage.Streams) > 25 || len(p.Coverage.Streams)+p.Coverage.OmittedStreams != n {
			tc.Fatal("display limit changed total")
		}
		i.Definition = nil
		if measurements.PlanInstrument(i, observations).Eligible {
			tc.Fatal("invented binding")
		}
	}, options(100)...)
}
