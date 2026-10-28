package verification

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
	"time"
)

func TestHegelInstrumentCapabilityHints(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		kind := []string{"gauge", "counter", "histogram"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		unit := []string{"s", "ms", "By", ""}[hegel.Draw(tc, hegel.Integers(0, 3))]
		n := hegel.Draw(tc, hegel.Integers(1, 70))
		metric, key := "http.server.request.duration", "http.request.method"
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			metric, key = "messaging.process.duration", "messaging.system"
		}
		i := measurements.Instrument{Metric: metric, Kind: kind, Unit: unit, Scope: "ordinary", Supported: true, Definition: &measurements.Definition{}}
		observations := []measurements.DiscoveryObservation{}
		for k := 0; k < n; k++ {
			attrs := map[string]string{}
			switch k % 3 {
			case 0:
				attrs[key] = "GET"
			case 1:
				attrs[key] = ""
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

// Vary actual evidence without changing the metadata candidate. Successful
// planning must not hide missing identity, missing values, or incomplete time.
func TestHegelInstrumentCandidateDoesNotPromoteEvidence(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		metric := []string{"http.server.request.duration", "messaging.process.duration", "custom.duration"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		unit := []string{"s", "ms"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		count := hegel.Draw(tc, hegel.Integers(1, 100000))
		change := hegel.Draw(tc, hegel.Integers(0, 4))
		d := measurements.Definition{APIVersion: measurements.Version, Name: "candidate", Title: "Candidate", Kind: "histogram", Environment: "test", LookbackMinutes: 2, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"ordinary"}, Scope: "ordinary", Metric: metric, Unit: unit}}
		resource := map[string]string{"service.name": "ordinary", "service.instance.id": "one", "deployment.environment.name": "test"}
		i := measurements.Instrument{Metric: metric, Kind: "histogram", Unit: unit, Scope: "ordinary", Temporality: 1, Supported: true, Definition: &d}
		observations := []measurements.DiscoveryObservation{{Resource: resource, Attributes: map[string]string{}}}
		plan := measurements.PlanInstrument(i, observations)
		before := measurements.Hash(plan)
		row := map[string]any{"resource": resource, "scope": "ordinary", "name": metric, "unit": unit, "attributes": map[string]string{}, "temporality": 1, "start_ns": "1000000000", "end_ns": "3000000000", "count": fmtInt(count), "bounds": []float64{1}, "bucket_counts": []string{"0", fmtInt(count)}}
		switch change {
		case 1:
			row["end_ns"] = "2000000000"
		case 2:
			row["flags"] = 1
		case 3:
			resource["service.instance.id"] = ""
		case 4:
			row["count"] = fmtInt(count + 1)
		}
		encoded, _ := json.Marshal(row)
		result, err := measurements.Evaluate(measurements.Question{Definition: d, FromMS: 1000, ToMS: 3000}, []json.RawMessage{encoded}, time.UnixMilli(4000))
		if change >= 3 {
			if err == nil || result != nil {
				tc.Fatal("invalid evidence accepted", change)
			}
		} else {
			want := "observed"
			if change > 0 {
				want = "unknown"
			}
			if err != nil || result.State != want {
				tc.Fatal("candidate concealed evidence state", change, err)
			}
		}
		if measurements.Hash(plan) != before || !plan.Eligible || plan.Verification != "not-run" {
			tc.Fatal("Run promoted or changed metadata plan")
		}
	}, options(100)...)
}
