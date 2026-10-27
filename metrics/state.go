package metrics

import (
	"fmt"
	"math"
)

// StateSample separates observation quality from the reported domain label.
// In particular, a freshly reported phase Unknown is still observed telemetry.
type StateSample struct {
	AtNS      string   `json:"atNS"`
	Code      *float64 `json:"code,omitempty"`
	Label     string   `json:"label,omitempty"`
	CodeState string   `json:"codeState,omitempty"`
	State     string   `json:"state"`
	Reason    string   `json:"reason,omitempty"`
}

type StateSeries struct {
	Stream  Stream        `json:"stream"`
	State   string        `json:"state"`
	Samples []StateSample `json:"samples"`
}

type StateResult struct {
	FromNS string        `json:"fromNS"`
	ToNS   string        `json:"toNS"`
	Series []StateSeries `json:"series"`
}

// EvaluatePodPhases decodes only the pinned k8sclusterreceiver v0.137.0
// unitless gauge profile. It does not infer pod health or container readiness.
// Callers remain responsible for selecting known-schema exact resource variants.
func EvaluatePodPhases(fromNS, toNS, maxAgeNS, podUID string, points []GaugePoint) (*StateResult, error) {
	if err := validateHostObservationID(podUID); err != nil {
		return nil, fmt.Errorf("require a nonempty bounded pod UID")
	}
	if len(points) > 10000 {
		return nil, fmt.Errorf("require at most 10000 state points")
	}
	canonical := make([]GaugePoint, len(points))
	for i, p := range points {
		if p.Stream.Name != "k8s.pod.phase" || p.Stream.Unit != "" ||
			p.Stream.Scope != "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver" ||
			p.Stream.Version != "0.137.0" || p.Stream.SchemaURL != "" {
			return nil, fmt.Errorf("unsupported pod phase profile")
		}
		p.Stream = cloneStream(p.Stream)
		if p.Value == 0 {
			p.Value = 0
		}
		canonical[i] = p
	}
	// Reuse timestamp, conflict, no-recorded-value and freshness semantics;
	// ordinary scalar gauges still require a nonempty unit.
	raw, err := evaluateGaugeObservations(fromNS, toNS, maxAgeNS, canonical,
		func(s Stream) bool { return s.Resource["k8s.pod.uid"] == podUID }, true)
	if err != nil {
		return nil, err
	}
	result := &StateResult{FromNS: raw.FromNS, ToNS: raw.ToNS, Series: []StateSeries{}}
	labels := [...]string{"", "Pending", "Running", "Succeeded", "Failed", "Unknown"}
	for _, series := range raw.Series {
		out := StateSeries{Stream: series.Stream, State: series.State, Samples: []StateSample{}}
		for _, sample := range series.Samples {
			decoded := StateSample{AtNS: sample.AtNS, Code: sample.Value, State: sample.State, Reason: sample.Reason}
			if sample.Value != nil {
				v := *sample.Value
				if v < 1 || v > 5 || math.Trunc(v) != v {
					decoded.CodeState = "unsupported"
				} else {
					decoded.CodeState = "supported"
					decoded.Label = labels[int(v)]
				}
			}
			out.Samples = append(out.Samples, decoded)
			out.State = decoded.State
		}
		result.Series = append(result.Series, out)
	}
	return result, nil
}
