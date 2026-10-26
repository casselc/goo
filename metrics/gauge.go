package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// Gauges are independent observations, never cumulative increments or rates.
type GaugePoint struct {
	Stream          Stream  `json:"stream"`
	EndNS           string  `json:"endNS"`
	Value           float64 `json:"value"`
	NoRecordedValue bool    `json:"noRecordedValue,omitempty"`
}
type GaugeSample struct {
	AtNS   string   `json:"atNS"`
	Value  *float64 `json:"value,omitempty"`
	State  string   `json:"state"`
	Reason string   `json:"reason,omitempty"`
}
type GaugeSeries struct {
	Stream  Stream        `json:"stream"`
	State   string        `json:"state"`
	Samples []GaugeSample `json:"samples"`
}
type GaugeResult struct {
	FromNS string        `json:"fromNS"`
	ToNS   string        `json:"toNS"`
	Series []GaugeSeries `json:"series"`
}

func EvaluateGauges(fromText, toText, maxAgeText string, points []GaugePoint) (*GaugeResult, error) {
	return evaluateGauges(fromText, toText, maxAgeText, points, serviceObservationIdentity)
}

// EvaluateHostGauges is an explicit host-resource path; service identity remains
// mandatory for the original EvaluateGauges entry point.
func EvaluateHostGauges(fromText, toText, maxAgeText, hostID string, points []GaugePoint) (*GaugeResult, error) {
	if err := validateHostObservationID(hostID); err != nil {
		return nil, err
	}
	return evaluateResourceGauges(fromText, toText, maxAgeText, "host.id", hostID, points)
}

// EvaluatePodGauges opts into reported pod UID identity without supplying fake
// service or instance labels. Full resource isolation remains in each stream.
func EvaluatePodGauges(fromText, toText, maxAgeText, podUID string, points []GaugePoint) (*GaugeResult, error) {
	if err := validateHostObservationID(podUID); err != nil {
		return nil, fmt.Errorf("require a nonempty bounded pod UID")
	}
	return evaluateResourceGauges(fromText, toText, maxAgeText, "k8s.pod.uid", podUID, points)
}
func evaluateResourceGauges(fromText, toText, maxAgeText, key, id string, points []GaugePoint) (*GaugeResult, error) {
	canonical := make([]GaugePoint, len(points))
	for index, point := range points {
		point.Stream = cloneStream(point.Stream)
		if point.Value == 0 {
			point.Value = 0
		}
		canonical[index] = point
	}
	return evaluateGauges(fromText, toText, maxAgeText, canonical, func(s Stream) bool { return s.Resource[key] == id })
}
func evaluateGauges(fromText, toText, maxAgeText string, points []GaugePoint, identity func(Stream) bool) (*GaugeResult, error) {
	from, e := nano(fromText)
	if e != nil {
		return nil, e
	}
	to, e := nano(toText)
	if e != nil || to <= from {
		return nil, fmt.Errorf("require nonempty gauge window")
	}
	age, e := nano(maxAgeText)
	if e != nil || age <= 0 || len(points) > 10000 {
		return nil, fmt.Errorf("require positive freshness and at most 10000 points")
	}
	groups := map[string][]GaugePoint{}
	streams := map[string]Stream{}
	for _, p := range points {
		at, e := nano(p.EndNS)
		if !identity(p.Stream) || p.Stream.Scope == "" || p.Stream.Name == "" || p.Stream.Unit == "" || e != nil || at < from || at >= to || math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
			return nil, fmt.Errorf("invalid gauge observation")
		}
		encoded, _ := json.Marshal(p.Stream)
		key := string(encoded)
		groups[key] = append(groups[key], p)
		streams[key] = cloneStream(p.Stream)
	}
	out := &GaugeResult{FromNS: fromText, ToNS: toText, Series: []GaugeSeries{}}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		ps := groups[k]
		sort.Slice(ps, func(i, j int) bool { a, _ := nano(ps[i].EndNS); b, _ := nano(ps[j].EndNS); return a < b })
		s := GaugeSeries{Stream: streams[k], State: "unknown", Samples: []GaugeSample{}}
		for i := 0; i < len(ps); {
			p := ps[i]
			end := i + 1
			conflict := false
			for end < len(ps) && ps[end].EndNS == p.EndNS {
				conflict = conflict || ps[end].Value != p.Value || ps[end].NoRecordedValue != p.NoRecordedValue
				end++
			}
			at, _ := nano(p.EndNS)
			sample := GaugeSample{AtNS: p.EndNS, State: "observed"}
			switch {
			case conflict:
				sample.State = "unknown"
				sample.Reason = "conflicting observations at one timestamp"
			case p.NoRecordedValue:
				sample.State = "unknown"
				sample.Reason = "no recorded value"
			case to-at > age:
				sample.State = "stale"
				sample.Reason = "older than declared freshness"
				v := p.Value
				sample.Value = &v
			default:
				v := p.Value
				sample.Value = &v
			}
			s.Samples = append(s.Samples, sample)
			s.State = sample.State
			i = end
		}
		out.Series = append(out.Series, s)
	}
	return out, nil
}
