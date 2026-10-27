package metrics

import (
	"fmt"
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
	return evaluateCategoricalStates(fromNS, toNS, maxAgeNS, points, PodPhaseProfile, ValidatePodPhaseStream,
		func(s Stream) bool { return s.Resource["k8s.pod.uid"] == podUID })
}

// EvaluateContainerReadiness observes one reported container ID. Pod UID and
// container name remain mandatory context, never a synthesized instance ID.
func EvaluateContainerReadiness(fromNS, toNS, maxAgeNS, containerID string, points []GaugePoint) (*StateResult, error) {
	if err := validateHostObservationID(containerID); err != nil {
		return nil, fmt.Errorf("require a nonempty bounded container.id")
	}
	return evaluateCategoricalStates(fromNS, toNS, maxAgeNS, points, ContainerReadyProfile, ValidateContainerReadyStream,
		func(s Stream) bool {
			return s.Resource["container.id"] == containerID && validateHostObservationID(s.Resource["k8s.pod.uid"]) == nil && validateHostObservationID(s.Resource["k8s.container.name"]) == nil
		})
}

func evaluateCategoricalStates(fromNS, toNS, maxAgeNS string, points []GaugePoint, profile string, validate func(Stream) error, identity func(Stream) bool) (*StateResult, error) {
	if len(points) > 10000 {
		return nil, fmt.Errorf("require at most 10000 state points")
	}
	canonical := make([]GaugePoint, len(points))
	for i, p := range points {
		if err := validate(p.Stream); err != nil {
			return nil, err
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
		identity, true)
	if err != nil {
		return nil, err
	}
	result := &StateResult{FromNS: raw.FromNS, ToNS: raw.ToNS, Series: []StateSeries{}}
	for _, series := range raw.Series {
		out := StateSeries{Stream: series.Stream, State: series.State, Samples: []StateSample{}}
		for _, sample := range series.Samples {
			decoded := StateSample{AtNS: sample.AtNS, Code: sample.Value, State: sample.State, Reason: sample.Reason}
			if sample.Value != nil {
				v := *sample.Value
				label, supported := StateCodeLabel(profile, v)
				if !supported {
					decoded.CodeState = "unsupported"
				} else {
					decoded.CodeState = "supported"
					decoded.Label = label
				}
			}
			out.Samples = append(out.Samples, decoded)
			out.State = decoded.State
		}
		result.Series = append(result.Series, out)
	}
	return result, nil
}

// PodPhaseProfile pins the domain mapping separately from its physical gauge kind.
const PodPhaseProfile = "k8scluster-pod-phase/0.137.0"

// ValidatePodPhaseStream checks the physical profile, not resource authority.
func ValidatePodPhaseStream(s Stream) error {
	if s.Name != "k8s.pod.phase" || s.Unit != "" || s.Scope != "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver" || s.Version != "0.137.0" || s.SchemaURL != "" {
		return fmt.Errorf("unsupported pod phase profile")
	}
	return nil
}

const ContainerReadyProfile = "k8scluster-container-ready/0.137.0"

func ValidateContainerReadyStream(s Stream) error {
	if s.Name != "k8s.container.ready" || s.Unit != "" || s.Scope != "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver" || s.Version != "0.137.0" || s.SchemaURL != "" {
		return fmt.Errorf("unsupported container readiness profile")
	}
	return nil
}

// StateCodeLabel maps only explicitly supported domain codes, independently of
// telemetry quality. It supplies no health or aggregation semantics.
func StateCodeLabel(profile string, code float64) (string, bool) {
	switch profile {
	case PodPhaseProfile:
		labels := map[float64]string{1: "Pending", 2: "Running", 3: "Succeeded", 4: "Failed", 5: "Unknown"}
		label, ok := labels[code]
		return label, ok
	case ContainerReadyProfile:
		switch code {
		case 0:
			return "Not ready", true
		case 1:
			return "Ready", true
		}
	}
	return "", false
}
